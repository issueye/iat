<template>
  <n-layout has-sider style="height: 100%">
    <n-layout-sider
      bordered
      collapse-mode="width"
      :collapsed-width="64"
      :width="240"
      :collapsed="collapsed"
      show-trigger
      @collapse="collapsed = true"
      @expand="collapsed = false"
    >
      <n-menu
        :collapsed="collapsed"
        :collapsed-width="64"
        :collapsed-icon-size="22"
        :options="menuOptions"
        :value="activeKey"
        @update:value="handleUpdateValue"
      />
    </n-layout-sider>
    <n-layout style="height: 100%">
      <n-layout-header bordered class="app-header">
        <div class="app-header__inner">
          <div class="app-header__left">
            <div class="app-header__title">iat</div>
            <n-breadcrumb class="app-breadcrumb">
              <n-breadcrumb-item
                v-for="(item, idx) in breadcrumbItems"
                :key="item.name"
              >
                <RouterLink
                  v-if="idx !== breadcrumbItems.length - 1"
                  :to="{ name: item.name }"
                >
                  {{ item.label }}
                </RouterLink>
                <span v-else>{{ item.label }}</span>
              </n-breadcrumb-item>
            </n-breadcrumb>
          </div>
        </div>
      </n-layout-header>
      <n-layout-content :content-style="contentStyle">
        <router-view v-slot="{ Component }">
          <keep-alive include="Chat">
            <component :is="Component" />
          </keep-alive>
        </router-view>
      </n-layout-content>
      <n-layout-footer bordered class="app-footer">
        <div class="app-footer__inner">
          <div class="app-footer__left">© iat Engine</div>
          <div class="app-footer__right">
            <n-tooltip trigger="hover">
              <template #trigger>
                <n-button
                  quaternary
                  circle
                  size="small"
                  @click="showScriptDocs = !showScriptDocs"
                >
                  <n-icon size="18">
                    <CodeIcon />
                  </n-icon>
                </n-button>
              </template>
              脚本 API 文档
            </n-tooltip>

            <div class="engine-status" title="Engine 状态">
              <span class="engine-status__dot" :class="engineStatusDotClass" />
              <span class="engine-status__text"
                >Engine: {{ engineStatusLabel }}</span
              >
            </div>
          </div>
        </div>
      </n-layout-footer>
    </n-layout>

    <!-- Floating Script Docs Window -->
    <ScriptDocsFloat :show="showScriptDocs" @close="showScriptDocs = false" />
  </n-layout>
</template>

<script setup>
import { h, ref, computed } from "vue";
import { NIcon } from "naive-ui";
import { RouterLink, useRoute, useRouter } from "vue-router";
import {
  HomeOutline as HomeIcon,
  ListOutline as ProjectIcon,
  HardwareChipOutline as ModelIcon,
  PeopleOutline as AgentIcon,
  HammerOutline as ToolIcon,
  ChatbubbleOutline as ChatIcon,
  OptionsOutline as ModeIcon,
  CodeSlashOutline as CodeIcon,
  ServerOutline as MCPIcon,
} from "@vicons/ionicons5";
import ScriptDocsFloat from "../components/ScriptDocsFloat.vue";
import { useEngineStatus } from "../useEngineStatus";

const collapsed = ref(true);
const showScriptDocs = ref(false);
const route = useRoute();
const router = useRouter();
const { engineStatus, engineStatusLabel } = useEngineStatus();

const routeTitleMap = {
  Home: "首页",
  Projects: "项目列表",
  Models: "模型管理",
  Agents: "智能体管理",
  Tools: "工具管理",
  MCPs: "MCP 管理",
  Modes: "模式管理",
  Chat: "智能对话",
};

const breadcrumbItems = computed(() => {
  return route.matched
    .filter((r) => typeof r.name === "string" && r.name)
    .map((r) => ({
      name: r.name,
      label: routeTitleMap[r.name] || r.name,
    }));
});

const engineStatusDotClass = computed(() => {
  if (engineStatus.value === "Online") return "engine-status__dot--online";
  if (engineStatus.value === "Offline") return "engine-status__dot--offline";
  if (engineStatus.value === "Error") return "engine-status__dot--error";
  return "engine-status__dot--checking";
});

const activeKey = computed(() => {
  // If showing docs and not on a specific route that matches, maybe we don't highlight?
  // Or just highlight current route. ScriptDocs menu item won't be highlighted unless we set a fake key.
  return route.name;
});

const contentStyle = computed(() => {
  return route.name === "Chat"
    ? "padding: 0; height: 100%"
    : "padding: 24px; min-height: 100%;";
});

function renderIcon(icon) {
  return () => h(NIcon, null, { default: () => h(icon) });
}

const menuOptions = [
  {
    label: "首页",
    key: "Home",
    icon: renderIcon(HomeIcon),
    onClick: () => router.push({ name: "Home" }),
  },
  {
    label: "项目列表",
    key: "Projects",
    icon: renderIcon(ProjectIcon),
    onClick: () => router.push({ name: "Projects" }),
  },
  {
    label: "模型管理",
    key: "Models",
    icon: renderIcon(ModelIcon),
    onClick: () => router.push({ name: "Models" }),
  },
  {
    label: "智能体管理",
    key: "Agents",
    icon: renderIcon(AgentIcon),
    onClick: () => router.push({ name: "Agents" }),
  },
  {
    label: "工具管理",
    key: "Tools",
    icon: renderIcon(ToolIcon),
    onClick: () => router.push({ name: "Tools" }),
  },
  {
    label: "MCP 管理",
    key: "MCPs",
    icon: renderIcon(MCPIcon),
    onClick: () => router.push({ name: "MCPs" }),
  },
  {
    label: "模式管理",
    key: "Modes",
    icon: renderIcon(ModeIcon),
    onClick: () => router.push({ name: "Modes" }),
  },
  {
    label: "智能对话",
    key: "Chat",
    icon: renderIcon(ChatIcon),
    onClick: () => router.push({ name: "Chat" }),
  },
];

function handleUpdateValue(key) {
  // handled by onClick in options
}
</script>
<style scoped>
:deep(.n-layout) {
  height: calc(100% - 48px - 45px);
}

.app-header {
  height: 48px;
}

.app-header__inner {
  height: 48px;
  padding: 0 16px;
  display: flex;
  align-items: center;
}

.app-header__left {
  display: flex;
  align-items: center;
  gap: 16px;
  min-width: 0;
}

.app-header__title {
  font-size: 16px;
  font-weight: 600;
}

.app-breadcrumb {
  font-size: 12px;
  opacity: 0.85;
  min-width: 0;
}

.app-breadcrumb :deep(a) {
  color: inherit;
  text-decoration: none;
}

.app-breadcrumb :deep(a:hover) {
  text-decoration: underline;
}

.app-footer {
  height: 44px;
}

.app-footer__inner {
  height: 44px;
  padding: 0 16px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.app-footer__left {
  opacity: 0.8;
  font-size: 12px;
}

.app-footer__right {
  display: flex;
  align-items: center;
  gap: 12px;
}

.engine-status {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 12px;
  opacity: 0.9;
  user-select: none;
}

.engine-status__dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
}

.engine-status__dot--online {
  background: #18a058;
}

.engine-status__dot--offline {
  background: #d03050;
}

.engine-status__dot--error {
  background: #f0a020;
}

.engine-status__dot--checking {
  background: #909399;
}
</style>
