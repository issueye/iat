# Agent Platform Design Document

## 1. Overview
The Agent Platform is designed to be a highly extensible, standardized environment for multi-agent collaboration. It supports both local and remote agents, providing them with a unified communication protocol, resource discovery, and task orchestration.

## 2. Communication Protocol (A2A v2)

### 2.1 Message Structure
The platform uses an enhanced message format to support tracing, streaming, and capability negotiation.

```go
type Message struct {
    ID        string         `json:"id"`
    From      string         `json:"from"`      // Agent ID or User ID
    To        string         `json:"to"`        // Target Agent ID
    Type      MessageType    `json:"type"`      // request, response, notification, error
    Action    string         `json:"action"`    // execute, register, status_update, etc.
    Payload   any            `json:"payload"`
    Trace     TraceContext   `json:"trace"`     // For multi-agent task tracing
    Timestamp int64          `json:"timestamp"`
}

type TraceContext struct {
    TraceID  string `json:"traceId"`
    SpanID   string `json:"spanId"`
    ParentID string `json:"parentId,omitempty"`
}
```

### 2.2 Capability Schema
Agents declare their capabilities during registration.

```go
type Capability struct {
    Name        string         `json:"name"`
    Description string         `json:"description"`
    Parameters  map[string]any `json:"parameters"` // JSON Schema for inputs
    Returns     map[string]any `json:"returns"`    // JSON Schema for outputs
}
```

## 3. Agent Registry
The Registry is the central hub for agent discovery.

- **Dynamic Registration**: Agents can register via HTTP/gRPC.
- **Heartbeat**: Agents must send heartbeats to maintain an "online" status.
- **Metadata**: Stores capabilities, endpoint URL, version, and health status.

## 4. Orchestration Flow
1. **Planner**: Receives a user goal, breaks it into sub-tasks (TaskTree).
2. **Router**: For each sub-task, queries the Registry for capable agents.
3. **Executor**: Dispatches messages to selected agents, manages retries and timeouts.
4. **Reviewer**: Evaluates output against the sub-task goal.

## 5. Security & Isolation
- **Sandbox**: Script-based tools run in an isolated environment.
- **Token-based Auth**: External agents must provide a valid token to register and communicate.
