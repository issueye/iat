package orchestrator

import (
	"context"
	"encoding/json"
	"fmt"
	"iat/common/model"
	"iat/common/protocol"
	"iat/engine/internal/repo"
	"iat/engine/internal/runtime"
	"sync"
	"time"
)

type ExecutionEngine struct {
	rt           *runtime.Runtime
	router       *Router
	reviewer     *Reviewer
	workflowRepo *repo.WorkflowRepo
	onStatusUp   func(taskId string, status string, output any)
}

func (e *ExecutionEngine) SetStatusCallback(cb func(taskId string, status string, output any)) {
	e.onStatusUp = cb
}

func NewExecutionEngine(rt *runtime.Runtime, router *Router, reviewer *Reviewer, workflowRepo *repo.WorkflowRepo) *ExecutionEngine {
	return &ExecutionEngine{
		rt:           rt,
		router:       router,
		reviewer:     reviewer,
		workflowRepo: workflowRepo,
	}
}

type TaskResult struct {
	TaskID string
	Output any
	Error  error
}

func (e *ExecutionEngine) ExecuteWorkflow(ctx context.Context, workflow *model.Workflow, tasks []model.WorkflowTask) error {
	now := time.Now()
	workflow.Status = model.WorkflowRunning
	workflow.StartedAt = &now

	// 0. Save Workflow to DB
	if e.workflowRepo != nil {
		if err := e.workflowRepo.Create(workflow); err != nil {
			fmt.Printf("Failed to create workflow in DB: %v\n", err)
		}
		for i := range tasks {
			tasks[i].WorkflowID = workflow.ID
			if err := e.workflowRepo.CreateTask(&tasks[i]); err != nil {
				fmt.Printf("Failed to create task %s in DB: %v\n", tasks[i].TaskID, err)
			}
		}
	}

	// Map to track task status and results
	results := make(map[string]*TaskResult)
	resultsMu := sync.Mutex{}

	// Map to track dependencies
	taskMap := make(map[string]*model.WorkflowTask)
	for i := range tasks {
		taskMap[tasks[i].TaskID] = &tasks[i]
	}

	// ... (rest of the logic)
	// Simple DAG execution using dependency counting
	pendingDeps := make(map[string]int)
	dependents := make(map[string][]string)

	for _, t := range tasks {
		var deps []string
		if t.DependsOn != "" {
			json.Unmarshal([]byte(t.DependsOn), &deps)
		}
		pendingDeps[t.TaskID] = len(deps)
		for _, dep := range deps {
			dependents[dep] = append(dependents[dep], t.TaskID)
		}
	}

	// Channel for tasks ready to execute
	readyTasks := make(chan string, len(tasks))
	for id, count := range pendingDeps {
		if count == 0 {
			readyTasks <- id
		}
	}

	wg := sync.WaitGroup{}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var workflowErr error
	once := sync.Once{}

	// Worker pool or dynamic goroutines
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case taskID := <-readyTasks:
				wg.Add(1)
				go func(id string) {
					defer wg.Done()
					task := taskMap[id]

					// 1. Route
					agent, err := e.router.Route(task.Capability)
					if err != nil {
						once.Do(func() { workflowErr = err; cancel() })
						return
					}
					if agent == nil {
						once.Do(func() { workflowErr = fmt.Errorf("no agent found for %s", task.Capability); cancel() })
						return
					}

					// 2. Execute with Retry
					var lastResp protocol.Message
					var execErr error

					task.Status = model.TaskStatusInProgress
					startTime := time.Now()
					task.StartedAt = &startTime
					if e.workflowRepo != nil {
						e.workflowRepo.UpdateTask(task)
					}
					if e.onStatusUp != nil {
						e.onStatusUp(id, "running", nil)
					}

					for attempt := 0; attempt < 3; attempt++ {
						lastResp, execErr = e.rt.Call(ctx, "orchestrator", fmt.Sprintf("agent_%d", agent.ID), "execute_task", task)
						if execErr == nil {
							// 3. Review
							if e.reviewer != nil {
								review, revErr := e.reviewer.Review(ctx, model.SubTask{
									ID:          task.TaskID,
									Title:       task.Title,
									Description: task.Description,
									Capability:  task.Capability,
								}, lastResp.Payload)

								if revErr == nil && review.Approved {
									break
								}
								if revErr != nil {
									execErr = revErr
								} else {
									execErr = fmt.Errorf("review failed: %s", review.Feedback)
									task.Description += "\nFeedback: " + review.Feedback
								}
							} else {
								break // No reviewer, accept result
							}
						}
						time.Sleep(time.Duration(attempt+1) * time.Second)
					}

					resultsMu.Lock()
					results[id] = &TaskResult{TaskID: id, Output: lastResp.Payload, Error: execErr}

					status := model.TaskStatusCompleted
					if execErr != nil {
						status = model.TaskStatusFailed
						task.Error = execErr.Error()
						once.Do(func() { workflowErr = execErr; cancel() })
					} else {
						outJSON, _ := json.Marshal(lastResp.Payload)
						task.Output = string(outJSON)
					}

					task.Status = status
					endTime := time.Now()
					task.CompletedAt = &endTime
					if e.workflowRepo != nil {
						e.workflowRepo.UpdateTask(task)
					}

					if e.onStatusUp != nil {
						e.onStatusUp(id, string(status), lastResp.Payload)
					}

					if execErr != nil {
						resultsMu.Unlock()
						return
					}

					// 4. Trigger dependents
					for _, depID := range dependents[id] {
						pendingDeps[depID]--
						if pendingDeps[depID] == 0 {
							readyTasks <- depID
						}
					}
					resultsMu.Unlock()
				}(taskID)
			}
		}
	}()

	// Wait for all tasks or error
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Check if all tasks finished
		if len(results) < len(tasks) && workflowErr == nil {
			workflowErr = fmt.Errorf("workflow execution incomplete")
		}
	case <-ctx.Done():
		if workflowErr == nil {
			workflowErr = ctx.Err()
		}
	}

	finishedAt := time.Now()
	workflow.CompletedAt = &finishedAt
	if workflowErr != nil {
		workflow.Status = model.WorkflowFailed
		workflow.Result = workflowErr.Error()
	} else {
		workflow.Status = model.WorkflowCompleted
		// Combine results (simple logic for now)
		resJSON, _ := json.Marshal(results)
		workflow.Result = string(resJSON)
	}

	if e.workflowRepo != nil {
		e.workflowRepo.Update(workflow)
	}

	return workflowErr
}
