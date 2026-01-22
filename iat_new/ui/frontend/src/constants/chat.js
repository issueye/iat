export const ChatRoles = {
  System: "system",
  User: "user",
  Assistant: "assistant",
  Tool: "tool",
};

export const ChatModes = {
  Chat: "chat",
  Plan: "plan",
  Build: "build",
};

export const ToolStages = {
  Call: "call",
  Result: "result",
};

export const ThinkingStatuses = {
  Start: "start",
  Thinking: "thinking",
  End: "end",
  Error: "error",
  Cancel: "cancel",
};

export const ThinkTags = {
  Open: "<think>",
  Close: "</think>",
};

export const SSE = {
  EventsUrl: "http://localhost:8080/api/chat/stream",
};
