<template>
  <div v-if="item.role === 'tool'" class="tool-call-bubble">
    <div class="tool-header">
      <n-icon><ContractOutline /></n-icon>
      <span>工具调用: {{ item.toolName }}</span>
    </div>
    <pre class="tool-args">{{ item.toolArguments }}</pre>
  </div>
  <div v-else>
    <Thinking
      v-if="
        parseThinkContent(item.content).think ||
        parseThinkContent(item.content).isThinkingOpen
      "
      :content="parseThinkContent(item.content).think"
    />
    <XMarkdown
      v-if="parseThinkContent(item.content).answer"
      :markdown="parseThinkContent(item.content).answer"
      default-theme-mode="light"
      style="text-align: left; margin-top: 8px"
      :code-x-props="{ enableCodeLineNumber: true }"
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
import SubAgentCard from "@/components/SubAgentCard.vue";
import Thinking from "@/components/Thinking.vue";
import { ContractOutline } from "@vicons/ionicons5";
import { ThinkTags } from "@/constants/chat";

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

// Helper Functions
function parseThinkContent(text) {
  const raw = String(text || "");
  const thinkOpenTag = ThinkTags.Open;
  const thinkCloseTag = ThinkTags.Close;
  let i = 0;
  let inThink = false;
  let answer = "";
  let think = "";

  while (i < raw.length) {
    const openAt = raw.indexOf(thinkOpenTag, i);
    const closeAt = raw.indexOf(thinkCloseTag, i);

    const nextAt =
      openAt === -1
        ? closeAt
        : closeAt === -1
          ? openAt
          : Math.min(openAt, closeAt);

    if (nextAt === -1) {
      const chunk = raw.slice(i);
      if (inThink) think += chunk;
      else answer += chunk;
      break;
    }

    const chunk = raw.slice(i, nextAt);
    if (inThink) think += chunk;
    else answer += chunk;

    if (nextAt === openAt) {
      inThink = true;
      i = nextAt + thinkOpenTag.length;
    } else {
      inThink = false;
      i = nextAt + thinkCloseTag.length;
    }
  }

  return {
    think: think.trim(),
    answer: answer.trim(),
    isThinkingOpen: inThink,
  };
}

function getTasksByMessage(msg) {
  const idx = props.messages.indexOf(msg);
  if (idx === -1) return [];
  return Array.from(props.taskMap.values()).filter(
    (t) => t.messageIndex === idx && !t.parentTaskId,
  );
}
</script>
<style scoped>
.sub-agent-tasks-container {
  margin-top: 16px;
}

.tool-call-bubble {
  background-color: #f8f9fa;
  border-radius: 12px;
  padding: 12px 16px;
  border-left: 4px solid #339af0;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.02);
}

.tool-header {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 600;
  font-size: 13px;
  color: #495057;
  margin-bottom: 8px;
}

.tool-args {
  margin: 0;
  font-size: 12px;
  color: #495057;
  background: #ffffff;
  padding: 10px;
  border-radius: 8px;
  border: 1px solid #e9ecef;
  overflow-x: auto;
  font-family: "Fira Code", monospace;
}
</style>
