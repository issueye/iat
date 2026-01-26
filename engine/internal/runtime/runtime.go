package runtime

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"

	"iat/common/model"
	"iat/common/protocol"
	"iat/engine/internal/service"
)

type Tool func(ctx context.Context, args map[string]any) (any, error)

type AgentHandler interface {
	Handle(ctx context.Context, r *Runtime, inst *AgentInstance, msg protocol.Message) (protocol.Message, error)
}

type AgentInstance struct {
	ID         string
	ModelAgent model.Agent
	Context    context.Context
	Cancel     context.CancelFunc
	Tools      map[string]Tool
	Memory     []string
	Handler    AgentHandler
	Inbox      chan protocol.Message
}

type Runtime struct {
	agents          map[string]*AgentInstance
	mu              sync.RWMutex
	bus             *EventBus
	chatService     *service.ChatService
	registryService *service.RegistryService
	toolService     *service.ToolService
	globalTools     map[string]Tool
	pending         map[string]chan protocol.Message
	pendingMu       sync.Mutex
	clock           func() int64
	timestampSerial atomic.Int64
}

func NewRuntime(chatService *service.ChatService, registryService *service.RegistryService, toolService *service.ToolService) *Runtime {
	return &Runtime{
		agents:          make(map[string]*AgentInstance),
		bus:             NewEventBus(),
		chatService:     chatService,
		registryService: registryService,
		toolService:     toolService,
		globalTools:     make(map[string]Tool),
		pending:         make(map[string]chan protocol.Message),
		clock:           func() int64 { return time.Now().UnixMilli() },
	}
}

func (r *Runtime) RegisterAgent(agent model.Agent) *AgentInstance {
	var handler AgentHandler
	if r.chatService != nil {
		handler = &ChatServiceAgentHandler{chatService: r.chatService}
	}
	return r.RegisterDetachedAgent(agent, handler)
}

func (r *Runtime) RegisterDetachedAgent(agent model.Agent, handler AgentHandler) *AgentInstance {
	r.mu.Lock()
	defer r.mu.Unlock()

	ctx, cancel := context.WithCancel(context.Background())
	instance := &AgentInstance{
		ID:         fmt.Sprintf("agent_%d", agent.ID),
		ModelAgent: agent,
		Context:    ctx,
		Cancel:     cancel,
		Tools:      make(map[string]Tool),
		Memory:     nil,
		Handler:    handler,
		Inbox:      make(chan protocol.Message, 100),
	}
	r.agents[instance.ID] = instance
	go r.runAgentLoop(instance)
	return instance
}

func (r *Runtime) runAgentLoop(instance *AgentInstance) {
	for {
		select {
		case <-instance.Context.Done():
			return
		case msg := <-instance.Inbox:
			if msg.Type == protocol.MsgRequest {
				reply, err := r.handleRequest(instance, msg)
				if err != nil {
					reply = protocol.Message{
						ID:        msg.ID,
						From:      instance.ID,
						To:        msg.From,
						Type:      protocol.MsgResponse,
						Action:    "task_result",
						Timestamp: r.nextTimestamp(),
						Payload:   map[string]any{"error": err.Error()},
					}
				} else {
					reply.ID = msg.ID
					reply.From = instance.ID
					reply.To = msg.From
					reply.Type = protocol.MsgResponse
					if reply.Action == "" {
						reply.Action = "task_result"
					}
					if reply.Timestamp == 0 {
						reply.Timestamp = r.nextTimestamp()
					}
				}
				_ = r.SendMessage(reply)
				continue
			}

			if msg.Type == protocol.MsgResponse {
				r.deliverPending(msg)
			}
		}
	}
}

func (r *Runtime) handleRequest(instance *AgentInstance, msg protocol.Message) (protocol.Message, error) {
	if instance.Handler != nil {
		return instance.Handler.Handle(instance.Context, r, instance, msg)
	}
	if r.chatService != nil {
		return (&ChatServiceAgentHandler{chatService: r.chatService}).Handle(instance.Context, r, instance, msg)
	}
	return protocol.Message{}, fmt.Errorf("no agent handler available")
}

func (r *Runtime) SendMessage(msg protocol.Message) error {
	if msg.Type == protocol.MsgResponse && msg.ID != "" {
		if r.deliverPending(msg) {
			return nil
		}
	}

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

func (r *Runtime) RegisterGlobalTool(name string, tool Tool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.globalTools[name] = tool
}

func (r *Runtime) Call(ctx context.Context, from, to, action string, payload any) (protocol.Message, error) {
	id := uuid.NewString()
	ch := make(chan protocol.Message, 1)

	r.pendingMu.Lock()
	r.pending[id] = ch
	r.pendingMu.Unlock()

	req := protocol.Message{
		ID:        id,
		From:      from,
		To:        to,
		Type:      protocol.MsgRequest,
		Action:    action,
		Payload:   payload,
		Timestamp: r.nextTimestamp(),
	}
	if err := r.SendMessage(req); err != nil {
		r.pendingMu.Lock()
		delete(r.pending, id)
		r.pendingMu.Unlock()
		return protocol.Message{}, err
	}

	select {
	case resp := <-ch:
		return resp, nil
	case <-ctx.Done():
		r.pendingMu.Lock()
		delete(r.pending, id)
		r.pendingMu.Unlock()
		return protocol.Message{}, ctx.Err()
	}
}

func (r *Runtime) deliverPending(msg protocol.Message) bool {
	if msg.ID == "" {
		return false
	}

	r.pendingMu.Lock()
	ch, ok := r.pending[msg.ID]
	if ok {
		delete(r.pending, msg.ID)
	}
	r.pendingMu.Unlock()

	if !ok {
		return false
	}

	select {
	case ch <- msg:
	default:
	}
	return true
}

func (r *Runtime) nextTimestamp() int64 {
	now := r.clock()
	serial := r.timestampSerial.Add(1)
	return now*1_000_000 + (serial % 1_000_000)
}

