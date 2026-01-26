package service

import (
	"context"
	"encoding/json"
	"fmt"
	"iat/common/protocol"
	"sync"

	"github.com/gorilla/websocket"
)

type Client struct {
	ID   string
	Conn *websocket.Conn
	Send chan protocol.Message
}

type WSHub struct {
	clients    map[string]*Client
	register   chan *Client
	unregister chan *Client
	broadcast  chan protocol.Message
	mu         sync.RWMutex
}

func NewWSHub() *WSHub {
	return &WSHub{
		clients:    make(map[string]*Client),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan protocol.Message),
	}
}

func (h *WSHub) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client.ID] = client
			h.mu.Unlock()
			fmt.Printf("WS Client registered: %s\n", client.ID)
		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client.ID]; ok {
				delete(h.clients, client.ID)
				close(client.Send)
				fmt.Printf("WS Client unregistered: %s\n", client.ID)
			}
			h.mu.Unlock()
		case message := <-h.broadcast:
			h.mu.RLock()
			// If To is specified, send only to that client
			if message.To != "" {
				if client, ok := h.clients[message.To]; ok {
					select {
					case client.Send <- message:
					default:
						// Handle blocked client
					}
				}
			} else {
				// Broadcast to all
				for _, client := range h.clients {
					select {
					case client.Send <- message:
					default:
					}
				}
			}
			h.mu.RUnlock()
		}
	}
}

func (h *WSHub) Broadcast(msg protocol.Message) {
	h.broadcast <- msg
}

func (h *WSHub) Register(c *Client) {
	h.register <- c
}

func (h *WSHub) Unregister(c *Client) {
	h.unregister <- c
}

func (c *Client) WritePump() {
	defer func() {
		c.Conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.Send:
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			data, _ := json.Marshal(message)
			if err := c.Conn.WriteMessage(websocket.TextMessage, data); err != nil {
				return
			}
		}
	}
}

func (c *Client) ReadPump(hub *WSHub) {
	defer func() {
		hub.Unregister(c)
		c.Conn.Close()
	}()
	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			break
		}
		var msg protocol.Message
		if err := json.Unmarshal(message, &msg); err == nil {
			msg.From = c.ID
			hub.Broadcast(msg)
		}
	}
}
