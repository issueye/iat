package sse

import (
	"fmt"
	"net/http"
)

type SSEHandler struct {
	clients map[chan string]bool
	new     chan chan string
	closed  chan chan string
	total   chan string
}

func NewSSEHandler() *SSEHandler {
	handler := &SSEHandler{
		clients: make(map[chan string]bool),
		new:     make(chan chan string),
		closed:  make(chan chan string),
		total:   make(chan string),
	}

	go handler.listen()

	return handler
}

func (h *SSEHandler) listen() {
	for {
		select {
		case s := <-h.new:
			h.clients[s] = true
		case s := <-h.closed:
			delete(h.clients, s)
			close(s)
		case event := <-h.total:
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
	h.new <- client

	defer func() {
		h.closed <- client
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
	h.total <- msg
}
