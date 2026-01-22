package api

import (
	"net/http"
	"iat/engine/api/handler"
	"iat/engine/internal/service"
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
	// Note: DB init should be done before Server.Start in main.go
	projectSvc := service.NewProjectService()
	mcpSvc := service.NewMCPService()
	taskSvc := service.NewTaskService(nil) // TODO: Handle SSE for tasks
	chatSvc := service.NewChatService(mcpSvc, taskSvc)
	
	// Initialize Handlers
	projectHandler := handler.NewProjectHandler(projectSvc)
	chatHandler := handler.NewChatHandler(chatSvc)

	// CORS middleware
	handler := corsMiddleware(mux)

	mux.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})
	
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
