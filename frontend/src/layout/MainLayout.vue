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
      <n-layout-content :content-style="contentStyle">
        <router-view />
      </n-layout-content>
    </n-layout>
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
} from "@vicons/ionicons5";

const collapsed = ref(false);
const route = useRoute();
const router = useRouter();

const activeKey = computed(() => route.name);

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
</style>
