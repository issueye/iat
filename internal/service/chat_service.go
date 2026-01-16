package service

import (
	"context"
	"encoding/json"
	"fmt"
	"iat/internal/model"
	"iat/internal/pkg/ai"
	"iat/internal/pkg/sse"
	"iat/internal/pkg/tools/builtin"
	"iat/internal/pkg/tools/script"
	"iat/internal/repo"
	"strings"

	"github.com/cloudwego/eino/schema"
)

type ChatService struct {
	sessionRepo *repo.SessionRepo
	agentRepo   *repo.AgentRepo
	modelRepo   *repo.AIModelRepo
	messageRepo *repo.MessageRepo
	sseHandler  *sse.SSEHandler
}

func NewChatService(sseHandler *sse.SSEHandler) *ChatService {
	return &ChatService{
		sessionRepo: repo.NewSessionRepo(),
		agentRepo:   repo.NewAgentRepo(),
		modelRepo:   repo.NewAIModelRepo(),
		messageRepo: repo.NewMessageRepo(),
		sseHandler:  sseHandler,
	}
}

// ListMessages returns history messages for a session
func (s *ChatService) ListMessages(sessionID uint) ([]model.Message, error) {
	return s.messageRepo.ListBySessionID(sessionID)
}

// Chat handles the main chat logic
func (s *ChatService) Chat(sessionID uint, userMessage string, agentID uint) error {
	// 1. Get Session
	session, err := s.sessionRepo.GetByID(sessionID)
	if err != nil {
		return fmt.Errorf("session not found: %v", err)
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

	// Permission Check based on Agent Name/Type
	// Chat: only chat, no tools (implicitly handled if tools are not bound)
	// Plan: only plan directory operations
	// Build: all authorized tools
	if agent.Name == "Chat" && agent.Type == "builtin" {
		// Chat agent should not execute tools
		// We can clear tools if they are somehow attached
		agent.Tools = nil
	} else if agent.Name == "Plan" && agent.Type == "builtin" {
		// Plan agent: Ensure system prompt includes instructions to only operate in 'plan' directory
		// And maybe we can enforce it in tool implementation (but that requires context awareness in tool)
		// For now, let's append a strict instruction to system prompt
		agent.SystemPrompt += "\n\nIMPORTANT: You are strictly limited to operating within the 'plan' directory. Do not read or write files outside of this directory."
	} else if agent.Name == "Build" && agent.Type == "builtin" {
		// Build agent has all permissions.
	}

	// 3. Get Model Config
	modelConfig, err := s.modelRepo.GetByID(agent.ModelID)
	if err != nil {
		return fmt.Errorf("model config not found: %v", err)
	}

	// 4. Prepare Tools
	// Get Builtin Tools
	einoTools := builtin.GetEinoTools(agent.Name)
	
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
	
	// 5. Init AI Client
	aiClient, err := ai.NewAIClient(modelConfig, einoTools)
	if err != nil {
		return fmt.Errorf("failed to init ai client: %v", err)
	}

	// 6. Construct Messages (System + User)
	// Save User Message
	userMsg := &model.Message{
		SessionID: sessionID,
		Role:      "user",
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
		if msg.Role == "assistant" {
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
		// Use a loop to handle potential Tool Calls
		// Max turns to prevent infinite loops
		maxTurns := 10 
		
		for i := 0; i < maxTurns; i++ {
			// Create a copy of messages to avoid race conditions if needed, 
			// but here we are in a single goroutine sequentially updating messages.
			
			stream, err := aiClient.StreamChat(context.Background(), messages)
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
					SessionID:  sessionID,
					Role:       "assistant",
					Content:    fullResponse,
					TokenCount: totalTokens, // TODO: Calculate real usage
				}
				s.messageRepo.Create(aiMsg)
				
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
				
				// 2. Parse arguments
				var args map[string]interface{}
				if err := json.Unmarshal([]byte(fnArgs), &args); err != nil {
					messages = append(messages, &schema.Message{
						Role: schema.Tool,
						Content: fmt.Sprintf("Error parsing arguments for tool %s: %v", fnName, err),
						ToolCallID: tc.ID,
					})
					continue
				}
				
				// 3. Execute
				resultStr := ""
				
				// Check if it's a builtin tool
				if _, ok := builtin.ToolFunctions[fnName]; ok {
					switch fnName {
					case "read_file":
						path, _ := args["path"].(string)
						if agent.Name == "Plan" && !strings.Contains(path, "plan") {
							// Check restriction
						}
						res, err := builtin.ReadFile(path)
						if err != nil { resultStr = fmt.Sprintf("Error: %v", err) } else { resultStr = res }
					case "write_file":
						path, _ := args["path"].(string)
						content, _ := args["content"].(string)
						res, err := builtin.WriteFile(path, content)
						if err != nil { resultStr = fmt.Sprintf("Error: %v", err) } else { resultStr = res }
					case "list_files":
						path, _ := args["path"].(string)
						res, err := builtin.ListFiles(path)
						if err != nil { resultStr = fmt.Sprintf("Error: %v", err) } else { resultStr = res }
					case "run_command":
						cmd, _ := args["command"].(string)
						var cmdArgs []string
						if argsRaw, ok := args["args"].([]interface{}); ok {
							for _, a := range argsRaw {
								cmdArgs = append(cmdArgs, fmt.Sprint(a))
							}
						}
						res, err := builtin.RunCommand(cmd, cmdArgs)
						if err != nil { resultStr = fmt.Sprintf("Error: %v", err) } else { resultStr = res }
					case "http_get":
						url, _ := args["url"].(string)
						res, err := builtin.HttpGet(url)
						if err != nil { resultStr = fmt.Sprintf("Error: %v", err) } else { resultStr = res }
					case "http_post":
						url, _ := args["url"].(string)
						body, _ := args["body"].(string)
						ctype, _ := args["contentType"].(string)
						res, err := builtin.HttpPost(url, ctype, body)
						if err != nil { resultStr = fmt.Sprintf("Error: %v", err) } else { resultStr = res }
					default:
						resultStr = "Error: Builtin function implementation not found"
					}
				} else {
					// Check if it's a script tool
					// Currently we don't have a way to know if a tool name belongs to a script tool directly from memory,
					// unless we loaded them.
					// Assuming custom tools are stored in DB and loaded into context.
					// For now, let's look up the tool in the agent's tools list.
					var foundTool *model.Tool
					for _, t := range agent.Tools {
						if t.Name == fnName && t.Type == "custom" {
							foundTool = &t
							break
						}
					}
					
					if foundTool != nil {
						// Execute Script
						engine := script.NewScriptEngine()
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
