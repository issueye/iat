<template>
  <div class="thinking-block">
    <div class="thinking-header" @click="collapsed = !collapsed">
      <n-icon
        :component="collapsed ? ChevronForwardOutline : ChevronDownOutline"
      />
      <span class="thinking-title">思考过程</span>
    </div>
    <div v-show="!collapsed" class="thinking-content">
      <div class="thinking-text">{{ content }}</div>
      <div v-if="!content" class="thinking-loading">思考中...</div>
    </div>
  </div>
</template>

<script setup>
import { ref } from "vue";
import { NIcon } from "naive-ui";
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
  margin: 8px 0;
  border-radius: 8px;
  background-color: #f5f5f5;
  border: 1px solid #e0e0e0;
  overflow: hidden;
}

.thinking-header {
  display: flex;
  align-items: center;
  padding: 8px 12px;
  cursor: pointer;
  font-size: 12px;
  color: #666;
  user-select: none;
  background-color: #fafafa;
}

.thinking-header:hover {
  background-color: #f0f0f0;
}

.thinking-title {
  margin-left: 4px;
  font-weight: 600;
}

.thinking-content {
  padding: 8px 12px;
  border-top: 1px solid #e0e0e0;
  background-color: #fff;
}

.thinking-text {
  white-space: pre-wrap;
  font-size: 13px;
  color: #666;
  line-height: 1.5;
  font-family: monospace;
}

.thinking-loading {
  font-style: italic;
  color: #999;
  font-size: 12px;
}
</style>
