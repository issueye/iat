<template>
  <div v-if="item.role === 'tool'" class="tool-call-bubble">
    <div class="tool-header">
      <n-icon><ContractOutline /></n-icon>
      <span>{{ item.toolName }}</span>
    </div>
    <pre class="tool-args">{{ item.toolArguments }}</pre>
  </div>
  <div v-else>
    <template v-for="(segment, index) in parsedContent.segments" :key="index">
      <Thinking
        v-if="segment.type === 'think'"
        :content="segment.content"
        :loading="parsedContent.isThinkingOpen && index === parsedContent.segments.length - 1"
        :default-collapsed="parsedContent.hasAnswer"
      />
      <XMarkdown
        v-else-if="segment.type === 'text'"
        :markdown="segment.content"
        default-theme-mode="light"
        style="text-align: left; margin-top: 8px"
        :code-x-props="{ enableCodeLineNumber: true }"
      />
    </template>
    
    <!-- Loading state when thinking is open but no content in the last segment yet -->
    <Thinking
      v-if="parsedContent.isThinkingOpen && (parsedContent.segments.length === 0 || parsedContent.segments[parsedContent.segments.length-1].type !== 'think')"
      content=""
      loading
      :default-collapsed="parsedContent.hasAnswer"
    />

    <!-- Sub-Agent Tasks -->
    <div
      v-if="getTasksByMessage(item).length > 0"
      class="sub-agent-tasks-container"
    >
      <SubAgentCard
        v-for="task in getTasksByMessage(item)"
        :key="task.taskId"
        v-bind="task"
        @abort="handleAbortSubAgent"
      />
    </div>
  </div>
</template>
<script setup>
import { computed } from "vue";
import SubAgentCard from "@/components/SubAgentCard.vue";
import Thinking from "@/components/Thinking.vue";
import { ContractOutline } from "@vicons/ionicons5";
import { parseThinkContent } from "@/utils/think";
import { api } from "@/api";
import { useMessage } from "naive-ui";

const message = useMessage();
const props = defineProps({
  item: {
    type: Object,
    default: () => ({}),
  },
  messages: {
    type: Array,
    default: () => [],
  },
  taskMap: {
    type: Object,
    default: () => ({}),
  },
});

const parsedContent = computed(() => parseThinkContent(props.item.content));

// Helper Functions
function getTasksByMessage(msg) {
  const idx = props.messages.indexOf(msg);
  if (idx === -1) return [];
  return Array.from(props.taskMap.values()).filter(
    (t) => t.messageIndex === idx && !t.parentTaskId,
  );
}

async function handleAbortSubAgent(taskId) {
  try {
    await api.abortSubAgentTask(taskId);
    message.success("子任务中止请求已发送");
  } catch (e) {
    message.error("中止失败: " + e.message);
  }
}
</script>
<style scoped>
.sub-agent-tasks-container {
  margin-top: var(--base-gap-md);
}

.tool-call-bubble {
  background-color: var(--color-grey-light);
  border-radius: var(--shiki-custom-brr);
  padding: var(--base-padding-sm) var(--base-padding-md);
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.02);
}

.tool-header {
  display: flex;
  align-items: center;
  gap: var(--base-gap-sm);
  font-weight: 600;
  font-size: var(--base-font-size-md);
  color: var(--color-text);
  margin-bottom: var(--base-gap-sm);
}

.tool-args {
  margin: 0;
  font-size: var(--base-font-size-sm);
  color: var(--color-text);
  background: var(--color-grey-light);
  padding: var(--base-padding-sm);
  border-radius: var(--base-radius);
  border: 1px solid var(--color-grey);
  overflow-x: auto;
  font-family: "Fira Code", monospace;
}
</style>
