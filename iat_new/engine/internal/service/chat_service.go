package service

import (
	"context"
	"encoding/json"
	"fmt"
	"iat/common/model"
	"iat/engine/pkg/ai"
	"iat/common/pkg/consts"
	"iat/engine/pkg/tools/builtin"
	"iat/engine/pkg/tools/script"
	"iat/engine/internal/repo"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/cloudwego/eino/schema"
	"github.com/eino-contrib/jsonschema"
)

func stripThinkContent(input string) string {
	s := input
	for {
		start := strings.Index(s, "<think>")
		if start == -1 {
			break
		}
		rest := s[start+len("<think>"):]
		endRel := strings.Index(rest, "</think>")
		if endRel == -1 {
			s = s[:start]
			break
		}
		end := start + len("<think>") + endRel + len("</think>")
		s = s[:start] + s[end:]
	}
	s = strings.ReplaceAll(s, "</think>", "")
	return strings.TrimSpace(s)
}

type sessionCancel struct {
	id     uint64
	cancel context.CancelFunc
}

type ChatService struct {
	projectRepo *repo.ProjectRepo
	sessionRepo *repo.SessionRepo
	agentRepo   *repo.AgentRepo
	modelRepo   *repo.AIModelRepo
	messageRepo *repo.MessageRepo
	toolRepo    *repo.ToolInvocationRepo
	mcpService  *MCPService
	taskService *TaskService
	mu          sync.Mutex
	genCounter  uint64
	cancelBySID map[uint]sessionCancel
}

func NewChatService(mcpService *MCPService, taskService *TaskService) *ChatService {
	return &ChatService{
		projectRepo: repo.NewProjectRepo(),
		sessionRepo: repo.NewSessionRepo(),
		agentRepo:   repo.NewAgentRepo(),
		modelRepo:   repo.NewAIModelRepo(),
		messageRepo: repo.NewMessageRepo(),
		toolRepo:    repo.NewToolInvocationRepo(),
		mcpService:  mcpService,
		taskService: taskService,
		cancelBySID: make(map[uint]sessionCancel),
	}
}

func (s *ChatService) sendToolEvent(sessionID uint, payload map[string]interface{}, eventChan chan<- ChatEvent) {
	// Construct generic event
	eventChan <- ChatEvent{
		Type:  ChatEventToolCall, // Generic tool event type
		Extra: payload,
	}
}

// ListMessages returns history messages for a session
func (s *ChatService) ListMessages(sessionID uint) ([]model.Message, error) {
	return s.messageRepo.ListBySessionID(sessionID)
}

func (s *ChatService) ClearMessages(sessionID uint) error {
	s.AbortSession(sessionID)
	if err := s.messageRepo.DeleteBySessionID(sessionID); err != nil {
		return err
	}
	_ = s.toolRepo.DeleteBySessionID(sessionID)
	return nil
}

