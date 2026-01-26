package protocol

type MessageType string

const (
	MsgRequest      MessageType = "request"
	MsgResponse     MessageType = "response"
	MsgNotification MessageType = "notification"
	MsgError        MessageType = "error"
	MsgStreamChunk  MessageType = "stream_chunk"
)

type TraceContext struct {
	TraceID  string `json:"traceId"`
	SpanID   string `json:"spanId"`
	ParentID string `json:"parentId,omitempty"`
}

type Capability struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  map[string]any `json:"parameters"` // JSON Schema for inputs
	Returns     map[string]any `json:"returns"`    // JSON Schema for outputs
}

type RetryPolicy struct {
	MaxAttempts int   `json:"maxAttempts"`
	Interval    int64 `json:"interval"` // in milliseconds
}

type Metadata struct {
	TTL         int64        `json:"ttl,omitempty"` // in milliseconds
	RetryPolicy *RetryPolicy `json:"retryPolicy,omitempty"`
	Source      string       `json:"source,omitempty"`
}

type Message struct {
	ID        string       `json:"id"`
	StreamID  string       `json:"streamId,omitempty"` // For A2A streaming
	From      string       `json:"from"`
	To        string       `json:"to"`
	Type      MessageType  `json:"type"`
	Action    string       `json:"action"`
	Payload   any          `json:"payload"`
	Trace     TraceContext `json:"trace"`
	Metadata  Metadata     `json:"metadata"`
	Timestamp int64        `json:"timestamp"`
}
