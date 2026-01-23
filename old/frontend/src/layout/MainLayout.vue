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
          <div class="app-header__title">iat</div>
        </div>
      </n-layout-header>
      <n-layout-content :content-style="contentStyle">
        <router-view />
      </n-layout-content>
      <n-layout-footer bordered class="app-footer">
        <div class="app-footer__inner">
          <div class="app-footer__left">© iat Engine</div>
          <div class="app-footer__right">
            <n-tooltip trigger="hover">
              <template #trigger>
                <n-button quaternary circle size="small" @click="showScriptDocs = !showScriptDocs">
                  <n-icon size="18">
                    <CodeIcon />
                  </n-icon>
                </n-button>
              </template>
              脚本 API 文档
            </n-tooltip>
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

const collapsed = ref(false);
const showScriptDocs = ref(false);
const route = useRoute();
const router = useRouter();

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
  height: 100%;
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

.app-header__title {
  font-size: 16px;
  font-weight: 600;
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
</style>
