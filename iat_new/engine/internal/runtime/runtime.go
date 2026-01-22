package runtime

import (
	"context"
	"fmt"
	"iat/common/model"
	"iat/common/protocol"
	"iat/engine/internal/service"
	"sync"
)

type AgentInstance struct {
	ID         string
	ModelAgent model.Agent
	Context    context.Context
	Cancel     context.CancelFunc
	Tools      map[string]interface{} // TODO: Define Tool interface
	Inbox      chan protocol.Message
}

type Runtime struct {
	agents      map[string]*AgentInstance
	mu          sync.RWMutex
	bus         *EventBus
	chatService *service.ChatService
}

func NewRuntime(chatService *service.ChatService) *Runtime {
	return &Runtime{
		agents:      make(map[string]*AgentInstance),
		bus:         NewEventBus(),
		chatService: chatService,
	}
}

func (r *Runtime) RegisterAgent(agent model.Agent) *AgentInstance {
	r.mu.Lock()
	defer r.mu.Unlock()

	ctx, cancel := context.WithCancel(context.Background())
	instance := &AgentInstance{
		ID:         fmt.Sprintf("agent_%d", agent.ID),
		ModelAgent: agent,
		Context:    ctx,
		Cancel:     cancel,
		Tools:      make(map[string]interface{}),
		Inbox:      make(chan protocol.Message, 100),
	}
	r.agents[instance.ID] = instance
	
	// Start agent loop
	go r.runAgentLoop(instance)
	
	return instance
}

func (r *Runtime) runAgentLoop(instance *AgentInstance) {
	for {
		select {
		case <-instance.Context.Done():
			return
		case msg := <-instance.Inbox:
			fmt.Printf("Agent %s received message: %v\n", instance.ID, msg)
			
			if msg.Type == protocol.MsgRequest {
				// Use RunAgentInternal for A2A request
				fmt.Printf("Agent %s executing A2A request...\n", instance.ID)
				
				go func() {
					// We need a session ID for logging/history, even for A2A.
					// For now, we pass 0 or a system session ID if we had one.
					// Assuming RunAgentInternal handles sessionID=0 gracefully or we mock it.
					// Looking at ChatService, it sends events to sessionID. 
					// Ideally we should create a temporary session or use a "Task" abstraction.
					// Let's assume we can use 0 for now, but events might be lost if no client listening.
					// But RunAgentInternal returns the final string result, which is what we need.
					
					// We need to resolve project root. Assuming A2A happens in context of a project.
					// We might need to pass project path in payload or agent config.
					// For now, let's use "." or empty string.
					projectRoot := "."
					
					result, err := r.chatService.RunAgentInternal(
						0, // sessionID
						instance.ModelAgent.Name,
						msg.Payload.(map[string]interface{})["content"].(string), // Assuming payload structure
						projectRoot,
						instance.ModelAgent.Mode.Key,
						nil, // No event channel for A2A intermediate events yet
					)
					
					reply := protocol.Message{
						From:      instance.ID,
						To:        msg.From,
						Type:      protocol.MsgResponse,
						Action:    "task_result",
						Timestamp: msg.Timestamp + 1,
					}
					
					if err != nil {
						reply.Payload = map[string]interface{}{"error": err.Error()}
					} else {
						reply.Payload = map[string]interface{}{"result": result}
					}
					
					r.SendMessage(reply)
				}()
			}
		}
	}
}

func (r *Runtime) SendMessage(msg protocol.Message) error {
	r.mu.RLock()
	target, ok := r.agents[msg.To]
	r.mu.RUnlock()

	if !ok {
		return fmt.Errorf("agent %s not found", msg.To)
	}

	select {
	case target.Inbox <- msg:
		return nil
	default:
		return fmt.Errorf("agent %s inbox full", msg.To)
	}
}
