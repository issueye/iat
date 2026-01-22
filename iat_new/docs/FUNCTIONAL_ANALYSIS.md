# IAT (Issueye AI Tool) 功能分析报告

## 1. 项目概述

IAT (Issueye AI Tool) 是一个基于 Wails 框架构建的现代化 AI Agent 集成开发环境 (IDE) 辅助工具。它结合了 Go 语言的高性能后端与 Vue 3 的现代化前端，旨在通过对话式交互、多 Agent 协作和强大的工具生态，辅助开发者高效完成代码编写、项目管理、任务编排等工作。

## 2. 系统架构分析

### 2.1 技术栈

*   **前端**: Vue 3, TypeScript, Naive UI, Pinia, Vite.
*   **后端**: Go (Golang), Wails v2.
*   **AI 引擎**: CloudWeGo Eino.
*   **脚本引擎**: Goja (纯 Go 实现的 ECMAScript 引擎).
*   **数据库**: SQLite (GORM), GoLevelDB (倒排索引).
*   **通信协议**: SSE (流式响应), MCP (工具扩展).

### 2.2 架构分层 (Clean Architecture)

系统遵循 Clean Architecture 原则，结构清晰：
1.  **Presentation Layer (Frontend)**: 负责 UI 展示与交互，通过 Wails Bridge 和 SSE 与后端通信。
2.  **Application Layer (Wails App)**: `app.go` 作为粘合层，转发请求到 Service 层。
3.  **Service Layer (Business Logic)**: 包含核心业务逻辑 (`ChatService`, `AgentService` 等)。
4.  **Repository Layer (Data Access)**: 负责数据持久化 (SQLite)。
5.  **Infrastructure Layer (Pkg)**: 通用工具包 (AI Client, Script Engine, Tools)。

## 3. 核心功能模块详解

### 3.1 多 Agent 协作系统
*   **自定义 Agent**: 支持创建不同角色的 Agent，配置独立的 System Prompt 和模型。
*   **Sub-Agent 机制**: 主 Agent 可通过 `call_subagent` 递归调用子 Agent，实现任务分治。
*   **工作模式 (Mode)**: 预设不同模式（Chat, Plan, Build）以适应不同场景。

### 3.2 工具生态系统 (Tools & MCP)
系统支持三种类型的工具来源：
*   **内置工具**: 文件操作 (`read_file`, `list_files`), 系统命令 (`run_command`), 网络请求等。
*   **脚本工具**: 基于 `Goja` 运行时的动态脚本，支持用户编写 JS/Python 脚本即时扩展能力。
*   **MCP 工具**: 完整支持 Model Context Protocol，可连接外部 MCP Server (如 Github, Postgres) 扩展上下文。

### 3.3 任务编排系统
*   **结果导向**: 强调任务目标的达成。
*   **生命周期管理**: Agent 可自主创建、拆解、更新任务状态 (Pending -> InProgress -> Completed)。
*   **可视化**: 前端提供任务面板实时展示与交互。

### 3.4 智能对话与会话管理
*   **流式响应**: 使用 SSE 实现打字机效果。
*   **思维链展示**: 解析并独立展示 AI 的思考过程 (`<think>` 标签)。
*   **上下文管理**: 支持智能压缩、多项目隔离和历史回溯。

### 3.5 项目代码索引
*   **本地索引**: 使用 GoLevelDB 构建倒排索引。
*   **功能**: 支持 Agent 快速检索项目中的代码片段和定义。

## 4. 数据模型

核心实体包括：
*   **Project**: 项目基础信息。
*   **Session**: 会话上下文容器。
*   **Message**: 聊天记录 (Text/Tool)。
*   **Agent**: 智能体配置。
*   **Task**: 任务单元。
*   **Tool/Script/MCPServer**: 工具定义与配置。
*   **AIModel**: 模型连接配置。

## 5. 现状与局限性分析

根据现有分析报告 (`PROJECT_ANALYSIS.md`)：

### 5.1 优点
*   架构清晰，模块化程度高。
*   核心功能（会话、Agent、工具、任务）完成度较高。
*   支持 MCP 协议，扩展性强。

### 5.2 待改进点
*   **错误处理**: 部分工具缺乏结构化错误码。
*   **并发控制**: Sub-Agent 调用目前为同步阻塞，可能导致超时。
*   **测试覆盖**: 缺乏单元测试和集成测试。
*   **工具输出**: 虽然修复了 `list_files` 溢出问题，但整体防御性编程仍需加强。

## 6. 未来规划建议

1.  **深度能力构建**:
    *   引入向量数据库实现 RAG。
    *   集成 Tree-sitter 进行 AST 代码分析。
    *   实现异步 Agent 支持长耗时任务。

2.  **生态扩展**:
    *   建立 Skill 市场。
    *   实现多模型路由策略。
    *   探索 IDE 插件化方向。

3.  **工程化完善**:
    *   增强日志系统。
    *   完善单元测试。
    *   优化 UI/UX 细节 (如任务拖拽排序)。
