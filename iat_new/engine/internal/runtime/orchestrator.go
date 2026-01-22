package runtime

import (
	"fmt"
	"iat/common/protocol"
)

// Orchestrator extensions
func (r *Runtime) DispatchTask(fromAgentID, toAgentID, taskContent string) error {
	msg := protocol.Message{
		From:   fromAgentID,
		To:     toAgentID,
		Type:   protocol.MsgRequest,
		Action: "execute_task",
		Payload: map[string]string{
			"content": taskContent,
		},
	}
	return r.SendMessage(msg)
}

func (r *Runtime) ReviewOutput(reviewerID, workerID, feedback string, approved bool) error {
	msg := protocol.Message{
		From:   reviewerID,
		To:     workerID,
		Type:   protocol.MsgResponse,
		Action: "review_result",
		Payload: map[string]interface{}{
			"feedback": feedback,
			"approved": approved,
		},
	}
	return r.SendMessage(msg)
}

func (r *Runtime) GrantTool(adminID, targetAgentID, toolName string) error {
	// Verify admin privileges
	// TODO: Check if adminID is Orchestrator
	
	r.mu.Lock()
	defer r.mu.Unlock()
	
	target, ok := r.agents[targetAgentID]
	if !ok {
		return fmt.Errorf("agent %s not found", targetAgentID)
	}
	
	// Mock tool granting
	target.Tools[toolName] = true 
	
	return nil
}
