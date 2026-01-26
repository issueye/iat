## 1. 当前项目深度分析
### 现有能力总结
- **多模块架构**: 基于 `go.work` 的 `common`, `engine`, `ui`, `golang_agent` 结构，解耦良好。
- **运行时核心**: 已具备基本的 Agent 执行逻辑、SSE 事件推送、以及基于脚本引擎（Go-based）的工具扩展。
- **协议雏形**: `common/protocol` 定义了基础的消息结构，但缺乏会话跟踪、流式处理和能力协商。
- **UI 基础**: 已有管理 Agent、Model、Tool、MCP 的界面，具备初步的对话能力。

### 转型 Agent 平台的核心挑战
- **标准化缺失**: 外部 Agent 如何以统一的方式“插入”平台并被其他 Agent 发现？
- **协作逻辑硬编码**: 目前的协作逻辑（如 `DispatchTask`）在 `Runtime` 中较为固定，需要转向基于策略或编排的动态分发。
- **可观测性弱**: 无法直观查看 Agent 间的思考过程、工具调用链路及子任务状态。

## 2. Agent 平台功能点详解
### A. 通信协议层 (A2A Protocol v2)
- **多模式消息**: 支持同步 Request-Response、异步 Notification、以及 Stream 流式响应。
- **会话上下文**: 引入 `TraceID` (全链路追踪) 和 `TaskID` (子任务关联)，支持任务树结构。
- **能力描述符**: 统一的 JSON Schema 描述 Agent 的输入输出参数及其具备的技能。

### B. 核心注册中心 (Agent Registry)
- **动态注册**: 提供 REST/gRPC 接口供本地或远程 Agent 动态注册。
- **元数据管理**: 维护 Agent 的版本、状态（在线/忙碌/离线）、标签（如：编码型、搜索型）。
- **心跳与探针**: 实时监测 Agent 健康状况，自动处理节点下线。

### C. 智能编排引擎 (Orchestrator)
- **任务规划 (Planner)**: 引入 `Plan` 阶段，将复杂 Goal 拆解为有序的 `Sub-task` 列表。
- **动态路由 (Router)**: 根据子任务需求，从 Registry 中匹配最合适的 Agent。
- **执行与监控**: 循环执行子任务，支持 `Self-Reflection`（自省）和 `Human-in-the-loop`（人工干预）。

### D. 扩展性设施 (Tools & MCP)
- **脚本沙箱**: 为 `common/pkg/script` 增加资源限制和权限控制。
- **MCP 管理中心**: 集中化管理 MCP Server，支持 Agent 间共享 MCP 上下文。

### E. 开发者工具 (SDK & CLI)
- **Go SDK**: 封装注册、消息处理、工具调用，使开发一个新 Agent 只需实现核心 Logic。
- **仿真环境**: 模拟平台环境，用于本地调试 Agent 的协作逻辑。

## 3. 详细开发计划
### 第一阶段：协议标准化与基础重构 (第1-2周)
1. **[Protocol]** 升级 `common/protocol`，增加 `TraceContext` 和 `Capability` 结构。
2. **[Registry]** 在 `engine` 中实现 `AgentRegistry` 服务，支持内存/数据库双态存储。
3. **[Model]** 扩展 `Agent` 模型，增加 `Status`, `Endpoint`, `Capabilities` 字段。
4. **[Refactor]** 调整 `engine` API，支持外部 Agent 注册回调。

### 第二阶段：编排引擎与协作逻辑 (第3-4周)
1. **[Planner]** 实现基于 Prompt 的任务拆解器，输出标准化的 `TaskTree`。
2. **[Router]** 实现基于能力的 Agent 选择算法（初步支持标签匹配）。
3. **[Executor]** 升级 `Runtime`，支持多 Agent 并发/顺序执行任务，并记录全链路 Trace。
4. **[Feedback]** 实现 `Reviewer` 逻辑，支持 Agent 间的任务成果打回重做。

### 第三阶段：SDK 开发与外部集成 (第5-6周)
1. **[Go SDK]** 重构 `golang_agent`，提取出通用的 `iat-go-sdk`。
2. **[Tooling]** 实现 MCP 服务的动态发现，允许 Agent 在运行时请求特定 MCP 权限。
3. **[Security]** 增强脚本引擎的隔离性，防止工具调用带来的安全风险。

### 第四阶段：可视化与可观测性 (第7-8周)
1. **[UI-Dashboard]** 增加平台概览页，显示在线 Agent 数、任务成功率等指标。
2. **[UI-Trace]** 实现 "协作链路图"，通过 Canvas 可视化展示任务在不同 Agent 间的流转。
3. **[UI-Market]** 增加 "Agent 市场" 界面，支持一键启用/禁用不同来源的 Agent。

## 4. 调整点清单
- **文件迁移**: 将部分 `engine` 内的逻辑下沉到 `common` 以便 SDK 引用。
- **API 变更**: 统一 SSE 事件格式，确保前端能解析复杂的 Trace 数据。
- **配置升级**: 引入 `config.yaml` 管理平台全局参数（如默认模型、Agent 注册有效期等）。

**请确认以上详细计划。确认后，我将首先创建 `docs/AGENT_PLATFORM_DESIGN.md` 详细设计文档并开始第一阶段的代码工作。**