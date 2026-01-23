package handler

import (
	"encoding/json"
	"fmt"
	"iat/common/pkg/chat"
	"iat/engine/internal/service"
	"net/http"
	"strconv"
	"time"
)

type ChatHandler struct {
	svc *service.ChatService
}

func NewChatHandler(svc *service.ChatService) *ChatHandler {
	return &ChatHandler{svc: svc}
}

func (h *ChatHandler) Stream(w http.ResponseWriter, r *http.Request) {
	// SSE Headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	var req struct {
		SessionID uint   `json:"sessionId"`
		Message   string `json:"message"`
		AgentID   uint   `json:"agentId"`
		Mode      string `json:"mode"`
	}

	// For SSE, usually we use GET with query params or POST. 
	// If using POST, the body is read once.
	// However, EventSource standard doesn't support POST. 
	// We often use a library like fetch-event-source or just query params.
	// But let's support POST for data carrying. 
	// NOTE: Standard JS EventSource does NOT support POST.
	// We will assume the client uses a fetch-based polyfill or we accept GET query params.
	// Let's implement GET for simplicity first, or accept a setup request.
	
	// Implementation: POST to /api/chat/start -> returns session ID (or just ack)
	// GET /api/chat/events?sessionId=... -> stream events
	
	// BUT, to keep it simple and similar to typical "chat" API:
	// We can use POST and write events to response body.
	
	if r.Method == "POST" {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
	} else {
		// Fallback to query params
		sid, _ := strconv.Atoi(r.URL.Query().Get("sessionId"))
		req.SessionID = uint(sid)
		req.Message = r.URL.Query().Get("message")
		aid, _ := strconv.Atoi(r.URL.Query().Get("agentId"))
		req.AgentID = uint(aid)
		req.Mode = r.URL.Query().Get("mode")
	}

	eventChan := make(chan chat.ChatEvent)
	
	// Start Chat in background
	go func() {
		defer close(eventChan)
		// Pass r.Context() to ensure cancellation if client disconnects
			if err := h.svc.Chat(r.Context(), req.SessionID, req.Message, req.AgentID, req.Mode, eventChan); err != nil {
			eventChan <- chat.ChatEvent{Type: chat.ChatEventError, Content: err.Error()}
		}
	}()

	// Stream events with heartbeat
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case event, ok := <-eventChan:
			if !ok {
				return
			}
			data, _ := json.Marshal(event)
			fmt.Fprintf(w, "data: %s\n\n", data)
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
		case <-ticker.C:
			// Send heartbeat comment
			fmt.Fprintf(w, ": keep-alive\n\n")
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
		case <-r.Context().Done():
			// Client disconnected
			return
		}
	}
}
