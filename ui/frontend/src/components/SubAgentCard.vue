<script setup lang="ts">
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

interface SubAgentTask {
  taskId: string;
  agentName: string;
  query: string;
  status: "pending" | "running" | "completed" | "failed" | "aborted";
  depth: number;
  result?: string;
  error?: string;
  chunks: string[];
  children: SubAgentTask[];
}

const ThinkTags = {
  Open: "<think>",
  Close: "</think>",
};

const props = defineProps<SubAgentTask>();

const emit = defineEmits(["abort"]);

function parseThinkContent(text: string) {
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

const parsedStream = computed(() => parseThinkContent(props.chunks.join("")));

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

const fullThought = computed(() => props.chunks.join(""));

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
  margin-top: 12px;
  margin-bottom: 12px;
}

.sub-agent-card {
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.05);
  background-color: #fafafa;
}

.agent-name {
  font-weight: 600;
  font-size: 14px;
}

.task-query {
  font-size: 13px;
  color: #666;
  margin-bottom: 8px;
  padding: 4px 8px;
  background: #f0f0f0;
  border-radius: 4px;
}

.task-details {
  margin-top: 8px;
}

.thought-content,
.result-content {
  font-size: 13px;
  line-height: 1.6;
}

.empty-text {
  color: #999;
  font-style: italic;
}

.error-text {
  color: #f5222d;
  white-space: pre-wrap;
}

.running-text {
  font-size: 12px;
  color: #1890ff;
}

.nested-tasks {
  margin-top: 12px;
  border-top: 1px dashed #eee;
  padding-top: 12px;
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
  padding: 8px 12px !important;
}

:deep(.n-card__content) {
  padding: 8px 12px !important;
}
</style>
