# golang_agent

一个测试用的分离式 Agent 进程，用于验证 A2A 消息处理、独立记忆与工具授权。

## 启动

```bash
go run ./cmd/golang_agent
```

环境变量：

- `PORT`：监听端口，默认 `18080`
- `AGENT_ID`：Agent ID（字符串），默认 `golang_agent`
- `AGENT_NAME`：Agent 名称，默认 `golang_agent`
- `ALLOWED_TOOLS`：允许工具列表（逗号分隔），默认空
- `AI_API_KEY`：OpenAI 兼容 API Key（必填，用于 eino）
- `AI_BASE_URL`：OpenAI 兼容 BaseURL，可为空使用官方
- `AI_MODEL_NAME`：模型名称，默认 `gpt-4.1-mini`

## HTTP

- `GET /health`：返回 `ok`
- `POST /a2a`：接收并返回 `iat/common/protocol.Message` JSON
- `POST/GET /a2a/stream`：基于 SSE 的聊天接口，body 或 query 传入：

```jsonc
{
  "content": "你好，帮我解释一下这个仓库的结构？",
  "system": "可选，系统提示词"
}
```

返回 SSE 流，`data` 字段为：

```jsonc
{ "type": "chunk", "text": "逐 token/段落返回的内容..." }
```

