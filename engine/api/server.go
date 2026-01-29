package api

import (
	"context"
	"encoding/json"
	"iat/common/protocol"
	"iat/engine/api/handler"
	"iat/engine/internal/orchestrator"
	"iat/engine/internal/repo"
	"iat/engine/internal/runtime"
	"iat/engine/internal/service"
	"iat/engine/pkg/ai"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

type Server struct {
	addr string
}

func NewServer(addr string) *Server {
	return &Server{addr: addr}
}

func (s *Server) Start() error {
	mux := http.NewServeMux()

	// Initialize Services
	projectSvc := service.NewProjectService()
	indexSvc := service.NewIndexService()
	mcpSvc := service.NewMCPService()
	wsHub := service.NewWSHub()
	go wsHub.Run(context.Background())

	toolSvc := service.NewToolService(mcpSvc)
	taskSvc := service.NewTaskService(nil)                 // TODO: Handle SSE for tasks
	subAgentTaskSvc := service.NewSubAgentTaskService(nil) // TODO: Handle SSE for sub-agent tasks
	hookSvc := service.NewHookService()
	chatSvc := service.NewChatService(mcpSvc, toolSvc, taskSvc, subAgentTaskSvc, hookSvc, wsHub)
	sessionSvc := service.NewSessionService()
	modelSvc := service.NewAIModelService()
	agentSvc := service.NewAgentService()
	modeSvc := service.NewModeService()
	registrySvc := service.NewRegistryService()
	workflowRepo := repo.NewWorkflowRepo()

	// Initialize Runtime
	rt := runtime.NewRuntime(chatSvc, registrySvc, toolSvc)

	// Initialize Orchestrator components
	// Note: Planner and Reviewer will be created dynamically in ChatService for now 
	// because they need session-specific AI clients.
	router := orchestrator.NewRouter(registrySvc)
	executor := orchestrator.NewExecutionEngine(rt, router, nil, workflowRepo) // Reviewer can be nil or set later
	executor.SetStatusCallback(func(taskId string, status string, output any) {
		wsHub.Broadcast(protocol.Message{
			Type:   protocol.MsgNotification,
			Action: "task_status",
			Payload: map[string]any{
				"taskId": taskId,
				"status": status,
				"output": output,
			},
			Timestamp: time.Now().UnixMilli(),
		})
	})
	chatSvc.SetExecutor(executor)
	chatSvc.SetPlannerFactory(func(client *ai.AIClient) service.TaskPlanner {
		return orchestrator.NewPlanner(client)
	})

	// Initialize Handlers
	projectHandler := handler.NewProjectHandler(projectSvc, indexSvc)
	chatHandler := handler.NewChatHandler(chatSvc)
	sessionHandler := handler.NewSessionHandler(sessionSvc, chatSvc)
	modelHandler := handler.NewAIModelHandler(modelSvc)
	agentHandler := handler.NewAgentHandler(agentSvc)
	toolHandler := handler.NewToolHandler(toolSvc)
	mcpHandler := handler.NewMCPHandler(mcpSvc)
	modeHandler := handler.NewModeHandler(modeSvc)
	hookHandler := handler.NewHookHandler(hookSvc)
	taskHandler := handler.NewTaskHandler(taskSvc)
	subAgentTaskHandler := handler.NewSubAgentTaskHandler(subAgentTaskSvc)
	runtimeTestHandler := handler.NewRuntimeTestHandler()
	registryHandler := handler.NewRegistryHandler(registrySvc)

	// Start registry cleanup goroutine
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		for range ticker.C {
			registrySvc.Cleanup()
		}
	}()

	// CORS middleware
	handler := corsMiddleware(mux)

	// WebSocket Endpoint
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
	mux.HandleFunc("/api/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		clientID := r.URL.Query().Get("id")
		if clientID == "" {
			clientID = "browser_" + time.Now().Format("150405")
		}
		client := &service.Client{
			ID:   clientID,
			Conn: conn,
			Send: make(chan protocol.Message, 256),
		}
		wsHub.Register(client)
		go client.WritePump()
		go client.ReadPump(wsHub)
	})

	// Tasks
	mux.HandleFunc("/api/tasks", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			taskHandler.List(w, r)
		case http.MethodPost:
			taskHandler.Create(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/tasks/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPut:
			taskHandler.Update(w, r)
		case http.MethodDelete:
			taskHandler.Delete(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Sub-Agent Tasks
	mux.HandleFunc("/api/subagent-tasks", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			subAgentTaskHandler.List(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/subagent-tasks/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if strings.HasSuffix(path, "/abort") {
			if r.Method == http.MethodPost {
				subAgentTaskHandler.Abort(w, r)
			} else {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
			return
		}
		if r.Method == http.MethodGet {
			subAgentTaskHandler.Get(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	mux.HandleFunc("/api/runtime/test", runtimeTestHandler.Run)

	// Projects
	mux.HandleFunc("/api/projects", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			projectHandler.List(w, r)
		case http.MethodPost:
			projectHandler.Create(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/projects/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if strings.HasSuffix(path, "/index") {
			if r.Method == http.MethodPost {
				projectHandler.Index(w, r)
			} else {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
			return
		}
		if strings.HasSuffix(path, "/index-all") {
			if r.Method == http.MethodPost {
				projectHandler.IndexAll(w, r)
			} else {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
			return
		}
		switch r.Method {
		case http.MethodPut:
			projectHandler.Update(w, r)
		case http.MethodDelete:
			projectHandler.Delete(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// AI Models
	mux.HandleFunc("/api/models", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			modelHandler.List(w, r)
		case http.MethodPost:
			modelHandler.Create(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/models/test", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			modelHandler.Test(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/models/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			modelHandler.Delete(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Agents
	mux.HandleFunc("/api/agents", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			agentHandler.List(w, r)
		case http.MethodPost:
			agentHandler.Create(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/agents/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPut:
			agentHandler.Update(w, r)
		case http.MethodDelete:
			agentHandler.Delete(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Tools
	mux.HandleFunc("/api/tools", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			toolHandler.List(w, r)
		case http.MethodPost:
			toolHandler.Create(w, r)
		case http.MethodPut:
			toolHandler.Update(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/tools/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			toolHandler.Delete(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// MCP Servers
	mux.HandleFunc("/api/mcp", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			mcpHandler.List(w, r)
		case http.MethodPost:
			mcpHandler.Create(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/mcp/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if strings.HasSuffix(path, "/tools") {
			// /api/mcp/{id}/tools
			if r.Method == http.MethodGet {
				mcpHandler.ListTools(w, r)
			} else {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
			return
		}

		switch r.Method {
		case http.MethodPut:
			mcpHandler.Update(w, r)
		case http.MethodDelete:
			mcpHandler.Delete(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Modes
	mux.HandleFunc("/api/modes", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			modeHandler.List(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Hooks
	mux.HandleFunc("/api/hooks", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			hookHandler.List(w, r)
		case http.MethodPost:
			hookHandler.Create(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/hooks/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPut:
			hookHandler.Update(w, r)
		case http.MethodDelete:
			hookHandler.Delete(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Sessions
	mux.HandleFunc("/api/sessions", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			sessionHandler.List(w, r)
		case http.MethodPost:
			sessionHandler.Create(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/sessions/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if strings.HasSuffix(path, "/messages") {
			// /api/sessions/{id}/messages
			if r.Method == http.MethodGet || r.Method == http.MethodDelete {
				sessionHandler.ListMessages(w, r)
			} else {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
			return
		}

		if strings.HasSuffix(path, "/abort") {
			// /api/sessions/{id}/abort
			if r.Method == http.MethodPost {
				sessionHandler.Abort(w, r)
			} else {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
			return
		}

		switch r.Method {
		case http.MethodPut:
			sessionHandler.Update(w, r)
		case http.MethodDelete:
			sessionHandler.Delete(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Messages
	mux.HandleFunc("/api/messages/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			sessionHandler.DeleteMessage(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Chat Stream
	mux.HandleFunc("/api/chat/stream", chatHandler.Stream)

	// Registry
	mux.HandleFunc("/api/registry/register", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			registryHandler.Register(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/registry/heartbeat/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			registryHandler.Heartbeat(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/registry/discover", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			registryHandler.Discover(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/registry/tools", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			tools, err := mcpSvc.ListAllTools(r.Context())
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			json.NewEncoder(w).Encode(tools)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	return http.ListenAndServe(s.addr, handler)
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		if r.Method == "OPTIONS" {
			return
		}

		next.ServeHTTP(w, r)
	})
}
