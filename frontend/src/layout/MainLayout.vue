<template>
  <n-layout has-sider style="height: 100vh">
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
    <n-layout>
      <n-layout-content content-style="padding: 24px;">
        <router-view />
      </n-layout-content>
    </n-layout>
  </n-layout>
</template>

<script setup>
import { h, ref, computed } from 'vue'
import { NIcon } from 'naive-ui'
import { RouterLink, useRoute, useRouter } from 'vue-router'
import {
  HomeOutline as HomeIcon,
  ListOutline as ProjectIcon,
  HardwareChipOutline as ModelIcon,
  CodeSlashOutline as ScriptIcon,
  ChatbubbleOutline as ChatIcon
} from '@vicons/ionicons5'

const collapsed = ref(false)
const route = useRoute()
const router = useRouter()

const activeKey = computed(() => route.name)

function renderIcon(icon) {
  return () => h(NIcon, null, { default: () => h(icon) })
}

const menuOptions = [
  {
    label: 'Home',
    key: 'Home',
    icon: renderIcon(HomeIcon),
    onClick: () => router.push({ name: 'Home' })
  },
  {
    label: 'Projects',
    key: 'Projects',
    icon: renderIcon(ProjectIcon),
    onClick: () => router.push({ name: 'Projects' })
  },
  {
    label: 'Models',
    key: 'Models',
    icon: renderIcon(ModelIcon),
    onClick: () => router.push({ name: 'Models' })
  },
  {
    label: 'Scripts',
    key: 'Scripts',
    icon: renderIcon(ScriptIcon),
    onClick: () => router.push({ name: 'Scripts' })
  },
  {
    label: 'Chat',
    key: 'Chat',
    icon: renderIcon(ChatIcon),
    onClick: () => router.push({ name: 'Chat' })
  }
]

function handleUpdateValue(key) {
  // handled by onClick in options
}
</script>
