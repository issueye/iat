## 1. 后端 (Common & Engine) 详细修改计划

### 第一阶段：协议与数据模型升级 (核心基石)
- **协议升级 (common/protocol)**:
    - 在 [message.go](file:///e:/code/issueye/suwei/iat/common/protocol/message.go) 中增加 `StreamID` 以支持 A2A 流式响应。
    - 增加 `Metadata` 字段，支持消息的 `TTL` (生存时间) 和 `RetryPolicy` (重试策略)。
    - 定义 `common/protocol/schema.go`，预置常见任务（如搜索、编码、命令执行）的标准化输入输出 Contract。
- **模型扩展 (common/model)**:
    - 修改 [agent.go](file:///e:/code/issueye/suwei/iat/common/model/agent.go)，增加 `ConfigSchema` (Agent 配置定义) 和 `MemoryPolicy` (记忆策略)。
    - 新增 `workflow.go`，用于持久化存储基于 DAG 的任务树结构及其执行快照。

### 第二阶段：编排引擎重构 (大脑重塑)
- **解耦运行时 (engine/internal/runtime)**:
    - 将 `Runtime` 职责精简为底层的“Agent 容器管理”与“消息总线”。
    - 创建 `engine/internal/orchestrator` 包，作为高级编排层。
- **实现 DAG 执行器 (Orchestrator)**:
    - 开发 `ExecutionEngine`，支持任务节点的并发与串行调度。
    - 迁移并增强 [planner.go](file:///e:/code/issueye/suwei/iat/engine/internal/runtime/planner.go)，支持更复杂的意图识别。
- **工具服务中心化**:
    - 建立 `ToolService`，统一管理内置工具、JS 脚本工具和 MCP Server 发现的远程工具。

### 第三阶段：通信层全双工化
- **WebSocket 枢纽**:
    - 新增 `engine/api/handler/ws_handler.go`，实现基于 WebSocket 的全双工通信，取代目前碎片化的 SSE 接口。
    - 实现事件多路复用，通过一个连接同时传输对话流、任务状态变更和 Agent 心跳。

---

## 2. 前端 (UI) 详细修改计划

### 第一阶段：工程化与状态管理 (底座升级)
- **Pinia 集成**:
    - 在 `ui/frontend/src/stores` 下建立 `agent.js` (在线状态/能力)、`workflow.js` (任务追踪) 和 `session.js` (对话上下文)。
    - 移除组件内散乱的 `ref` 数据请求，改为由 Store 统一驱动。
- **路由重构**:
    - 增加 `Workspace` 概念，支持同时打开多个项目或多个对话任务。

### 第二阶段：可视化与交互升级 (体验飞跃)
- **工作流画布 (Workflow Canvas)**:
    - 引入 `vue-flow` 库，在 `ui/frontend/src/views/Workflow.vue` 中实现 Agent 协作关系的实时动态展示。
- **动态结果渲染器**:
    - 开发 `ResultRenderer.vue` 组件，根据 Tool 返回的类型自动切换渲染方式（代码 Diff 视图、Markdown 表格、多层级 JSON 树）。
- **增强型对话框**:
    - 重构 `Chat.vue`，支持“多 Agent 并发思考”的 UI 展示，并增加对 Planner 计划的手动确认/修改入口。

---

## 3. 实施路径 (Roadmap)

1.  **[基础期] (第1-3天)**: 完成 `common` 层的模型与协议重构，执行数据库迁移。
2.  **[引擎期] (第4-7天)**: 完成 `Orchestrator` 与 `DAG 执行器` 的核心逻辑，上线 WebSocket 通信。
3.  **[前端期] (第8-11天)**: 完成 Pinia 状态迁移，开发可视化画布与动态结果渲染器。
4.  **[集成期] (第12-14天)**: 全链路联调，优化 Agent 间的协作响应延迟，完善开发者调试面板。

**请确认该详细修改计划。一旦通过，我将立即从“Phase 1: 协议与模型升级”开始编码实施。**