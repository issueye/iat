package service

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
	Type    ChatEventType
	Content string
	Extra   map[string]interface{}
}
