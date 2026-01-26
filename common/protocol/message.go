package protocol

type MessageType string

const (
	MsgRequest      MessageType = "request"
	MsgResponse     MessageType = "response"
	MsgNotification MessageType = "notification"
)

type Message struct {
	ID        string      `json:"id"`
	From      string      `json:"from"`
	To        string      `json:"to"`
	Type      MessageType `json:"type"`
	Action    string      `json:"action"`
	Payload   any         `json:"payload"`
	Timestamp int64       `json:"timestamp"`
}
