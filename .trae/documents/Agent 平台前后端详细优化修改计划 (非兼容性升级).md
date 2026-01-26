## Agent 平台深度优化详细修改计划 (非兼容性升级)

### 1. 后端 (Common & Engine) 核心重构

#### **Phase 1: 协议与模型标准化 (1-2天)**
- **`common/protocol` 升级**:
    - **[message.go](file:///e:/code/issueye/suwei/iat/common/protocol/message.go)**: 
        - 增加 `StreamID string` (支持 A2A 异步流式数据)。
        - 增加 `Metadata` 结构（包含 `TTL`, `RetryPolicy`, `Priority`）。
    - **[contract.go](file:///e:/code/issueye/suwei/iat/common/protocol/contract.go)** (新): 
        - 定义 `TaskContract` 接口，预置 `Search`, `Code`, `FileSystem` 等标准能力的 JSON Schema 契约。
- **`common/model` 扩展**:
    - **[agent.go](file:///e:/code/issueye/suwei/iat/common/model/agent.go)**: 增加 `ConfigSchema` (Agent 私有配置约束), `MemoryPolicy` (决定短期/长期记忆存储策略)。
    - **[workflow.go](file:///e:/code/issueye/suwei/iat/common/model/workflow.go)** (新): 定义 `Workflow` (DAG 结构) 与 `NodeExecution` (单节点快照)。

#### **Phase 2: 编排引擎 (Orchestrator) 与 DAG 执行器 (3-4天)**
- **`engine/internal/orchestrator` (新包)**:
    - **`engine.go`**: 实现基于 `goroutine` 池的 DAG 调度器，负责解析 `Workflow` 并按依赖顺序并发触发任务。
    - **`planner.go`**: 重构原 Planner，引入“意图树”分析，将复杂 User Goal 拆解为带条件分支的 DAG。
- **`engine/internal/runtime` 精简化**:
    - **[runtime.go](file:///e:/code/issueye/suwei/iat/engine/internal/runtime/runtime.go)**: 移除复杂的业务逻辑，仅保留 Agent 实例的 `Context` 生命周期管理、消息路由（Bus）和 `Call/Stream` 基础方法。
- **`engine/internal/service/tool_service.go`**:
    - 建立中心化工具库，支持 `LocalTool`, `RemoteTool (MCP)`, `ScriptTool` 的统一寻址与权限校验。

#### **Phase 3: 通信总线 WebSocket 化 (1-2天)**
- **`engine/api/handler/ws_handler.go`** (新):
    - 实现全双工通信，通过 `MsgType` 区分 `ChatStream`, `AgentStatus`, `WorkflowProgress`。
    - 实现“事件多路复用”，解决 SSE 在多任务并发时连接数受限的问题。

---

### 2. 前端 (UI) 架构升级

#### **Phase 1: 状态管理 Store 化 (2-3天)**
- **集成 Pinia**:
    - `ui/frontend/src/stores/agent.js`: 实时同步注册中心的所有 Agent 状态及其能力集。
    - `ui/frontend/src/stores/workflow.js`: 追踪当前所有活动工作流的实时进度（节点状态、耗时、输出）。
- **[router/index.js](file:///e:/code/issueye/suwei/iat/ui/frontend/src/router/index.js)**: 引入 `Workspace` 嵌套路由，支持同时在侧边栏切换多个独立的项目空间。

#### **Phase 2: 可视化与交互重塑 (3-4天)**
- **`ui/frontend/src/views/Workflow.vue`** (新):
    - 基于 `vue-flow` 实现可视化画布，支持：
        - 实时展示 Agent 间的“思考脉络”。
        - 点击节点查看该步骤的原始消息、工具调用及 Trace 链路。
- **`ui/frontend/src/components/ResultRenderer.vue`** (新):
    - 统一工具结果展示：
        - `CodeDiff`: 针对文件修改。
        - `DataTable`: 针对结构化搜索结果。
        - `LogViewer`: 针对命令行输出。
- **`ui/frontend/src/views/Chat.vue` 增强**:
    - 侧边弹出式“任务面板”，展示 Planner 生成的任务树，支持用户在执行前微调计划。

---

### 3. 集成测试与开发者工具 (2天)
- **`engine/cmd/main.go` 启动逻辑**: 自动根据 `config.yaml` 初始化的默认 Agent 集群（内置搜索、编码、评审智能体）。
- **调试面板**: 前端增加 `DevTools` 入口，支持模拟发送 A2A 协议报文，方便调试外部 Agent。

**确认计划后，我将首先进行 `common/protocol` 和 `common/model` 的非兼容性重写。**