package runtime

import (
	"iat/common/protocol"
	"sync"
)

type EventBus struct {
	subscribers map[string][]chan protocol.Message
	mu          sync.RWMutex
}

func NewEventBus() *EventBus {
	return &EventBus{
		subscribers: make(map[string][]chan protocol.Message),
	}
}

func (b *EventBus) Subscribe(topic string) chan protocol.Message {
	b.mu.Lock()
	defer b.mu.Unlock()

	ch := make(chan protocol.Message, 10)
	b.subscribers[topic] = append(b.subscribers[topic], ch)
	return ch
}

func (b *EventBus) Publish(topic string, msg protocol.Message) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	for _, ch := range b.subscribers[topic] {
		select {
		case ch <- msg:
		default:
			// Drop message if channel is full
		}
	}
}
