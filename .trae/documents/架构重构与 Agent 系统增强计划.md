# IAT 架构重构与 Agent 系统增强计划 (Go Workspace 版)

根据您的最新指示，我们将采用 `go.work` 工作区模式管理项目，并按照 `ui`、`engine`、`common` 的结构进行拆分。

## 1. 项目结构规划 (Go Workspace)

采用多模块管理，根目录下包含 `go.work` 文件。

```text
iat_new/
├── go.work            # Workspace 配置文件
├── common/            # [Module] 公共依赖模块
│   ├── go.mod
│   ├── model/         # 数据库模型与实体定义 (Project, Task, Agent)
│   ├── protocol/      # Agent 通信协议 (A2A) 定义
│   └── pkg/           # 通用工具 (Logger, Result, Utils)
├── engine/            # [Module] 核心后端服务
│   ├── go.mod
│   ├── cmd/           # 启动入口 (main.go)
│   ├── api/           # HTTP API & SSE Handlers
│   ├── runtime/       # Agent 运行时与调度逻辑
│   └── service/       # 业务服务层
└── ui/                # [Module] Wails 桌面客户端
    ├── go.mod
    ├── main.go        # Wails 启动入口
    ├── app.go         # 进程管理 Bridge (启动 Engine)
    └── frontend/      # Vue 3 前端资源
```

## 2. 架构重构实施细节

### 2.1 Common 模块
*   **职责**: 定义所有模块共享的基础数据结构和协议。
*   **内容**: 迁移 `old/internal/model` 到 `common/model`，确保前后端（如果 UI 需要 Go 代码）和 Engine 共享同一套模型。

### 2.2 Engine 模块 (Core)
*   **通信**: 实现 HTTP Server (基于 `net/http` 或 `gin`)。
    *   `POST /api/v1/agent/run`: 触发任务。
    *   `GET /api/v1/events`: SSE 流式输出。
*   **逻辑迁移**: 将原有的 `Service` 层逻辑迁移至此，并适配新的 HTTP 接口。

### 2.3 UI 模块 (Desktop Shell)
*   **Wails 集成**: 仅保留 Wails 作为 UI 容器和系统交互层。
*   **Engine 托管**: 在 Wails 启动时，自动在后台启动 Engine 进程，并随机分配或指定端口。
*   **前端改造**: 将 API 请求指向本地 Engine 服务端口。

## 3. Agent 系统分离与增强

### 3.1 分离式 Agent Runtime (在 Engine 中实现)
*   **独立实例**: 每个 Agent 实例化为一个对象，包含独立的 `Context` 和 `ToolSet`。
*   **A2A 协议**: 在 `common/protocol` 中定义消息结构，Engine 负责消息路由。

### 3.2 主 Agent (Orchestrator) 增强
*   **功能实现**:
    *   **协调**: 实现 `DispatchTask` 工具。
    *   **评审**: 实现 `ReviewOutput` 工具。
    *   **裁决**: 在 Prompt 中植入决策逻辑。
    *   **共享工具库**: 在 Engine 中实现 `GlobalToolRegistry`，并提供 `GrantTool` 接口供主 Agent 调用。

## 4. 实施步骤

1.  **初始化工作区**: 创建目录结构，初始化 `go.mod` 和 `go.work`。
2.  **Common 迁移**: 提取模型和通用工具代码。
3.  **Engine 开发**: 搭建 HTTP/SSE 服务框架，实现 Agent Runtime。
4.  **UI 适配**: 创建新的 Wails 项目结构，集成前端代码，实现进程管理。
5.  **联调验证**: 验证 UI -> Engine 的通信链路及 Agent 协作流程。
