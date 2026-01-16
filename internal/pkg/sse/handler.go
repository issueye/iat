package sse

import (
	"fmt"
	"net/http"
)

type SSEHandler struct {
	clients map[chan string]bool
	New     chan chan string
	Closed  chan chan string
	Total   chan string
}

func NewSSEHandler() *SSEHandler {
	handler := &SSEHandler{
		clients: make(map[chan string]bool),
		New:     make(chan chan string),
		Closed:  make(chan chan string),
		Total:   make(chan string),
	}

	go handler.listen()

	return handler
}

func (h *SSEHandler) listen() {
	for {
		select {
		case s := <-h.New:
			h.clients[s] = true
		case s := <-h.Closed:
			delete(h.clients, s)
			close(s)
		case event := <-h.Total:
			for client := range h.clients {
				client <- event
			}
		}
	}
}

func (h *SSEHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	client := make(chan string)
	h.New <- client

	defer func() {
		h.Closed <- client
	}()

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	ctx := r.Context()

	for {
		select {
		case msg := <-client:
			fmt.Fprintf(w, "data: %s\n\n", msg)
			flusher.Flush()
		case <-ctx.Done():
			return
		}
	}
}

func (h *SSEHandler) Send(msg string) {
	h.Total <- msg
}
