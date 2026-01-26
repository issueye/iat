# 产品需求文档 (PRD): A2A Agent 发现、发现与鉴权方案

## 1. 文档控制

- **版本**: v1.0
- **日期**: 2026-01-23
- **描述**: 定义 IAT 引擎与外部 Sub-Agents 建立连接、识别身份及发现能力的标准化流程。

---

## 2. 角色定义

- **IAT Host (Client)**: 桌面应用后端，发起协作请求的一方。
- **IAT Sub-Agent (Server)**: 独立进程或远程服务，接收任务并返回结果。

---

## 3. 需求详解

### 3.1 Agent 发现机制 (Discovery)

#### [F01] 自动发现 (局域网/跨进程)

- **方案**: 采用简单的 HTTP 探测或通过环境变量注入。
- **逻辑**: Host 扫描特定的本地端口范围（如 18080-18090），发送 `GET /manifest` 获取 Agent 元数据。

#### [F02] 静态配置

- **方案**: 在 UI 界面支持手动添加 A2A Endpoint。
- **信息**: `Name`, `URL`, `AuthSecret`, `Description`。

### 3.2 身份鉴权 (Authentication & Handshake)

#### [F03] 握手流程 (Handshake)

1. **Host -> Agent**: 发起携带 `HELO` 动作的消息，并附带 Host 的版本及支持的协议版本。
2. **Agent -> Host**: 返回其 `AgentID`、支持的 `tools` 列表及其 `Action` 适配表。

#### [F04] 令牌鉴权 (Token-based Auth)

- 所有请求的 Header 中必须包含 `X-IAT-Token`。
- Token 在 Agent 启动时通过环境变量 `IAT_SECRET` 同步，或在首次握手时由用户手动在 UI 确认授权产生。

### 3.3 生命周期管理 (Lifecycle)

- **[F05] 心跳维持 (Keep-Alive)**: Host 定期每 30s 发动一个 `action: pulse`，若 3 次未收到 Ack，标记该 Agent 为 "Offline"。
- **[F06] 资源清理**: Session 结束或 Host 关闭时，发送 `action: terminate` 强制要求 Sub-Agent 清理临时文件和进程。

---

## 4. 异常处理与降级

- **[E01] 版本不匹配**: 若协议版本不兼容，Host 需提示用户“Sub-Agent 版本过旧，部分功能（如流式输出）将无法使用”。
- **[E02] 连接中断**: 任务进行中连接断开，Host 应支持 3 次重连逻辑，失败后标记任务为 "Error: Agent Unreachable"。

---

## 5. UI 交互设计建议

1. **Agent 管理中心**: 在设置页增加 "A2A 智能体列表"，显示每个 Agent 的状态（在线/离线）、延时及支持的工具数量。
2. **连接弹窗**: 首次连接未授权 Agent 时，弹出“是否允许主引擎连接外部智能体 [AgentName]？”的风险提示。

---

**验收标准**: 成功跑通一轮从 Host 发现 Sub-Agent、握手交换工具集、到发送 execute 指令并正确回传心跳的全流程测试。
