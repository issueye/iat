# IAT (Issueye AI Tool) 项目设计与功能说明书

## 1. 项目概述

IAT (Issueye AI Tool) 是一个基于 Wails 框架构建的现代化 AI Agent 集成开发环境 (IDE) 辅助工具。旨在通过对话式交互、多 Agent 协作和强大的工具生态，辅助开发者高效完成代码编写、项目管理、任务编排等工作。

本项目结合了 Go 语言的高性能后端与 Vue 3 的现代化前端，集成 Model Context Protocol (MCP) 和 CloudWeGo Eino 框架，提供了一个可扩展、智能化的开发助手平台。

## 2. 系统架构

### 2.1 技术栈

*   **后端 (Backend)**:
    *   **语言**: Go (Golang)
    *   **应用框架**: [Wails v2](https://wails.io/) (负责 GUI 桥接与资源管理)
    *   **AI 引擎**: [Eino](https://github.com/cloudwego/eino) (CloudWeGo 出品的 AI 编排框架)
    *   **数据库**: SQLite (配合 GORM ORM), GoLevelDB (用于倒排索引)
    *   **脚本引擎**: Goja (纯 Go 实现的 ECMAScript 引擎)
    *   **通信协议**: SSE (Server-Sent Events) 用于流式对话，MCP (Model Context Protocol) 用于工具扩展

*   **前端 (Frontend)**:
    *   **框架**: Vue 3 + TypeScript
    *   **UI 组件库**: Naive UI
    *   **状态管理**: Pinia (支持持久化)
    *   **构建工具**: Vite
    *   **网络**: Axios (HTTP), EventSource (SSE)

### 2.2 架构分层

系统遵循 Clean Architecture 原则，主要分为以下几层：

1.  **Presentation Layer (Frontend)**:
    *   负责 UI 展示与用户交互。
    *   通过 Wails Bridge 调用后端方法，或通过 HTTP/SSE 与后端通信。
2.  **Application Layer (Wails App)**:
    *   `app.go` 定义了暴露给前端的方法。
    *   作为前后端的粘合层，转发请求到 Service 层。
3.  **Service Layer (Business Logic)**:
    *   包含核心业务逻辑，如 `ChatService`, `AgentService`, `TaskService` 等。
    *   负责协调 Repo 层和 AI 引擎。
4.  **Repository Layer (Data Access)**:
    *   负责数据的持久化存储 (SQLite)。
    *   提供 CRUD 接口供 Service 层调用。
5.  **Infrastructure Layer (Pkg)**:
    *   包含通用工具包，如 `pkg/ai` (AI Client), `pkg/tools` (工具实现), `pkg/script` (脚本引擎)。

## 3. 核心功能模块详解

### 3.1 多 Agent 协作系统 (Agent System)

*   **自定义 Agent**: 支持创建不同角色的 Agent，每个 Agent 拥有独立的 System Prompt、模型配置和工具集。
*   **Sub-Agent 机制**: 主 Agent 可以通过 `call_subagent` 工具递归调用其他 Agent，实现任务的分层与分治。
*   **模式 (Mode)**: 支持不同的工作模式（如 Chat, Plan, Build），不同模式预设了不同的 Prompt 模板和工具权限。

### 3.2 工具生态系统 (Tools & MCP)

系统支持多种类型的工具来源，构建了强大的扩展能力：

*   **内置工具 (Builtin Tools)**:
    *   **文件操作**: `read_file`, `write_file`, `list_files`, `diff_file` 等。
    *   **系统命令**: `run_command` (Shell 执行)。
    *   **网络请求**: `http_get`, `http_post`。
    *   **项目索引**: `index_project` (基于倒排索引的代码搜索)。
*   **脚本工具 (Script Tools)**:
    *   基于 `Goja` 运行时，支持动态加载和执行 JS/Python (通过外部解释器) 脚本。
    *   允许用户在前端编写脚本并即时生效。
*   **MCP 工具 (Model Context Protocol)**:
    *   完整支持 MCP 协议，可连接任意标准的 MCP Server。
    *   支持 `call_tool`, `list_tools`, `list_resources` 等 MCP 标准操作。
    *   实现了 MCP Client，允许 IAT 作为宿主连接外部生态工具（如 Github, Postgres 等 MCP Server）。

### 3.3 任务编排系统 (Task Orchestration)

*   **结果导向 (Result-Oriented)**: 强调任务目标的达成。
*   **任务管理**:
    *   Agent 可自主创建、拆解、更新和完成任务。
    *   支持任务状态流转：Pending -> InProgress -> Completed。
*   **可视化**: 前端 Side Panel 实时展示任务列表，支持查看任务详情和状态。
*   **数据模型**: 任务与 Session 关联，保证上下文的一致性。

### 3.4 智能对话与会话管理

*   **流式响应 (SSE)**: 使用 Server-Sent Events 实现打字机效果，提升交互体验。
*   **思维链展示**: 解析模型输出的 `<think>` 标签，独立展示思考过程。
*   **上下文管理**:
    *   **智能压缩**: 当上下文过长时，自动触发压缩机制，保留关键信息。
    *   **多项目隔离**: 会话归属于特定项目，保证环境隔离。
    *   **历史回溯**: 支持搜索和查看历史会话。

### 3.5 项目代码索引 (Code Indexing)

*   **倒排索引**: 使用 GoLevelDB 构建本地倒排索引。
*   **语义搜索 (规划中)**: 未来计划引入向量数据库支持语义检索 (RAG)。
*   **功能**: 支持 Agent 快速检索项目中的代码片段、函数定义等。

## 4. 数据模型设计 (Data Models)

核心实体定义在 `internal/model` 包中：

| 实体 (Entity) | 描述 (Description) | 关键字段 |
| :--- | :--- | :--- |
| **Project** | 项目 | `Name`, `Path`, `Description` |
| **Session** | 会话 | `ProjectID`, `Title`, `Summary` |
| **Message** | 消息 | `SessionID`, `Role`, `Content`, `Type` (Text/Tool) |
| **Agent** | 智能体 | `Name`, `Description`, `SystemPrompt`, `ModelConfig` |
| **Task** | 任务 | `SessionID`, `ParentID`, `Content`, `Status` |
| **Tool** | 工具定义 | `Name`, `Type` (Builtin/Script/MCP), `Schema` |
| **Script** | 脚本 | `Name`, `Content`, `Language`, `Parameters` |
| **MCPServer** | MCP 服务配置 | `Name`, `Command`, `Args`, `Env` |
| **AIModel** | 模型配置 | `Provider`, `ModelName`, `BaseURL`, `APIKey` |

## 5. 接口与交互设计

### 5.1 Wails Bridge
前端通过 `window.go` 对象调用后端方法，如：
*   `ListProjects()`
*   `CreateSession(projectId)`
*   `GetAgents()`

### 5.2 SSE 通信
*   **Endpoint**: `/api/chat/stream`
*   **Events**:
    *   `message`: 文本消息片段。
    *   `tool_call`: 工具调用请求。
    *   `tool_result`: 工具执行结果。
    *   `error`: 异常信息。
    *   `done`: 生成结束。

## 6. 安全性设计

*   **命令执行限制**: 敏感命令需要用户确认 (规划中)。
*   **文件访问控制**: 限制 Agent 只能访问项目目录下的文件，防止越权访问系统文件。
*   **API Key 管理**: 模型 API Key 本地加密存储。

## 7. 未来规划

*   **Skill 市场**: 建立 Skill 分享与导入机制。
*   **多模态支持**: 支持图片、音频等多模态输入输出。
*   **IDE 插件化**: 将核心能力通过 LSP 或插件形式集成到 VSCode/JetBrains。
