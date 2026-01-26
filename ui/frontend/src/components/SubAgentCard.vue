<script setup>
import { computed } from "vue";
import {
  NCard,
  NTag,
  NButton,
  NIcon,
  NCollapse,
  NCollapseItem,
  NSpace,
  NSpin,
} from "naive-ui";
import {
  CloseCircleOutline,
  CheckmarkCircleOutline,
  SyncOutline,
  HourglassOutline,
  BanOutline,
  ChevronDownOutline,
} from "@vicons/ionicons5";
import Thinking from "./Thinking.vue";

import XMarkdown from "./renderers/XMarkdown.vue";

const ThinkTags = {
  Open: "<think>",
  Close: "</think>",
};

const props = defineProps({
  taskId: String,
  agentName: String,
  query: String,
  status: String,
  depth: Number,
  result: String,
  error: String,
  chunks: {
    type: Array,
    default: () => [],
  },
  children: {
    type: Array,
    default: () => [],
  },
});

const emit = defineEmits(["abort"]);

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

const parsedStream = computed(() => {
  const chunks = props.chunks || [];
  return parseThinkContent(chunks.join(""));
});

const statusColor = computed(() => {
  switch (props.status) {
    case "pending":
      return "#faad14";
    case "running":
      return "#1890ff";
    case "completed":
      return "#52c41a";
    case "failed":
      return "#f5222d";
    case "aborted":
      return "#8c8c8c";
    default:
      return "#d9d9d9";
  }
});

const statusIcon = computed(() => {
  switch (props.status) {
    case "pending":
      return HourglassOutline;
    case "running":
      return SyncOutline;
    case "completed":
      return CheckmarkCircleOutline;
    case "failed":
      return CloseCircleOutline;
    case "aborted":
      return BanOutline;
    default:
      return HourglassOutline;
  }
});

const statusLabel = computed(() => {
  switch (props.status) {
    case "pending":
      return "等待中";
    case "running":
      return "运行中";
    case "completed":
      return "已完成";
    case "failed":
      return "失败";
    case "aborted":
      return "已中止";
    default:
      return "未知";
  }
});

const fullThought = computed(() => {
  const chunks = props.chunks || [];
  return chunks.join("");
});

const handleAbort = () => {
  emit("abort", props.taskId);
};
</script>

<template>
  <div
    class="sub-agent-card-wrapper"
    :style="{ marginLeft: depth > 0 ? '16px' : '0' }"
  >
    <n-card
      size="small"
      :style="{ borderLeft: `4px solid ${statusColor}` }"
      class="sub-agent-card"
    >
      <template #header>
        <n-space align="center" justify="space-between">
          <n-space align="center" :size="8">
            <n-icon
              :color="statusColor"
              :component="statusIcon"
              :class="{ 'spin-animation': status === 'running' }"
            />
            <span class="agent-name">{{ agentName }}</span>
            <n-tag
              :color="{ textColor: statusColor, borderColor: statusColor }"
              size="tiny"
              round
              :bordered="false"
            >
              层级 {{ depth }}
            </n-tag>
          </n-space>

          <n-button
            v-if="status === 'running' || status === 'pending'"
            size="tiny"
            type="error"
            secondary
            @click="handleAbort"
          >
            中止
          </n-button>
        </n-space>
      </template>

      <div class="task-query"><strong>任务:</strong> {{ query }}</div>

      <n-collapse arrow-placement="right" class="task-details">
        <n-collapse-item title="过程与思考" name="thought">
          <template #header-extra>
            <span v-if="status === 'running'" class="running-text"
              >生成中...</span
            >
          </template>
          <div class="thought-content">
            <Thinking
              v-if="parsedStream.think || parsedStream.isThinkingOpen"
              :content="parsedStream.think"
            />
            <div
              v-if="parsedStream.answer"
              class="stream-answer"
              style="margin-top: 8px"
            >
              <XMarkdown :markdown="parsedStream.answer" />
            </div>
            <div
              v-if="
                !parsedStream.think &&
                !parsedStream.isThinkingOpen &&
                !parsedStream.answer
              "
              class="empty-text"
            >
              暂无处理过程
            </div>
          </div>
        </n-collapse-item>

        <n-collapse-item
          v-if="result || error"
          :title="status === 'failed' ? '错误详情' : '最终结果'"
          name="result"
        >
          <div :class="['result-content', status]">
            <XMarkdown v-if="result" :markdown="result" />
            <div v-else-if="error" class="error-text">{{ error }}</div>
          </div>
        </n-collapse-item>
      </n-collapse>

      <!-- 递归渲染子任务 -->
      <div v-if="children && children.length > 0" class="nested-tasks">
        <SubAgentCard
          v-for="child in children"
          :key="child.taskId"
          v-bind="child"
          @abort="(id) => $emit('abort', id)"
        />
      </div>
    </n-card>
  </div>
</template>

<style scoped>
.sub-agent-card-wrapper {
  margin-top: 16px;
  margin-bottom: 16px;
}

.sub-agent-card {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.05);
  background-color: #ffffff;
  border-radius: 12px !important;
  transition: transform 0.2s;
}

.sub-agent-card:hover {
  transform: translateY(-2px);
}

.agent-name {
  font-weight: 600;
  font-size: 15px;
  color: #212529;
}

.task-query {
  font-size: 13px;
  color: #495057;
  margin-bottom: 12px;
  padding: 10px 14px;
  background: #f8f9fa;
  border-radius: 8px;
  line-height: 1.5;
}

.task-details {
  margin-top: 12px;
}

.thought-content,
.result-content {
  font-size: 13px;
  line-height: 1.6;
  color: #495057;
}

.empty-text {
  color: #adb5bd;
  font-style: italic;
  font-size: 12px;
}

.error-text {
  color: #fa5252;
  white-space: pre-wrap;
  background: #fff5f5;
  padding: 12px;
  border-radius: 8px;
  border: 1px solid #ffe3e3;
}

.running-text {
  font-size: 12px;
  color: #228be6;
  font-weight: 500;
}

.nested-tasks {
  margin-top: 16px;
  border-top: 1px dashed #e9ecef;
  padding-top: 16px;
}

.spin-animation {
  animation: spin 2s linear infinite;
}

@keyframes spin {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}

:deep(.n-card-header) {
  padding: 12px 16px !important;
  background-color: #f8f9fa;
  border-bottom: 1px solid #f1f3f5;
}

:deep(.n-card__content) {
  padding: 16px !important;
}

:deep(.n-collapse-item__header) {
  font-size: 13px !important;
  font-weight: 500 !important;
}
</style>
