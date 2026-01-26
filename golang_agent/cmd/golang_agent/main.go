package main

import (
	"context"
	"encoding/json"
	"iat/common/protocol"
	"iat/golang_agent/internal/agent"
	"iat/golang_agent/internal/ai"
	"iat/golang_agent/internal/tools"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/cloudwego/eino/schema"
)

func main() {
	port := envInt("PORT", 18080)
	agentID := envString("AGENT_ID", "golang_agent")
	agentName := envString("AGENT_NAME", "golang_agent")
	allowedTools := parseCSV(envString("ALLOWED_TOOLS", ""))

	aiClient, err := ai.NewClientFromEnv()
	if err != nil {
		panic(err)
	}

	a := agent.New(agentID, agentName)
	builtins := tools.Builtins()
	for _, name := range allowedTools {
		if t, ok := builtins[name]; ok {
			a.GrantTool(name, t)
		}
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})
	mux.HandleFunc("/a2a/stream", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		var req struct {
			Content string `json:"content"`
			System  string `json:"system"`
		}
		if r.Method == http.MethodPost {
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, "Invalid JSON", http.StatusBadRequest)
				return
			}
		} else {
			req.Content = r.URL.Query().Get("content")
			req.System = r.URL.Query().Get("system")
		}

		if strings.TrimSpace(req.Content) == "" {
			http.Error(w, "content required", http.StatusBadRequest)
			return
		}

		systemMsg := req.System
		if systemMsg == "" {
			systemMsg = "You are golang_agent, an AI assistant."
		}

		messages := []*schema.Message{
			{Role: schema.System, Content: systemMsg},
			{Role: schema.User, Content: req.Content},
		}

		ctx, cancel := context.WithCancel(r.Context())
		defer cancel()

		stream, err := aiClient.Stream(ctx, messages)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer stream.Close()

		ticker := time.NewTicker(15 * time.Second)
		defer ticker.Stop()

		flusher, _ := w.(http.Flusher)

		for {
			select {
			case <-ticker.C:
				_, _ = w.Write([]byte(": keep-alive\n\n"))
				if flusher != nil {
					flusher.Flush()
				}
			default:
				chunk, err := stream.Recv()
				if err != nil {
					return
				}
				if chunk.Content != "" {
					data, _ := json.Marshal(map[string]any{
						"type": "chunk",
						"text": chunk.Content,
					})
					_, _ = w.Write([]byte("data: " + string(data) + "\n\n"))
					if flusher != nil {
						flusher.Flush()
					}
				}
			}
		}
	})
	mux.HandleFunc("/a2a", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var msg protocol.Message
		if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
		defer cancel()

		resp := a.Handle(ctx, msg)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	})

	srv := &http.Server{
		Addr:              ":" + strconv.Itoa(port),
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
	_ = srv.ListenAndServe()
}

func envString(key, def string) string {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return def
	}
	return v
}

func envInt(key string, def int) int {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}

func parseCSV(s string) []string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
