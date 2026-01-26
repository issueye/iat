# 前端 Sub-Agent UI 实施计划

## 1. 现状分析

当前前端采用 Vue 3 + Naive UI 架构，通过 `Chat.vue` 中的 `fetch` 流式处理 SSE 事件。现有的事件处理机制仅限于主智能体的文本块（chunk）和基础工具调用（tool_call）。

## 2. 核心改动点

### 2.1 API 层扩展 (src/api/index.ts)

新增子任务管理相关接口：

- `abortSubAgentTask(taskId: string)`: 中止指定的异步子任务。
- `getSubAgentTasks(sessionId: number)`: 获取当前会话的所有子任务历史。

### 2.2 状态管理 (Chat.vue)

引入 `subAgentTaskMap` (Reactive Map):

- 键为 `taskId` (UUID)。
- 值为任务详情：`{ agentName, query, status, depth, chunks, result, error, children: [] }`。
- 通过 `parentTaskId` 建立父子引用关系，形成树状结构。

### 2.3 SSE 事件流接收 (Chat.vue -> handleSSEEvent)

扩展事件处理逻辑以支持子任务：

- **`subagent_start`**: 初始化任务对象，并将其挂载到父级任务或主消息下。
- **`subagent_chunk`**: 找到对应 `taskId`，追加中间输出内容（Thought 或中间过程）。
- **`subagent_done`**: 更新最终结果和状态（completed/failed）。

### 2.4 UI 组件化设计 (NEW: SubAgentCard.vue)

设计一个专门用于展示子任务的组件，具备以下特征：

- **嵌套渲染**: 支持递归调用自身以展示多级 Sub-Agent 协作。
- **状态感知**: 不同状态（Running, Done, Failed）采用不同的视觉风格。
- **交互功能**: 提供展开/折叠按钮点击查看详情，提供中止按钮（仅运行中状态可见）。

## 3. 详细交互逻辑

### 3.1 任务卡片状态

- **思考中 (Thinking)**: 显示子智能体的思维链（基于 `chunks` 累加）。
- **执行中 (Running)**: 显示正在调用的工具或中间结果。
- **任务归档**: 任务完成后，展示 `result` 摘要，点击展开可追溯完整过程。

### 3.2 布局分层

- 主对话流保持不变。
- Sub-Agent 任务以嵌套缩进的形式展示在发起该任务的消息/卡片下方，层级感清晰。

## 4. 验证计划

1. **功能验证**:
   - 触发一个耗时较长的 Sub-Agent 任务，验证 UI 是否实时出现 Loading 并展示中间内容。
   - 验证递归调用（如双层 Agent）是否呈现出树状缩进效果。
2. **交互验证**:
   - 点击中止按钮，验证 SSE 是否收到 `subagent_done` 事件且状态为 `aborted`。
   - 刷新页面后，通过 `getSubAgentTasks` 接口验证历史子任务能够正确还原。

---

**日期**: 2026-01-23
**编写人**: Antigravity
