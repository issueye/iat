# 项目分析与优化报告

**生成时间**: 2026-01-17

## 1. 问题分析：`list_files` 调用后会话结束

### 现象描述
用户在对话中调用 `list_files` 工具成功获取文件列表后，会话意外终止，LLM 没有对结果进行总结或继续执行。

### 原因分析
经过对 `internal/service/chat_service.go` 和 `internal/pkg/tools/builtin/impl.go` 的代码审查，发现潜在原因如下：
1.  **Context Window 溢出**: 原有的 `ListFiles` 实现没有对输出内容进行限制。如果目标目录包含大量文件（如 `node_modules` 或 `.git`），工具返回的字符串可能极大，导致下一次调用 LLM 时超出了 Context Window 限制，或导致网络传输异常。
2.  **缺乏保护机制**: 系统缺乏对工具输出过大的防御性编程，导致异常状态下静默失败。

### 解决方案（已修复）
我们对 `internal/pkg/tools/builtin/impl.go` 进行了以下优化：
1.  **增加忽略列表**: 默认忽略 `.git`, `node_modules`, `dist` 等常见的非业务大目录。
2.  **增加数量限制**: 限制单次列出的文件数量上限为 500 个，超过部分截断并提示。
3.  **增加大小限制**: 对 `ReadFile` 增加了 100KB 的大小限制，引导用户使用 `read_file_range` 读取大文件。

---

## 2. 项目现状分析

### 架构与技术栈
*   **架构模式**: Clean Architecture (Model -> Repo -> Service -> App)。结构清晰，职责分明。
*   **后端**: Go + Wails + Eino (AI) + GORM (SQLite)。
*   **前端**: Vue 3 + Naive UI + TypeScript。
*   **通信**: SSE (Server-Sent Events) 实现流式响应，MCP 协议扩展能力。

### 核心功能完成度
*   **会话管理**: ✅ (支持多项目、压缩、搜索)
*   **Agent 系统**: ✅ (支持自定义角色、Sub-Agent 递归调用)
*   **工具系统**: ✅ (内置文件/命令/网络工具，支持 Script 和 MCP)
*   **任务编排**: ✅ (Result-Oriented Task Management)

### 代码质量评估
*   **优点**: 模块化程度高，代码风格统一，关键路径（如 ChatService）逻辑清晰。
*   **改进点**:
    *   **错误处理**: 部分工具错误直接返回字符串，缺乏结构化错误码，不利于前端精确提示。
    *   **并发控制**: Sub-Agent 调用目前是同步阻塞的，对于长任务可能会导致主会话超时。
    *   **测试覆盖**: 缺乏单元测试和集成测试，重构风险较高。

---

## 3. 优化与阶段开发计划

### Phase 1: 健壮性与体验优化 (当前阶段)
*   [x] **工具输出限制**: 修复 `list_files` 和 `read_file` 的潜在溢出问题。
*   [ ] **错误处理增强**: 统一工具调用的错误返回格式。
*   [ ] **日志系统**: 完善后端日志，记录关键路径的 Request/Response，便于排查问题。
*   [ ] **UI 细节**: 优化任务面板的交互体验，支持拖拽排序任务。

### Phase 2: 深度能力构建 (短期计划)
*   [ ] **RAG (检索增强生成)**: 引入向量数据库，对项目代码进行 Embedding 索引，实现基于语义的代码问答。
*   [ ] **AST 代码分析**: 集成 Tree-sitter，让 Agent 能更精准地理解代码结构（函数、类定义），而不仅仅是文本搜索。
*   [ ] **异步 Agent**: 实现后台运行的 Agent，支持耗时较长的任务（如全量重构、测试运行），并通过通知系统告知用户。

### Phase 3: 生态与扩展 (中长期计划)
*   [ ] **Skill 市场**: 允许用户导入/导出 Skill 包（一组 Prompt + Tool 的集合）。
*   [ ] **多模型路由**: 根据任务复杂度自动选择模型（如简单任务用 Flash 模型，复杂任务用 Pro 模型）。
*   [ ] **IDE 插件化**: 尝试将核心能力封装为 VSCode/JetBrains 插件，而不仅仅是独立应用。

---

## 4. 项目 Q&A

**Q: 为什么 Agent 有时候会重复调用同一个工具？**
A: 这通常是因为 LLM 对工具返回的结果不满意或没看懂。建议检查工具的返回格式是否清晰，或者在 Agent 的 System Prompt 中增加明确的指令。

**Q: Sub-Agent 是如何工作的？**
A: 主 Agent 通过 `call_subagent` 工具发起调用。后端会创建一个独立的上下文环境（包括独立的 System Prompt 和工具集）来运行 Sub-Agent。目前这是同步调用的，即主 Agent 会等待 Sub-Agent 完成后才继续。

**Q: 如何添加自定义工具？**
A: 目前可以通过两种方式：
1.  **Script**: 在前端编写 JS/Python 脚本并保存，Agent 即可调用。
2.  **MCP**: 启动一个 MCP Server，并在 Agent 设置中绑定该 Server。

**Q: 任务编排系统的数据存在哪里？**
A: 任务数据存储在本地 SQLite 数据库的 `tasks` 表中，与 `sessions` 关联。

**Q: 如果我想让 Agent 能够操作浏览器怎么办？**
A: 推荐使用 MCP 协议。可以运行一个支持 Puppeteer/Playwright 的 MCP Server，然后将其绑定到 Agent，无需修改核心代码。
