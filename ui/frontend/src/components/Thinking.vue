<template>
  <div class="thinking-block" :class="{ 'is-collapsed': collapsed }">
    <div class="thinking-header" @click="collapsed = !collapsed">
      <div class="header-left">
        <n-icon
          :component="collapsed ? ChevronForwardOutline : ChevronDownOutline"
          class="toggle-icon"
        />
        <span class="thinking-title">思考过程</span>
      </div>
      <div v-if="!content && !collapsed" class="thinking-status">
        <n-spin size="small" />
      </div>
    </div>
    <div v-show="!collapsed" class="thinking-content">
      <div v-if="content" class="thinking-text">{{ content }}</div>
      <div v-else class="thinking-placeholder">正在梳理思路...</div>
    </div>
  </div>
</template>

<script setup>
import { ref } from "vue";
import { NIcon, NSpin } from "naive-ui";
import { ChevronDownOutline, ChevronForwardOutline } from "@vicons/ionicons5";

defineProps({
  content: {
    type: String,
    default: "",
  },
});

const collapsed = ref(false);
</script>

<style scoped>
.thinking-block {
  margin: 12px 0;
  border-radius: 8px;
  background-color: #f8f9fa;
  border: 1px solid #e9ecef;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.thinking-block.is-collapsed {
  opacity: 0.8;
}

.thinking-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 16px;
  cursor: pointer;
  user-select: none;
  transition: background-color 0.2s;
}

.thinking-header:hover {
  background-color: #f1f3f5;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 8px;
}

.toggle-icon {
  font-size: 14px;
  color: #adb5bd;
  transition: transform 0.3s;
}

.thinking-title {
  font-size: 13px;
  font-weight: 500;
  color: #495057;
  letter-spacing: 0.02em;
}

.thinking-status {
  display: flex;
  align-items: center;
}

.thinking-content {
  padding: 0 16px 12px;
  max-height: 400px;
  overflow-y: auto;
}

.thinking-text {
  white-space: pre-wrap;
  font-size: 13px;
  color: #6c757d;
  line-height: 1.6;
  font-family: "Fira Code", "JetBrains Mono", monospace;
  border-left: 2px solid #dee2e6;
  padding-left: 12px;
}

.thinking-placeholder {
  font-style: italic;
  color: #adb5bd;
  font-size: 12px;
  padding-left: 12px;
}

/* Scrollbar styling */
.thinking-content::-webkit-scrollbar {
  width: 4px;
}
.thinking-content::-webkit-scrollbar-track {
  background: transparent;
}
.thinking-content::-webkit-scrollbar-thumb {
  background: #dee2e6;
  border-radius: 4px;
}
.thinking-content::-webkit-scrollbar-thumb:hover {
  background: #ced4da;
}
</style>
