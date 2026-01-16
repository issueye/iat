package service

import (
	"context"
	"encoding/json"
	"fmt"
	"iat/internal/model"
	"iat/internal/pkg/ai"
	"iat/internal/pkg/consts"
	"iat/internal/pkg/sse"
	"iat/internal/pkg/tools/builtin"
	"iat/internal/pkg/tools/script"
	"iat/internal/repo"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/cloudwego/eino/schema"
	"github.com/eino-contrib/jsonschema"
)

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
	sseHandler  *sse.SSEHandler
	mu          sync.Mutex
	genCounter  uint64
	cancelBySID map[uint]sessionCancel
}

func NewChatService(sseHandler *sse.SSEHandler) *ChatService {
	return &ChatService{
		projectRepo: repo.NewProjectRepo(),
		sessionRepo: repo.NewSessionRepo(),
		agentRepo:   repo.NewAgentRepo(),
		modelRepo:   repo.NewAIModelRepo(),
		messageRepo: repo.NewMessageRepo(),
		toolRepo:    repo.NewToolInvocationRepo(),
		sseHandler:  sseHandler,
		cancelBySID: make(map[uint]sessionCancel),
	}
}

func (s *ChatService) sendToolEvent(sessionID uint, payload map[string]interface{}) {
	msg, err := json.Marshal(map[string]interface{}{
		"sessionId": sessionID,
		"tool":      payload,
	})
	if err != nil {
		return
	}
	s.sseHandler.Send(string(msg))
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
func (s *ChatService) Chat(sessionID uint, userMessage string, agentID uint, modeKey string) error {
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
		if msg.Role == consts.RoleAssistant {
			role = schema.Assistant
		}
		messages = append(messages, &schema.Message{
			Role:    role,
			Content: msg.Content,
		})
	}

	// 6. Stream Chat
	// Run in goroutine to not block
	go func() {
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
				termMsg, _ := json.Marshal(map[string]interface{}{
					"sessionId":  sessionID,
					"terminated": true,
					"done":       true,
					"error":      "terminated",
				})
				s.sseHandler.Send(string(termMsg))
				return
			}
			// Create a copy of messages to avoid race conditions if needed,
			// but here we are in a single goroutine sequentially updating messages.

			// 376: stream, err := aiClient.StreamChat(ctx, messages)
			// Capture the prompt sent to AI
			promptJSON, _ := json.Marshal(messages)
			currentPrompt := string(promptJSON)

			stream, err := aiClient.StreamChat(ctx, messages)
			if err != nil {
				errMsg, _ := json.Marshal(map[string]interface{}{
					"sessionId": sessionID,
					"error":     err.Error(),
				})
				s.sseHandler.Send(string(errMsg))
				return
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
					msg, _ := json.Marshal(map[string]interface{}{
						"sessionId": sessionID,
						"delta":     chunk.Content,
					})
					s.sseHandler.Send(string(msg))
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
				termMsg, _ := json.Marshal(map[string]interface{}{
					"sessionId":  sessionID,
					"terminated": true,
					"done":       true,
					"error":      "terminated",
				})
				s.sseHandler.Send(string(termMsg))
				return
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
				usageMsg, _ := json.Marshal(map[string]interface{}{
					"sessionId": sessionID,
					"usage":     aiMsg.TokenCount,
				})
				s.sseHandler.Send(string(usageMsg))

				// Append to conversation context for next turn
				messages = append(messages, &schema.Message{
					Role:      schema.Assistant,
					Content:   fullResponse,
					ToolCalls: toolCalls,
				})
			}

			// If no tool calls, we are done
			if len(toolCalls) == 0 {
				// Send done signal
				doneMsg, _ := json.Marshal(map[string]interface{}{
					"sessionId": sessionID,
					"done":      true,
					"usage":     totalTokens,
				})
				s.sseHandler.Send(string(doneMsg))
				return
			}

			// Execute Tools
			for _, tc := range toolCalls {
				// 1. Find function
				fnName := tc.Function.Name
				fnArgs := tc.Function.Arguments
				s.sendToolEvent(sessionID, map[string]interface{}{
					"stage":      "call",
					"name":       fnName,
					"arguments":  fnArgs,
					"toolCallId": tc.ID,
				})
				_ = s.toolRepo.UpsertCall(sessionID, tc.ID, fnName, fnArgs)

				// 2. Parse arguments
				var args map[string]interface{}
				if err := json.Unmarshal([]byte(fnArgs), &args); err != nil {
					s.sendToolEvent(sessionID, map[string]interface{}{
						"stage":      "result",
						"name":       fnName,
						"toolCallId": tc.ID,
						"ok":         false,
						"output":     fmt.Sprintf("Error parsing arguments for tool %s: %v", fnName, err),
					})
					messages = append(messages, &schema.Message{
						Role:       schema.Tool,
						Content:    fmt.Sprintf("Error parsing arguments for tool %s: %v", fnName, err),
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
						resultStr = fmt.Sprintf("Error: Tool function '%s' not found", fnName)
					}
				}

				ok := true
				if strings.HasPrefix(resultStr, "Error:") {
					ok = false
				}
				_ = s.toolRepo.UpsertResult(sessionID, tc.ID, fnName, resultStr, ok)
				s.sendToolEvent(sessionID, map[string]interface{}{
					"stage":      "result",
					"name":       fnName,
					"toolCallId": tc.ID,
					"ok":         ok,
					"output":     resultStr,
				})

				// 4. Append Tool Message
				messages = append(messages, &schema.Message{
					Role:       schema.Tool,
					Content:    resultStr,
					ToolCallID: tc.ID,
				})
			}

			// Loop continues to send Tool Results back to LLM
		}
	}()

	return nil
}