// RunAgentInternal runs an agent synchronously for internal calls (like sub-agents)
func (s *ChatService) RunAgentInternal(sessionID uint, agentName, query, projectRoot string, modeKey string, eventChan chan<- ChatEvent) (string, error) {
	// 1. Find Agent
	agents, err := s.agentRepo.List() // TODO: Optimize by Name lookup
	if err != nil {
		return "", fmt.Errorf("failed to list agents: %v", err)
	}
	var targetAgent *model.Agent
	for _, a := range agents {
		if a.Name == agentName {
			targetAgent = &a
			break
		}
	}
	if targetAgent == nil {
		return "", fmt.Errorf("agent '%s' not found", agentName)
	}

	// Permission Check
	effectiveMode := targetAgent.Mode.Key
	if modeKey != "" {
		effectiveMode = targetAgent.Mode.Key
	}

	// 2. Get Model Config
	var modelConfig *model.AIModel
	if targetAgent.ModelID != 0 {
		modelConfig, err = s.modelRepo.GetByID(targetAgent.ModelID)
	} else {
		modelConfig, err = s.modelRepo.GetDefault()
	}
	if err != nil || modelConfig == nil {
		return "", fmt.Errorf("model config not found for agent")
	}

	// 3. Prepare Tools
	einoTools := builtin.GetEinoTools(effectiveMode)
	mcpTools, err := s.mcpService.GetToolsForServers(targetAgent.MCPServers)
	if err == nil {
		einoTools = append(einoTools, mcpTools...)
	}

	for _, t := range targetAgent.Tools {
		if t.Type == consts.ToolTypeCustom || t.Type == consts.ToolTypeScript {
			var s jsonschema.Schema
			if err := json.Unmarshal([]byte(t.Parameters), &s); err == nil {
				einoTools = append(einoTools, &schema.ToolInfo{
					Name:        t.Name,
					Desc:        t.Description,
					ParamsOneOf: schema.NewParamsOneOfByJSONSchema(&s),
				})
			}
		}
	}

	// 4. Init AI Client
	aiClient, err := ai.NewAIClient(modelConfig, einoTools)
	if err != nil {
		return "", fmt.Errorf("failed to init ai client: %v", err)
	}

	// 5. Messages
	messages := []*schema.Message{
		{Role: schema.System, Content: targetAgent.SystemPrompt},
		{Role: schema.User, Content: query},
	}

	// 6. Loop
	ctx := context.Background()
	maxTurns := 30 

	for i := 0; i < maxTurns; i++ {
		resp, err := aiClient.Chat(ctx, messages)
		if err != nil {
			return "", err
		}

		// Emit intermediate thought/content as event
		if resp.Content != "" {
			if eventChan != nil {
				s.sendToolEvent(sessionID, map[string]interface{}{
					"stage":   "subagent_chunk",
					"content": resp.Content,
				}, eventChan)
			}
		}

		// Check for Tool Calls
		if len(resp.ToolCalls) == 0 {
			// Done
			return stripThinkContent(resp.Content), nil
		}

		// Append Assistant Message with Tool Calls
		messages = append(messages, resp)

		// Execute Tools
		for _, tc := range resp.ToolCalls {
			fnName := tc.Function.Name
			fnArgs := tc.Function.Arguments

			var args map[string]interface{}
			if err := json.Unmarshal([]byte(fnArgs), &args); err != nil {
				output := fmt.Sprintf("Error parsing arguments: %v", err)
				messages = append(messages, &schema.Message{
					Role: schema.Tool, Content: output, ToolCallID: tc.ID,
				})
				continue
			}

			resultStr := ""
			handled := false

			// Builtins
			switch fnName {
			case "read_file":
				path, _ := args["path"].(string)
				p, _ := builtin.ResolvePathInBase(projectRoot, path)
				resultStr, _ = builtin.ReadFile(p)
				handled = true
			case "read_file_range":
				path, _ := args["path"].(string)
				startLine, _ := args["startLine"].(float64)
				limit, _ := args["limit"].(float64)
				p, _ := builtin.ResolvePathInBase(projectRoot, path)
				resultStr, _ = builtin.ReadFileRange(p, int(startLine), int(limit))
				handled = true
			case "diff_file":
				path1, _ := args["path1"].(string)
				path2, _ := args["path2"].(string)
				p1, _ := builtin.ResolvePathInBase(projectRoot, path1)
				p2, _ := builtin.ResolvePathInBase(projectRoot, path2)
				resultStr, _ = builtin.DiffFile(p1, p2)
				handled = true
			case "write_file":
				path, _ := args["path"].(string)
				content, _ := args["content"].(string)
				p, _ := builtin.ResolvePathInBase(projectRoot, path)
				resultStr, _ = builtin.WriteFile(p, content)
				handled = true
			case "list_files":
				path, _ := args["path"].(string)
				p, _ := builtin.ResolvePathInBase(projectRoot, path)
				resultStr, _ = builtin.ListFiles(p)
				handled = true
			case "run_command":
				cmd, _ := args["command"].(string)
				var cmdArgs []string
				if argsRaw, ok := args["args"].([]interface{}); ok {
					for _, a := range argsRaw {
						cmdArgs = append(cmdArgs, fmt.Sprint(a))
					}
				}
				resultStr, _ = builtin.RunCommand(cmd, cmdArgs)
				handled = true
			case "run_script":
				scriptPath, _ := args["scriptPath"].(string)
				sp, _ := builtin.ResolvePathInBase(projectRoot, scriptPath)
				var scriptArgs []string
				if argsRaw, ok := args["args"].([]interface{}); ok {
					for _, a := range argsRaw {
						scriptArgs = append(scriptArgs, fmt.Sprint(a))
					}
				}
				resultStr, _ = builtin.RunScript(sp, scriptArgs)
				handled = true
			case "call_subagent":
				// Recursive call
				an, _ := args["agentName"].(string)
				q, _ := args["query"].(string)
				var subErr error
				resultStr, subErr = s.RunAgentInternal(sessionID, an, q, projectRoot, effectiveMode, eventChan)
				if subErr != nil {
					resultStr = fmt.Sprintf("Error calling sub-agent: %v", subErr)
				}
				handled = true
			case "manage_tasks":
				action, _ := args["action"].(string)
				content, _ := args["content"].(string)
				idVal, _ := args["id"].(float64)
				id := uint(idVal)
				status, _ := args["status"].(string)
				priority, _ := args["priority"].(string)
				pidVal, _ := args["parentId"].(float64)
				var parentID *uint
				if pidVal > 0 {
					pid := uint(pidVal)
					parentID = &pid
				}

				switch action {
				case "add":
					if priority == "" {
						priority = "medium"
					}
					t, err := s.taskService.CreateTask(sessionID, content, priority, parentID)
					if err != nil {
						resultStr = fmt.Sprintf("Error: %v", err)
					} else {
						resultStr = fmt.Sprintf("Task created: ID=%d", t.ID)
					}
				case "update":
					if id == 0 {
						resultStr = "Error: ID is required for update"
						break
					}
					if status != "" {
						err := s.taskService.UpdateTask(id, status)
						if err != nil {
							resultStr = fmt.Sprintf("Error: %v", err)
						} else {
							resultStr = "Task updated"
						}
					} else {
						resultStr = "Error: Status is required for update"
					}
				case "delete":
					if id == 0 {
						resultStr = "Error: ID is required for delete"
						break
					}
					err := s.taskService.DeleteTask(id)
					if err != nil {
						resultStr = fmt.Sprintf("Error: %v", err)
					} else {
						resultStr = "Task deleted"
					}
				case "list":
					tasks, err := s.taskService.ListTasks(sessionID)
					if err != nil {
						resultStr = fmt.Sprintf("Error: %v", err)
					} else {
						b, _ := json.Marshal(tasks)
						resultStr = string(b)
					}
				default:
					resultStr = fmt.Sprintf("Unknown action: %s", action)
				}
				handled = true
			}

			if !handled {
				// Script Tool Lookup
				var foundTool *model.Tool
				for _, t := range targetAgent.Tools {
					if t.Name == fnName && (t.Type == consts.ToolTypeCustom || t.Type == consts.ToolTypeScript) {
						foundTool = &t
						break
					}
				}
				if foundTool != nil {
					engine := script.NewScriptEngineWithBaseDir(projectRoot)
					engine.RegisterTool("args", args)
					res, err := engine.Run(foundTool.Content)
					if err != nil {
						resultStr = fmt.Sprintf("Error: %v", err)
					} else {
						resultStr = fmt.Sprintf("%v", res)
					}
					handled = true
				}
			}

			if !handled && s.mcpService != nil {
				res, err := s.mcpService.CallTool(ctx, fnName, args)
				if err == nil {
					resultStr = res
					handled = true
				} else {
					resultStr = fmt.Sprintf("Error: %v", err)
				}
			}

			if !handled && resultStr == "" {
				resultStr = fmt.Sprintf("Error: Tool '%s' not found", fnName)
			}

			messages = append(messages, &schema.Message{
				Role: schema.Tool, Content: resultStr, ToolCallID: tc.ID,
			})
		}
	}

	return "", fmt.Errorf("max turns exceeded")
}

