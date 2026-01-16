我将更新项目文档以包含“AI 模型管理”功能。具体修改计划如下：

1.  **更新 [task/需求.md](file:///e:/code/issueye/suwei/iat/task/需求.md)**
    - 在“功能”章节下新增 **AI 模型管理模块**，明确支持模型的增删改查及配置（如 BaseURL, API Key, 模型名称）。

2.  **更新 [task/需求分析.md](file:///e:/code/issueye/suwei/iat/task/需求分析.md)**
    - 在 **3.1 核心业务模块** 中新增 **3.1.7 AI 模型 (Model) 管理**。
    - 详细描述功能点：
        - 支持多提供商（OpenAI, DeepSeek, Ollama 等）。
        - 模型参数配置（Temperature, Token Limit 等）。
        - 连通性测试功能。
    - 补充说明 Agent 与模型的关联逻辑。

3.  **更新 [task/开发计划.md](file:///e:/code/issueye/suwei/iat/task/开发计划.md)**
    - 在 **Phase 3: 业务功能开发 - 基础** 中添加具体的开发任务：
        - 数据库：设计 `Model` 表。
        - 后端：实现模型 CRUD 及连接测试 Service。
        - 前端：开发模型管理配置页面。

确认后，我将立即执行文档更新。