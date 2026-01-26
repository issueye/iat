package chat

type ChatEventType string

const (
	ChatEventChunk      ChatEventType = "chunk"
	ChatEventToolCall   ChatEventType = "tool_call"
	ChatEventToolResult ChatEventType = "tool_result"
	ChatEventError      ChatEventType = "error"
	ChatEventDone       ChatEventType = "done"
	ChatEventUsage      ChatEventType = "usage"
	ChatEventTerminated ChatEventType = "terminated"
	ChatEventSubAgentStart ChatEventType = "subagent_start"
	ChatEventSubAgentChunk ChatEventType = "subagent_chunk"
)

type ChatEvent struct {
	Type    ChatEventType          `json:"type"`
	Content string                 `json:"content"`
	Extra   map[string]interface{} `json:"extra,omitempty"`
}