func (s *ChatService) ListToolInvocations(sessionID uint) ([]model.ToolInvocation, error) {
	return s.toolRepo.ListBySessionID(sessionID)
}

func (s *ChatService) CompressSession(sessionID uint) error {
	s.AbortSession(sessionID)

	session, err := s.sessionRepo.GetByID(sessionID)
	if err != nil {
		return fmt.Errorf("session not found: %v", err)
	}
	history, err := s.messageRepo.ListBySessionID(sessionID)
	if err != nil {
		return fmt.Errorf("failed to load history: %v", err)
	}

	maxMsgs := 60
	start := 0
	if len(history) > maxMsgs {
		start = len(history) - maxMsgs
	}
	conv := ""
	for _, msg := range history[start:] {
		conv += fmt.Sprintf("[%s]\n%s\n\n", msg.Role, msg.Content)
	}

	summary := strings.TrimSpace(conv)
	if summary == "" {
		summary = "（空会话）"
	} else {
		if len(summary) > 4000 {
			summary = summary[len(summary)-4000:]
		}
		summary = "（简易压缩）\n\n" + summary
	}

	if session.AgentID != 0 {
		agent, aerr := s.agentRepo.GetByID(session.AgentID)
		if aerr == nil {
			var modelConfig *model.AIModel
			if agent.ModelID != 0 {
				modelConfig, err = s.modelRepo.GetByID(agent.ModelID)
			} else {
				modelConfig, err = s.modelRepo.GetDefault()
			}
			if err == nil && modelConfig != nil {
				aiClient, cerr := ai.NewAIClient(modelConfig, nil)
				if cerr == nil {
					prompt := []*schema.Message{
						{
							Role:    schema.System,
							Content: "你是一个会话压缩器。请将给定对话压缩为可用于后续继续对话的摘要，要求：1) 保留用户目标/约束/关键决定；2) 列出重要实体/文件/命令；3) 用中文要点输出；4) 不要编造不存在的信息。",
						},
						{
							Role:    schema.User,
							Content: conv,
						},
					}
					resp, serr := aiClient.Chat(context.Background(), prompt)
					if serr == nil {
						if s2 := strings.TrimSpace(resp.Content); s2 != "" {
							summary = s2
						}
					}
				}
			}
		}
	}

	if err := s.messageRepo.DeleteBySessionID(sessionID); err != nil {
		return err
	}
	_ = s.toolRepo.DeleteBySessionID(sessionID)
	aiMsg := &model.Message{
		SessionID:  sessionID,
		Role:       consts.RoleAssistant,
		Content:    summary,
		TokenCount: len(summary) / 4,
	}
	if err := s.messageRepo.Create(aiMsg); err != nil {
		return err
	}

	session.Compressed = true
	session.Summary = summary
	if err := s.sessionRepo.Update(session); err != nil {
		return err
	}
	return nil
}

