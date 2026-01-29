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
    <div v-if="!content && loading && !collapsed" class="thinking-status">
        <n-spin size="small" />
      </div>
    </div>
    <div v-show="!collapsed" class="thinking-content">
      <div v-if="content" class="thinking-text">{{ content }}</div>
      <div v-else-if="loading" class="thinking-placeholder">正在梳理思路...</div>
      <div v-else class="thinking-placeholder">无思考过程记录</div>
    </div>
  </div>
</template>

<script setup>
import { ref, watch } from "vue";
import { NIcon, NSpin } from "naive-ui";
import { ChevronDownOutline, ChevronForwardOutline } from "@vicons/ionicons5";

const props = defineProps({
  content: {
    type: String,
    default: "",
  },
  loading: {
    type: Boolean,
    default: false,
  },
  defaultCollapsed: {
    type: Boolean,
    default: false,
  }
});

const collapsed = ref(props.defaultCollapsed);

// Sync with parent control but allow internal toggle
watch(() => props.defaultCollapsed, (newVal) => {
  collapsed.value = newVal;
});
</script>

<style scoped>
.thinking-block {
  margin: var(--base-gap-sm) 0;
  border-radius: var(--base-radius);
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  border: 1px solid #e9ecef;
  background-color: #f8f9fa;
  overflow: hidden;
}

.thinking-block.is-collapsed {
  border-color: #dee2e6;
  background-color: #f1f3f5;
}

.thinking-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 6px 12px;
  cursor: pointer;
  user-select: none;
  transition: all 0.2s;
  background-color: #f1f3f5;
  border-bottom: 1px solid transparent;
}

.thinking-block.is-collapsed .thinking-header {
  background-color: transparent;
}

.thinking-header:hover {
  background-color: #e9ecef;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 8px;
}

.toggle-icon {
  font-size: 14px;
  color: #868e96;
  transition: transform 0.3s;
}

.thinking-title {
  font-size: 12px;
  font-weight: 600;
  color: #495057;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.thinking-status {
  display: flex;
  align-items: center;
}

.thinking-content {
  padding: 12px;
  max-height: 300px;
  overflow-y: auto;
  background-color: #f8f9fa;
  border-top: 1px solid #e9ecef;
}

.thinking-text {
  white-space: pre-wrap;
  font-size: 13px;
  color: #495057;
  line-height: 1.6;
  font-family: "Fira Code", "JetBrains Mono", monospace;
  padding-left: 4px;
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
