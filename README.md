# iat - Intelligent Agent Tool

`iat` 是一个基于桌面端的 AI 辅助工具，提供类似 Claude Code 的交互体验。结合本地脚本引擎 (`goja`) 与先进 AI 模型 (`eino`)，为开发者提供灵活的 Agent 定义与执行环境。

## 功能特性

*   **项目管理**: 多项目隔离，清晰管理不同任务上下文。
*   **AI 模型管理**: 支持配置 OpenAI, DeepSeek, Ollama 等多种模型提供商，支持连接测试。
*   **脚本系统**: 内置 JS 脚本引擎，支持编写、测试和运行自动化脚本，可作为 Agent 的工具使用。
*   **Agent 编排**: 自定义 Agent 角色、System Prompt 及关联模型。
*   **智能对话**:
    *   流式响应 (SSE)，实时输出 AI 回复。
    *   Markdown 渲染与代码高亮。
    *   自动保存会话历史。

## 技术栈

*   **前端**: Vue 3, Naive UI, Pinia, Vue Router, Vite
*   **后端**: Go (Wails v2), Gorm (SQLite), Eino (AI Framework), Goja (JS Engine)

## 快速开始

### 开发环境

1.  **环境要求**:
    *   Go 1.21+
    *   Node.js 16+
    *   NPM / Yarn

2.  **安装依赖**:
    ```bash
    # 后端依赖
    go mod tidy

    # 前端依赖
    cd frontend
    npm install
    ```

3.  **运行开发模式**:
    ```bash
    wails dev
    ```
    应用启动后，可以在浏览器访问 `http://localhost:34115` 进行调试，或直接使用弹出的桌面窗口。

### 构建发布

```bash
wails build
```
构建产物将生成在 `build/bin/iat.exe`。

## 使用指南

1.  **配置模型**: 进入 "Models" 页面，添加你的 API Key (如 OpenAI 或 DeepSeek)。点击 "Test Connection" 验证。
2.  **创建 Agent**: 进入 "Agents" 页面，定义一个新的 Agent (如 "Coding Assistant")，并绑定刚才配置的模型。
3.  **开始对话**: 进入 "Chat" 页面，选择一个项目，点击 "New Chat"，选择刚才创建的 Agent，即可开始对话。
4.  **编写脚本**: 进入 "Scripts" 页面，可以编写 JS 脚本来辅助任务 (如 `http.get`, `console.log`)。

## 目录结构

*   `cmd/`: 辅助工具与测试代码。
*   `frontend/`: Vue 前端源码。
*   `internal/`: Go 后端源码。
    *   `model/`: 数据库模型。
    *   `pkg/`: 核心组件 (ai, script, sse, db)。
    *   `repo/`: 数据访问层。
    *   `service/`: 业务逻辑层。
*   `app.go`: Wails 应用主入口与 API 绑定。
*   `main.go`: 程序入口。

## License

MIT