func (s *ChatService) AbortSession(sessionID uint) {
	s.mu.Lock()
	entry, ok := s.cancelBySID[sessionID]
	if ok && entry.cancel != nil {
		delete(s.cancelBySID, sessionID)
	}
	s.mu.Unlock()
	if ok && entry.cancel != nil {
		entry.cancel()
	}
}

// Chat handles the main chat logic
func (s *ChatService) Chat(sessionID uint, userMessage string, agentID uint, modeKey string, eventChan chan<- ChatEvent) error {
	// 1. Get Session
	session, err := s.sessionRepo.GetByID(sessionID)
	if err != nil {
		return fmt.Errorf("session not found: %v", err)
	}

	projectRoot := ""
	var project *model.Project
	if session.ProjectID != 0 {
		p, perr := s.projectRepo.GetByID(session.ProjectID)
		if perr == nil {
			project = p
			projectRoot = p.Path
		}
	}

	// 2. Get Agent
	// If agentID is provided (e.g. from toolbar), use it.
	// Otherwise, use the session's default agent.
	targetAgentID := session.AgentID
	if agentID > 0 {
		targetAgentID = agentID
	}

	if targetAgentID == 0 {
		return fmt.Errorf("session has no agent assigned and no agent selected")
	}
	agent, err := s.agentRepo.GetByID(targetAgentID)
	if err != nil {
		return fmt.Errorf("agent not found: %v", err)
	}

	// Permission Check based on Agent Mode
	// If modeKey is provided, it overrides the agent's default mode.
	effectiveMode := agent.Mode.Key
	if modeKey != "" {
		effectiveMode = modeKey
	}

	if effectiveMode == "chat" {
		// Chat agent should not execute tools
		// We can clear tools if they are somehow attached
		agent.Tools = nil
	} else if effectiveMode == "plan" {
		// Plan agent: Ensure system prompt includes instructions to only operate in 'plan' directory
		// And maybe we can enforce it in tool implementation (but that requires context awareness in tool)
		// For now, let's append a strict instruction to system prompt
		agent.SystemPrompt += consts.SystemPromptPlanRestriction
	} else if effectiveMode == "build" {
		// Build agent has all permissions.
	}

	// 3. Get Model Config
	var modelConfig *model.AIModel
	if agent.ModelID != 0 {
		modelConfig, err = s.modelRepo.GetByID(agent.ModelID)
		if err != nil {
			return fmt.Errorf("model config not found: %v", err)
		}
	} else {
		// Use default model
		modelConfig, err = s.modelRepo.GetDefault()
		if err != nil {
			return fmt.Errorf("no default model found and agent has no model assigned")
		}
	}

	// 4. Prepare Tools
	// Get Builtin Tools
	einoTools := builtin.GetEinoTools(effectiveMode)

	// Get MCP Tools bound to Agent
	mcpTools, err := s.mcpService.GetToolsForServers(agent.MCPServers)
	if err == nil {
		einoTools = append(einoTools, mcpTools...)
	}

	// Get Custom Tools (if any) - assuming agent.Tools is populated (need preload in GetByID)
	// Currently GetByID preloads Tools, so we are good.
	// We need to convert agent.Tools to schema.ToolInfo
	// Note: Custom tool conversion logic is similar to builtin.GetEinoTools
	// For now, let's assume we just use builtin tools filtered by agent name as requested.
	// If we want to support custom tools attached to agent, we should iterate agent.Tools.

	// Merge tools
	// TODO: Implement custom tools conversion if needed.
	// Since we are in "optimizing" phase and user emphasized builtin agent roles,
	// we stick to builtin.GetEinoTools logic which handles the filtering.
	// However, if we want to support custom tools (scripts), we should add them here.
	// Let's iterate agent.Tools and convert them to schema.ToolInfo if they are 'custom' type.

	for _, t := range agent.Tools {
		if t.Type == consts.ToolTypeCustom {
			var s jsonschema.Schema
			if err := json.Unmarshal([]byte(t.Parameters), &s); err != nil {
				fmt.Printf("Failed to parse schema for custom tool %s: %v\n", t.Name, err)
				continue
			}
			einoTools = append(einoTools, &schema.ToolInfo{
				Name:        t.Name,
				Desc:        t.Description,
				ParamsOneOf: schema.NewParamsOneOfByJSONSchema(&s),
			})
		}
	}
	// However, if we want to support custom tools (scripts), we should add them here.
	// Let's iterate agent.Tools and convert them to schema.ToolInfo if they are 'custom' type.

	for _, t := range agent.Tools {
		if t.Type == "custom" {
			var s jsonschema.Schema
			if err := json.Unmarshal([]byte(t.Parameters), &s); err != nil {
				fmt.Printf("Failed to parse schema for custom tool %s: %v\n", t.Name, err)
				continue
			}
			einoTools = append(einoTools, &schema.ToolInfo{
				Name:        t.Name,
				Desc:        t.Description,
				ParamsOneOf: schema.NewParamsOneOfByJSONSchema(&s),
			})
		}
	}

	// 5. Init AI Client
	aiClient, err := ai.NewAIClient(modelConfig, einoTools)
	if err != nil {
		return fmt.Errorf("failed to init ai client: %v", err)
	}

	// 6. Construct Messages (System + User)
	// Save User Message
	userMsg := &model.Message{
		SessionID: sessionID,
		Role:      consts.RoleUser,
		Content:   userMessage,
	}
	if err := s.messageRepo.Create(userMsg); err != nil {
		return fmt.Errorf("failed to save user message: %v", err)
	}

	// Load History (including the one just saved, but we need structure for AI client)
	history, err := s.messageRepo.ListBySessionID(sessionID)
	if err != nil {
		return fmt.Errorf("failed to load history: %v", err)
	}

	messages := []*schema.Message{
		{
			Role:    schema.System,
			Content: agent.SystemPrompt,
		},
	}

	for _, msg := range history {
		role := schema.User
		content := msg.Content
		if msg.Role == consts.RoleAssistant {
			role = schema.Assistant
			content = stripThinkContent(content)
		}
		messages = append(messages, &schema.Message{
			Role:    role,
			Content: content,
		})
	}

	// 6. Stream Chat
	// Run synchronously (caller handles concurrency)
	ctx, cancel := context.WithCancel(context.Background())
	genID := atomic.AddUint64(&s.genCounter, 1)
	s.mu.Lock()
	if prev, ok := s.cancelBySID[sessionID]; ok && prev.cancel != nil {
		prev.cancel()
	}
	s.cancelBySID[sessionID] = sessionCancel{id: genID, cancel: cancel}
	s.mu.Unlock()
	defer func() {
		cancel()
		s.mu.Lock()
		if cur, ok := s.cancelBySID[sessionID]; ok && cur.id == genID {
			delete(s.cancelBySID, sessionID)
		}
		s.mu.Unlock()
	}()

	// Use a loop to handle potential Tool Calls
	// Max turns to prevent infinite loops
	maxTurns := 10

	for i := 0; i < maxTurns; i++ {
		if ctx.Err() != nil {
			eventChan <- ChatEvent{Type: ChatEventTerminated}
			return nil
		}
		// Create a copy of messages to avoid race conditions if needed,
		// but here we are in a single goroutine sequentially updating messages.

		// 376: stream, err := aiClient.StreamChat(ctx, messages)
		// Capture the prompt and tools sent to AI
		promptData := map[string]interface{}{
			"messages": messages,
			"tools":    einoTools,
		}
		promptJSON, _ := json.Marshal(promptData)
		currentPrompt := string(promptJSON)

		stream, err := aiClient.StreamChat(ctx, messages)
		if err != nil {
			eventChan <- ChatEvent{Type: ChatEventError, Content: err.Error()}
			return nil
		}

		totalTokens := 0
		fullResponse := ""
		// Map to accumulate tool calls by index
		toolCallsMap := make(map[int]*schema.ToolCall)

		for {
			chunk, err := stream.Recv()
			if err != nil {
				// EOF or Error
				// Check if error is EOF, otherwise log it
				// stream.Recv returns io.EOF when done
				break
			}

			// Handle Content
			if chunk.Content != "" {
				fullResponse += chunk.Content
				eventChan <- ChatEvent{Type: ChatEventChunk, Content: chunk.Content}
			}

			// Handle Tool Calls (Accumulate)
			for _, tc := range chunk.ToolCalls {
				// Index is usually present in streaming tool calls
				idx := 0
				if tc.Index != nil {
					idx = *tc.Index
				}

				if _, exists := toolCallsMap[idx]; !exists {
					toolCallsMap[idx] = &schema.ToolCall{
						Index: tc.Index,
						ID:    tc.ID,
						Type:  tc.Type,
						Function: schema.FunctionCall{
							Name:      "",
							Arguments: "",
						},
					}
				}

				// Merge fields
				if tc.ID != "" {
					toolCallsMap[idx].ID = tc.ID
				}
				if tc.Type != "" {
					toolCallsMap[idx].Type = tc.Type
				}
				if tc.Function.Name != "" {
					toolCallsMap[idx].Function.Name += tc.Function.Name
				}
				if tc.Function.Arguments != "" {
					toolCallsMap[idx].Function.Arguments += tc.Function.Arguments
				}
			}
		}
		stream.Close()
		if ctx.Err() != nil {
			eventChan <- ChatEvent{Type: ChatEventTerminated}
			return nil
		}

		// Convert map back to slice
		var toolCalls []schema.ToolCall
		for idx := 0; idx < len(toolCallsMap); idx++ {
			if tc, ok := toolCallsMap[idx]; ok {
				toolCalls = append(toolCalls, *tc)
			}
		}

		// Save Assistant Message (Content)
		// Only save if there is content or tool calls
		if fullResponse != "" || len(toolCalls) > 0 {
			aiMsg := &model.Message{
				SessionID: sessionID,
				Role:      consts.RoleAssistant,
				Content:   fullResponse,
				Prompt:    currentPrompt,
			}

			// Calculate real usage
			// Simple approximation: 1 token ~= 4 chars
			aiMsg.TokenCount = len(fullResponse) / 4

			s.messageRepo.Create(aiMsg)
			totalTokens += aiMsg.TokenCount
			eventChan <- ChatEvent{
				Type: ChatEventUsage,
				Extra: map[string]interface{}{"usage": aiMsg.TokenCount},
			}

			// Append to conversation context for next turn
			messages = append(messages, &schema.Message{
				Role:      schema.Assistant,
				Content:   stripThinkContent(fullResponse),
				ToolCalls: toolCalls,
			})
		}

		// If no tool calls, we are done
		if len(toolCalls) == 0 {
			// Send done signal
			eventChan <- ChatEvent{
				Type: ChatEventDone,
				Extra: map[string]interface{}{"usage": totalTokens},
			}
			return nil
		}

		// Execute Tools
		for _, tc := range toolCalls {
			// 1. Find function
			fnName := tc.Function.Name
			fnArgs := tc.Function.Arguments
			fmt.Printf("[ChatService] Executing tool: %s\n", fnName) // LOG

			s.sendToolEvent(sessionID, map[string]interface{}{
				"stage":      consts.ToolStageCall,
				"name":       fnName,
				"arguments":  fnArgs,
				"toolCallId": tc.ID,
			}, eventChan)
			// ... (upsert calls)
			if err := s.toolRepo.UpsertCall(sessionID, tc.ID, fnName, fnArgs); err != nil {
				fmt.Printf("[ChatService] Failed to upsert tool call to repo: %v\n", err)
			}
			if err := s.messageRepo.UpsertToolCall(sessionID, tc.ID, fnName, fnArgs); err != nil {
				fmt.Printf("[ChatService] Failed to upsert tool call to message repo: %v\n", err)
			}

			// 2. Parse arguments
			var args map[string]interface{}
			if err := json.Unmarshal([]byte(fnArgs), &args); err != nil {
				output := fmt.Sprintf("Error parsing arguments for tool %s: %v", fnName, err)
				s.sendToolEvent(sessionID, map[string]interface{}{
					"stage":      consts.ToolStageResult,
					"name":       fnName,
					"toolCallId": tc.ID,
					"ok":         false,
					"output":     output,
				}, eventChan)
				_ = s.toolRepo.UpsertResult(sessionID, tc.ID, fnName, output, false)
				_ = s.messageRepo.UpsertToolResult(sessionID, tc.ID, fnName, output, false)
				messages = append(messages, &schema.Message{
					Role:       schema.Tool,
					Content:    output,
					ToolCallID: tc.ID,
				})
				continue
			}

			// 3. Execute
			resultStr := ""

			switch fnName {
			case "read_file":
				path, _ := args["path"].(string)
				p, rerr := builtin.ResolvePathInBase(projectRoot, path)
				if rerr != nil {
					resultStr = fmt.Sprintf("Error: %v", rerr)
					break
				}
				if agent.Mode.Key == "plan" && !strings.Contains(p, "plan") {
				}
				res, err := builtin.ReadFile(p)
				if err != nil {
					resultStr = fmt.Sprintf("Error: %v", err)
				} else {
					resultStr = res
				}
			case "read_file_range":
				path, _ := args["path"].(string)
				startLine, _ := args["startLine"].(float64)
				limit, _ := args["limit"].(float64)

				p, rerr := builtin.ResolvePathInBase(projectRoot, path)
				if rerr != nil {
					resultStr = fmt.Sprintf("Error: %v", rerr)
					break
				}

				res, err := builtin.ReadFileRange(p, int(startLine), int(limit))
				if err != nil {
					resultStr = fmt.Sprintf("Error: %v", err)
				} else {
					resultStr = res
				}
			case "diff_file":
				path1, _ := args["path1"].(string)
				path2, _ := args["path2"].(string)

				p1, rerr1 := builtin.ResolvePathInBase(projectRoot, path1)
				if rerr1 != nil {
					resultStr = fmt.Sprintf("Error: %v", rerr1)
					break
				}

				p2, rerr2 := builtin.ResolvePathInBase(projectRoot, path2)
				if rerr2 != nil {
					resultStr = fmt.Sprintf("Error: %v", rerr2)
					break
				}

				res, err := builtin.DiffFile(p1, p2)
				if err != nil {
					resultStr = fmt.Sprintf("Error: %v", err)
				} else {
					resultStr = res
				}
			case "write_file":
				path, _ := args["path"].(string)
				content, _ := args["content"].(string)
				p, rerr := builtin.ResolvePathInBase(projectRoot, path)
				if rerr != nil {
					resultStr = fmt.Sprintf("Error: %v", rerr)
					break
				}
				res, err := builtin.WriteFile(p, content)
				if err != nil {
					resultStr = fmt.Sprintf("Error: %v", err)
				} else {
					resultStr = res
				}
			case "list_files":
				path, _ := args["path"].(string)
				p, rerr := builtin.ResolvePathInBase(projectRoot, path)
				if rerr != nil {
					resultStr = fmt.Sprintf("Error: %v", rerr)
					break
				}
				res, err := builtin.ListFiles(p)
				if err != nil {
					resultStr = fmt.Sprintf("Error: %v", err)
				} else {
					resultStr = res
				}
			case "run_command":
				cmd, _ := args["command"].(string)
				var cmdArgs []string
				if argsRaw, ok := args["args"].([]interface{}); ok {
					for _, a := range argsRaw {
						cmdArgs = append(cmdArgs, fmt.Sprint(a))
					}
				}
				res, err := builtin.RunCommand(cmd, cmdArgs)
				if err != nil {
					resultStr = fmt.Sprintf("Error: %v", err)
				} else {
					resultStr = res
				}
			case "run_script":
				scriptPath, _ := args["scriptPath"].(string)
				sp, rerr := builtin.ResolvePathInBase(projectRoot, scriptPath)
				if rerr != nil {
					resultStr = fmt.Sprintf("Error: %v", rerr)
					break
				}
				var scriptArgs []string
				if argsRaw, ok := args["args"].([]interface{}); ok {
					for _, a := range argsRaw {
						scriptArgs = append(scriptArgs, fmt.Sprint(a))
					}
				}
				res, err := builtin.RunScript(sp, scriptArgs)
				if err != nil {
					resultStr = fmt.Sprintf("Error: %v", err)
				} else {
					resultStr = res
				}
			case "http_get":
				url, _ := args["url"].(string)
				res, err := builtin.HttpGet(url)
				if err != nil {
					resultStr = fmt.Sprintf("Error: %v", err)
				} else {
					resultStr = res
				}
			case "http_post":
				url, _ := args["url"].(string)
				body, _ := args["body"].(string)
				ctype, _ := args["contentType"].(string)
				res, err := builtin.HttpPost(url, ctype, body)
				if err != nil {
					resultStr = fmt.Sprintf("Error: %v", err)
				} else {
					resultStr = res
				}
			case "index_project":
				all, _ := args["all"].(bool)
				pid := uint(0)
				if pv, ok := args["projectId"].(float64); ok && pv > 0 {
					pid = uint(pv)
				}
				if !all && pid == 0 {
					if project != nil {
						pid = project.ID
					} else {
						pid = session.ProjectID
					}
				}
				indexSvc := NewIndexService()
				if all {
					info, err := indexSvc.IndexAllProjects()
					if err != nil {
						resultStr = fmt.Sprintf("Error: %v", err)
					} else {
						resultStr = fmt.Sprintf("success: indexed=%d files=%d dbPath=%s", info.Indexed, info.Files, info.DBPath)
					}
					break
				}
				if pid == 0 {
					resultStr = "Error: projectId is required"
					break
				}
				info, err := indexSvc.IndexProject(pid)
				if err != nil {
					resultStr = fmt.Sprintf("Error: %v", err)
				} else {
					resultStr = fmt.Sprintf("success: indexed=%d files=%d dbPath=%s", info.Indexed, info.Files, info.DBPath)
				}
			case "call_subagent":
				agentName, _ := args["agentName"].(string)
				query, _ := args["query"].(string)

				// Notify Start of SubAgent
				s.sendToolEvent(sessionID, map[string]interface{}{
					"stage":      "subagent_start",
					"name":       "call_subagent",
					"agentName":  agentName,
					"toolCallId": tc.ID,
				}, eventChan)

				// Run Internal Agent
				res, err := s.RunAgentInternal(sessionID, agentName, query, projectRoot, agent.Mode.Key, eventChan)
				if err != nil {
					resultStr = fmt.Sprintf("Error calling sub-agent '%s': %v", agentName, err)
				} else {
					resultStr = res
				}

				// Notify End of SubAgent (with partial content if needed, but resultStr is final)
				// The resultStr will be sent via standard tool result event

			case "manage_tasks":
				action, _ := args["action"].(string)
				content, _ := args["content"].(string)
				idVal, _ := args["id"].(float64)
				id := uint(idVal)
				status, _ := args["status"].(string)
				priority, _ := args["priority"].(string)
				pidVal, _ := args["parentId"].(float64)
				var parentID *uint
				if pidVal > 0 {
					pid := uint(pidVal)
					parentID = &pid
				}

				switch action {
				case "add":
					if priority == "" {
						priority = "medium"
					}
					t, err := s.taskService.CreateTask(sessionID, content, priority, parentID)
					if err != nil {
						resultStr = fmt.Sprintf("Error: %v", err)
					} else {
						resultStr = fmt.Sprintf("Task created: ID=%d", t.ID)
					}
				case "update":
					if id == 0 {
						resultStr = "Error: ID is required for update"
						break
					}
					if status != "" {
						err := s.taskService.UpdateTask(id, status)
						if err != nil {
							resultStr = fmt.Sprintf("Error: %v", err)
						} else {
							resultStr = "Task updated"
						}
					} else {
						resultStr = "Error: Status is required for update (currently only status update is supported)"
					}
				case "delete":
					if id == 0 {
						resultStr = "Error: ID is required for delete"
						break
					}
					err := s.taskService.DeleteTask(id)
					if err != nil {
						resultStr = fmt.Sprintf("Error: %v", err)
					} else {
						resultStr = "Task deleted"
					}
				case "list":
					tasks, err := s.taskService.ListTasks(sessionID)
					if err != nil {
						resultStr = fmt.Sprintf("Error: %v", err)
					} else {
						b, _ := json.Marshal(tasks)
						resultStr = string(b)
					}
				default:
					resultStr = fmt.Sprintf("Unknown action: %s", action)
				}
			case "review_output":
				passed, _ := args["passed"].(bool)
				comment, _ := args["comment"].(string)
				taskIdVal, _ := args["taskId"].(float64)
				taskId := uint(taskIdVal)

				resultStr = fmt.Sprintf("Review recorded. Passed: %v, Comment: %s", passed, comment)

				if taskId > 0 {
					status := "failed"
					if passed {
						status = "completed"
					}
					if err := s.taskService.UpdateTask(taskId, status); err != nil {
						resultStr += fmt.Sprintf(" (Failed to update task status: %v)", err)
					} else {
						resultStr += fmt.Sprintf(" (Task %d marked as %s)", taskId, status)
					}
				}
			default:
				// Check if it's a script tool
				// Currently we don't have a way to know if a tool name belongs to a script tool directly from memory,
				// unless we loaded them.
				// Assuming custom tools are stored in DB and loaded into context.
				// For now, let's look up the tool in the agent's tools list.
				var foundTool *model.Tool
				for _, t := range agent.Tools {
					// Support both "custom" and "script" types for now, as UI sends "script"
					if t.Name == fnName && (t.Type == consts.ToolTypeCustom || t.Type == consts.ToolTypeScript) {
						foundTool = &t
						break
					}
				}

				if foundTool != nil {
					// Execute Script
					engine := script.NewScriptEngineWithBaseDir(projectRoot)
					// Inject arguments as global variables or call a function wrapper
					// Better approach: wrap the script content in a function call or set variables and run.
					// Simplest: Set 'args' variable and run script.
					// BUT, usually script tool content IS the function body or a full script.
					// Let's assume content is the script code.

					// Inject args
					engine.RegisterTool("args", args)

					// Run script
					res, err := engine.Run(foundTool.Content)
					if err != nil {
						resultStr = fmt.Sprintf("Error executing script tool: %v", err)
					} else {
						resultStr = fmt.Sprintf("%v", res)
					}
				} else {
					handled := false
					if s.mcpService != nil {
						res, err := s.mcpService.CallTool(ctx, fnName, args)
						if err == nil {
							resultStr = res
							handled = true
						}
					}
					if !handled {
						resultStr = fmt.Sprintf("Error: Tool function '%s' not found", fnName)
					}
				}
			}

			// ... (upsert result)

			// LOG RESULT LENGTH
			fmt.Printf("[ChatService] Tool %s result length: %d\n", fnName, len(resultStr))

			ok := true
			if strings.HasPrefix(resultStr, "Error:") {
				ok = false
			}
			_ = s.toolRepo.UpsertResult(sessionID, tc.ID, fnName, resultStr, ok)
			_ = s.messageRepo.UpsertToolResult(sessionID, tc.ID, fnName, resultStr, ok)
			s.sendToolEvent(sessionID, map[string]interface{}{
				"stage":      consts.ToolStageResult,
				"name":       fnName,
				"toolCallId": tc.ID,
				"ok":         ok,
				"output":     resultStr,
			}, eventChan)

			// 4. Append Tool Message
			messages = append(messages, &schema.Message{
				Role:       schema.Tool,
				Content:    resultStr,
				ToolCallID: tc.ID,
			})
		}

		// Loop continues to send Tool Results back to LLM
		fmt.Printf("[ChatService] Tool execution done, continuing loop to LLM...\n")
	}

	return nil
}
