# 架构方案对比：gRPC vs HTTP

在决定是否采用 gRPC 之前，以下是针对本项目（桌面端 AI 助手）的两种方案深度对比：

| 维度 | gRPC (推荐) | HTTP API (REST/JSON) |
| :--- | :--- | :--- |
| **核心优势** | **流式支持 (Streaming)** | **简单易用 (Simplicity)** |
| **通信协议** | HTTP/2 + Protobuf (二进制) | HTTP/1.1 + JSON (文本) |
| **AI 对话体验** | **完美支持**: 原生双向流，极低延迟，适合实现 ChatGPT 式的“打字机”效果。 | **较差**: 需要使用 SSE (Server-Sent Events) 或 WebSocket，在 Go 客户端实现上比 gRPC 麻烦。 |
| **类型安全** | **强**: `.proto` 文件定义契约，自动生成 Client/Server 代码。接口变更时编译会报错，避免运行时错误。 | **弱**: Server 和 Client 需手动保持 Struct 同步。容易出现字段拼写错误或类型不匹配。 |
| **开发成本** | **前期高**: 需安装 `protoc`，编写 `.proto` 文件。 | **前期低**: 也就是写普通的 Web Server (Gin/Echo)，调试方便 (Postman)。 |
| **性能** | **极高**: 序列化效率高，适合频繁的内部调用。 | **高**: 本地调用差异不明显，但 JSON 解析有额外开销。 |

### 结论与建议
**推荐选择 gRPC**。
理由：
1.  **AI 核心需求**: 项目核心是 Chat，AI 响应是流式的。gRPC 的 `stream` 是处理这种场景的标准答案，代码比手动处理 HTTP Chunk 或 SSE 更干净。
2.  **代码健壮性**: 分离 UI 和 Engine 后，两者代码库可能物理分离。gRPC 的强类型契约能防止“改了后端接口忘了改前端调用”这类低级错误。

---

# 实施计划 (基于 gRPC)

如果您同意使用 gRPC，我们将执行以下计划：

## 1. 目录结构调整
```text
iat/
├── api/proto/              # Protobuf 定义
│   ├── model.proto         # 数据模型 (Project, Task...)
│   └── service.proto       # 服务接口 (Chat, ProjectService...)
├── cmd/
│   ├── engine/             # [新增] Engine 进程入口
│   └── app/                # Wails UI 入口
├── internal/
│   ├── grpc/               # [新增] gRPC Handler 实现
│   └── service/            # 现有业务逻辑 (复用)
```

## 2. 实施步骤

### Step 1: 协议定义 (Protobuf)
*   编写 `model.proto`: 映射 `internal/model` 中的核心结构。
*   编写 `service.proto`: 定义 `Chat` (流式), `Project`, `Task` 等服务接口。

### Step 2: Engine 开发 (Server)
*   建立 `cmd/engine/main.go`，初始化 DB 和 Listener。
*   在 `internal/grpc` 中实现 Handler，封装现有的 `internal/service` 调用。
*   **重点**: 将 `ChatService` 的 SSE 推送改造为 gRPC Stream 发送。

### Step 3: UI 适配 (Client)
*   在 `app.go` 中引入 gRPC Client。
*   替换原有的 Service 直接调用为 gRPC 远程调用。
*   实现子进程管理：Wails 启动时自动拉起 Engine。

## 3. 立即行动
确认后，我将首先创建 `api/proto` 目录并编写初步的 `.proto` 文件，定义核心数据结构和服务接口。
