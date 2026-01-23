# A2A (Agent to Agent) 通信协议标准需求文档

## 1. 目标与范围

### 1.1 目标

建立一套标准化的、可扩展的 Agent 间通信协议，使 IAT 引擎能够与不同语言编写、不同物理部署位置的智能体（Sub-Agents）进行高效协作。

### 1.2 适用范围

- IAT Core Engine (Host) 与 Golang Sub-Agent。
- 未来支持的 Python, Node.js 等其他语言实现的 Sub-Agents。
- 跨网络的远程 Agent 协作。

---

## 2. 协议分层设计 (Protocol Layering)

### 2.1 传输层 (Transport)

- **Primary**: HTTP/2 (支持 SSE 流式返回)。
- **Optional**: WebSocket (用于长连接实时双向通信)。

### 2.2 消息层 (Message Envelope)

所有消息必须遵循以下基础结构：

```json
{
  "header": {
    "msg_id": "uuid",
    "parent_id": "uuid",
    "type": "request|response|notification|chunk",
    "from": "agent_id_1",
    "to": "agent_id_2",
    "timestamp": 1674480000
  },
  "metadata": {
    "trace_id": "global_trace_id",
    "ttl": 3,
    "auth_token": "bearer ..."
  },
  "payload": {}
}
```

---

## 3. 核心功能需求

### 3.1 异步任务与中间态同步

- **[R01] 非阻塞请求**: 支持发送不等待立即回复的请求，使用 `notification` 同步中间状态。
- **[R02] 流式 Chunk 支持**: Sub-Agent 在思考或生成长文本时，必须以 `type: chunk` 形式发送增量内容。

### 3.2 动态能力协商 (Capability Negotiation)

- **[R03] 握手协议**: Agent 启动时应发送 `action: register` 消息，声明其支持的工具列表样式。
- **[R04] 工具描述标准化**: 采用 JSON Schema 描述工具输入输出，确保主引擎能正确序列化参数。

### 3.3 交互式授权 (Human-in-the-Loop)

- **[R05] 授权阻塞请求**: 当 Sub-Agent 欲执行高风险操作（如删除、大规模修改）时，应发送 `action: require_approval` 消息。
- **[R06] 审批反馈**: 主引擎通过 Response 返回 `approved: true|false`。

---

## 4. 重点 Actions 定义

| Action 类型 | 描述                   | Payload 关键字段                              |
| :---------- | :--------------------- | :-------------------------------------------- |
| `execute`   | 发起任务指令           | `query`, `context_fragments`, `allowed_tools` |
| `think`     | 上报思考逻辑 (CoT)     | `thought_text`                                |
| `call_tool` | 请求代为执行主引擎工具 | `tool_name`, `arguments`                      |
| `log`       | 调试日志同步           | `level`, `message`                            |
| `done`      | 任务完成报文           | `result_summary`, `usage_stats`               |

---

## 5. 安全性需求

- **鉴权 (Authentication)**: 每次请求必须携带基于 Session 的签名或 Token。
- **沙箱隔离 (Isolation)**: 协议层应支持 `workspace_root` 声明，限制 Sub-Agent 仅能访问特定目录。

---

**编写者**: Antigravity (PM Mode)
**日期**: 2026-01-23
