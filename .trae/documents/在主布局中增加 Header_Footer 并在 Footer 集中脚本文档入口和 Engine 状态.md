## 总体思路
- 在当前主布局组件 MainLayout.vue 上下增加统一的 Header 和 Footer，引导所有页面共享同一套顶/底栏。
- 将原来侧边菜单里的“脚本API文档”入口从菜单中移除，改为放到 Footer 中一个图标按钮上，仍然弹出现有的 ScriptDocsFloat 浮窗。
- 在 Footer 中接入现有 `/api/health` 接口，实时展示 Engine 状态（在线/离线/错误），并配合颜色/小图标做简单状态提示。

## 现有结构梳理（只读）
- 主布局壳：`iat_new/ui/frontend/src/layout/MainLayout.vue`
  - 使用 `n-layout` + `n-layout-sider` + `n-menu` + `n-layout-content` 包裹 `<router-view />`。
  - 菜单中已有 `label: "脚本API文档"`，点击后切换 `showScriptDocs`，并在模板中渲染 `<ScriptDocsFloat :show="showScriptDocs" @close="showScriptDocs = false" />`。
- Engine 健康检查：
  - 前端 API：`src/api/index.ts` 中 `api.checkHealth()` 调用 `/api/health`。
  - 后端路由：`engine/api/server.go` 中 `/api/health` 直接返回 `"ok"`。
  - Home 页面已有示例：`views/Home.vue` 使用 `api.checkHealth()` 在本页显示 Engine Status。

## 详细实现计划

### 一、在 MainLayout 增加 Header/Footer 框架
- 在 `MainLayout.vue` 的模板中：
  - 保留现有 `n-layout-sider` + `n-layout-content` 结构不变。
  - 在最外层 `n-layout` 内部增加：
    - 顶部：`<n-layout-header>` 作为 Header，展示应用标题（如 "iat"）和简单占位区域（为后续可能的全局按钮预留）。
    - 底部：`<n-layout-footer>` 作为 Footer，内部放置：
      - 左侧：版权/版本等占位文本（可先放简单文案，例如 "iat Engine"）。
      - 右侧：脚本文档图标按钮 + Engine 状态显示区域。
- 样式层面：
  - 复用现有布局的暗色/亮色配置（参考当前 `n-layout-sider` 的风格），通过简单的内联 class 或 scoped 样式控制 Header/Footer 高度和内边距。

### 二、将“脚本API文档”入口移到 Footer 图标
- 在 `MainLayout.vue` 中修改菜单配置：
  - 从 `menuOptions` 中删除当前 "脚本API文档" 菜单项，避免在侧边栏和 Footer 同时出现重复入口。
  - 保留/复用现有的 `showScriptDocs` 响应式变量以及 `<ScriptDocsFloat>` 组件挂载位置不变。
- 在 Footer 中新增图标按钮：
  - 使用 Naive UI 的 `n-button` + `n-icon` 组合（例如文档/书本图标，优先选择项目已有的图标库；若无，则使用简单文本按钮先实现）。
  - 点击按钮时切换 `showScriptDocs`，打开/关闭 `ScriptDocsFloat`。
  - 为按钮增加 `n-tooltip` 或 `title` 属性，提示文案为“脚本 API 文档”。

### 三、在 Footer 显示 Engine 状态
- 复用现有 `api.checkHealth()` 能力，在布局级别统一展示 Engine 状态：
  - 在 `MainLayout.vue` 中新增：
    - `engineStatus`：字符串或枚举，例如 "Checking" | "Online" | "Offline" | "Error"。
    - `engineStatusColor`：用于 Footer 中状态点的颜色（如 green/red/orange），可根据 `engineStatus` 计算得到。
  - 在 `onMounted` 钩子中：
    - 调用一次 `api.checkHealth()`，根据返回值设置状态：
      - 请求成功并返回预期数据 → `Online`。
      - 请求失败（异常） → `Offline`。
  - 可选：增加定时轮询（例如每 10 秒调用一次），在 `onUnmounted` 中清理 `setInterval`，保证状态相对实时。
- Footer 展示形式：
  - 在右侧区域显示：
    - 一个小圆点或图标（使用 `n-badge` 或自定义 `span` + `border-radius`），颜色与状态对应。
    - 文本如：`Engine: Online` / `Engine: Offline`。
  - 当状态为 `Checking` 或 `Error` 时，使用不同颜色/文案区分，便于排查。

### 四、是否抽取复用逻辑（可选优化）
- 为避免 Home 页和 Footer 重复各自维护 Engine 状态，可以考虑：
  - 新建 `src/composables/useEngineStatus.ts`：
    - 封装 `engineStatus` 响应式变量、`checkEngineHealth` 方法和可选的轮询逻辑。
  - `Home.vue` 和 `MainLayout.vue` 都通过该 composable 获取状态，实现逻辑复用。
- 如果当前需求以尽快在 Footer 展示为主，可先在 `MainLayout` 内部实现，后续再抽取为 composable。

### 五、老前端（old/frontend）的处理策略
- 当前计划优先在 `iat_new/ui/frontend` 这套新前端上实现：
  - 如果你仍在使用 `old/frontend`，可以在确认新方案后，将同样的布局/入口改造同步到旧目录中的 `layout/MainLayout.vue`，保持用户体验一致。

### 六、交互与文案细节
- Header 文案：简洁展示，例如 "iat" 或 "iat Studio"。
- Footer 左侧：预留简单文案，如“© iat Engine”，方便后续补充版本号等信息。
- Footer 按钮与状态：
  - 脚本文档按钮：图标+tooltip，点击打开 ScriptDocsFloat。
  - Engine 状态：使用醒目但不过分抢眼的颜色区分 Online/Offline/Error，尽量不干扰主要内容区域。

如果你确认这个方案，我会在 MainLayout.vue 等相关文件中按上述步骤实现，并确保不影响现有路由和 ScriptDocsFloat 的行为。