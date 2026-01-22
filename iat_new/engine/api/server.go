package api

import (
	"iat/engine/api/handler"
	"iat/engine/internal/service"
	"net/http"
	"strings"
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
	taskSvc := service.NewTaskService(nil) // TODO: Handle SSE for tasks
	chatSvc := service.NewChatService(mcpSvc, taskSvc)
	sessionSvc := service.NewSessionService()
	modelSvc := service.NewAIModelService()
	agentSvc := service.NewAgentService()
	toolSvc := service.NewToolService()
	modeSvc := service.NewModeService()
	
	// Initialize Handlers
	projectHandler := handler.NewProjectHandler(projectSvc, indexSvc)
	chatHandler := handler.NewChatHandler(chatSvc)
	sessionHandler := handler.NewSessionHandler(sessionSvc, chatSvc)
	modelHandler := handler.NewAIModelHandler(modelSvc)
	agentHandler := handler.NewAgentHandler(agentSvc)
	toolHandler := handler.NewToolHandler(toolSvc)
	mcpHandler := handler.NewMCPHandler(mcpSvc)
	modeHandler := handler.NewModeHandler(modeSvc)

	// CORS middleware
	handler := corsMiddleware(mux)

	mux.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})
	
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

	// Chat Stream
	mux.HandleFunc("/api/chat/stream", chatHandler.Stream)
	
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
