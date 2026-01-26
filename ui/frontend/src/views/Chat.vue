<template>
  <div class="chat-container">
    <!-- Left Sidebar: Sessions -->
    <div class="chat-sidebar">
      <div class="sidebar-header">
        <n-select
          v-model:value="currentProjectId"
          :options="projectOptions"
          placeholder="选择项目"
          size="small"
        />
        <n-button
          block
          dashed
          style="margin-top: 8px"
          @click="showCreateModal = true"
          :disabled="!currentProjectId"
        >
          <template #icon>
            <n-icon><InformationCircleOutline /></n-icon>
          </template>
          新建会话
        </n-button>
      </div>

      <div class="session-search">
        <n-input
          v-model:value="sessionSearchQuery"
          placeholder="搜索会话..."
          size="small"
          clearable
        >
          <template #prefix>
            <n-icon><InformationCircleOutline /></n-icon>
          </template>
        </n-input>
      </div>

      <div class="session-list">
        <div
          v-for="session in displaySessions"
          :key="session.id"
          class="session-item"
          :class="{ active: currentSessionId === session.id }"
          @click="currentSessionId = session.id"
        >
          <div class="session-name">{{ session.name }}</div>
          <div class="session-time">
            {{ new Date(session.createdAt).toLocaleString() }}
          </div>
        </div>
        <div v-if="displaySessions.length === 0" class="empty-sessions">
          无会话
        </div>
      </div>
    </div>

    <!-- Right Main: Chat -->
    <div class="chat-main">
      <div class="chat-header-bar">
        <div class="header-controls">
          <n-select
            v-model:value="currentChatAgentId"
            :options="agentOptions"
            placeholder="选择智能体"
            style="width: 200px"
            size="small"
            clearable
          />
          <n-select
            v-model:value="currentChatMode"
            :options="modeOptions"
            placeholder="模式"
            style="width: 120px"
            size="small"
          />
          <n-button
            size="small"
            type="error"
            secondary
            @click="handleClear"
            :disabled="!currentSessionId"
          >
            <template #icon>
              <n-icon><TrashOutline /></n-icon>
            </template>
            清空上下文
          </n-button>
          <n-button
            size="small"
            type="warning"
            secondary
            @click="handleStop"
            :disabled="!isGenerating"
          >
            <template #icon>
              <n-icon><StopCircleOutline /></n-icon>
            </template>
            停止
          </n-button>
        </div>
      </div>

      <div class="chat-main-content">
        <WorkflowCanvas
          v-if="currentWorkflowTasks.length > 0"
          :tasks="currentWorkflowTasks"
          style="margin-bottom: 10px"
        />
        <BubbleList
          :list="messages"
          :loading="isGenerating"
          class="messages-area"
          :bubble-props="{ showTime: true }"
        >
          <template #avatar="{ item }">
            <n-avatar size="medium">
              {{ item.role === "用户" ? "user" : "智能体" }}
            </n-avatar>
          </template>
          <template #header="{ item }">
            <div class="message-header">
              <span>{{ item.role === "user" ? "用户" : "智能体" }}</span>
              <span class="message-time">{{ formatTime(item.createdAt) }}</span>
            </div>
          </template>
          <template #content="{ item }">
            <div v-if="item.type === 'tool'" class="tool-call-bubble">
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
          <template #footer="{ item }">
            <div class="message-footer">
              <n-button text color="#8a2be2" @click="showPromptModalFn(item)">
                {{ item.type === "tool" ? "输出" : "输入" }}
              </n-button>
              <n-button
                text
                size="tiny"
                style="margin-left: 8px"
                @click="showDebugDrawer = true"
              >
                调试
              </n-button>

              <!-- token -->
              <span class="token-usage" v-if="item.type !== 'tool'">
                TOKEN {{ item.tokenUsage }}
              </span>
            </div>
          </template>
        </BubbleList>

        <div class="input-area">
          <Sender
            v-model="inputText"
            :disabled="isGenerating"
            :loading="isGenerating"
            placeholder="输入消息... (Ctrl+Enter 发送)"
            @submit="handleSend"
            @cancel="handleStop"
          />
        </div>
      </div>
    </div>

    <!-- Create Session Modal -->
    <n-modal
      v-model:show="showCreateModal"
      preset="dialog"
      title="新建会话"
      positive-text="创建"
      negative-text="取消"
      @positive-click="
        async () => {
          if (!newSessionName) return;
          try {
            await api.createSession(
              newSessionName,
              currentProjectId,
              currentChatAgentId || 0,
            );
            await loadSessions(currentProjectId);
            showCreateModal = false;
            newSessionName = '';
          } catch (e) {
            message.error(e.message);
          }
        }
      "
    >
      <n-input
        v-model:value="newSessionName"
        placeholder="会话名称"
        autofocus
      />
    </n-modal>
    <!-- Prompt Modal -->
    <n-modal
      v-model:show="showPromptModal"
      preset="dialog"
      :title="
        viewType === 'diff'
          ? '文件差异'
          : viewType === 'tree'
            ? '文件目录'
            : '详细信息'
      "
      style="width: 80%"
    >
      <div style="max-height: 80vh; overflow: auto">
        <ResultRenderer
          :content="currentViewPrompt"
          :type="viewType"
          :metadata="viewMetadata"
        />
      </div>
    </n-modal>

    <!-- Debug Console Drawer -->
    <n-drawer v-model:show="showDebugDrawer" :width="500" placement="right">
      <n-drawer-content title="系统调试日志" closable>
        <template #header-extra>
          <n-button size="tiny" secondary @click="debugStore.clear()"
            >清空</n-button
          >
        </template>
        <div class="debug-logs">
          <div v-for="log in logs" :key="log.id" class="log-item">
            <div class="log-meta">
              <span class="log-time">{{ log.time }}</span>
              <n-tag size="small" :bordered="false" type="info">{{
                log.type
              }}</n-tag>
            </div>
            <pre class="log-data">{{ log.data }}</pre>
          </div>
          <div v-if="logs.length === 0" class="empty-logs">暂无实时日志</div>
        </div>
      </n-drawer-content>
    </n-drawer>
  </div>
</template>

<script setup>
defineOptions({ name: "Chat" });

import { ref, onMounted, onUnmounted, computed, watch, nextTick } from "vue";
import { useRoute, useRouter } from "vue-router";
import {
  useMessage,
  useDialog,
  NIcon,
  NInput,
  NButton,
  NSelect,
  NTag,
  NModal,
  NAvatar,
} from "naive-ui";
import {
  TrashOutline,
  ArchiveOutline,
  ContractOutline,
  InformationCircleOutline,
  ChevronForwardOutline,
  ChevronDownOutline,
  StopCircleOutline,
} from "@vicons/ionicons5";
import { api } from "../api";
import { useAgentStore } from "../store/agent";
import { useChatStore } from "../store/chat";
import { useProjectStore } from "../store/project";
import { useWorkflowStore } from "../store/workflow";
import { useDebugStore } from "../store/debug";
import SubAgentCard from "../components/SubAgentCard.vue";
import Thinking from "../components/Thinking.vue";
import WorkflowCanvas from "../components/workflow/WorkflowCanvas.vue";
import ResultRenderer from "../components/renderers/ResultRenderer.vue";
import {
  ChatModes,
  ChatRoles,
  SSE,
  ThinkTags,
  ThinkingStatuses,
  ToolStages,
} from "../constants/chat";

const route = useRoute();
const router = useRouter();
const message = useMessage();
const dialog = useDialog();

const agentStore = useAgentStore();
const chatStore = useChatStore();
const projectStore = useProjectStore();
const workflowStore = useWorkflowStore();
const debugStore = useDebugStore();

// State from stores
const projects = computed(() => projectStore.projects);
const sessions = computed(() => chatStore.sessions);
const agents = computed(() => agentStore.agents);
const logs = computed(() => debugStore.logs);
const showDebugDrawer = ref(false);
const currentProjectId = computed({
  get: () => projectStore.currentProjectId,
  set: (val) => projectStore.setCurrentProject(val),
});
const currentSessionId = computed({
  get: () => chatStore.currentSessionId,
  set: (val) => chatStore.fetchMessages(val),
});
const currentChatMode = ref(ChatModes.Chat);
const currentChatAgentId = ref(null);
const messages = computed(() => chatStore.messages);
const inputText = computed({
  get: () => chatStore.input,
  set: (val) => (chatStore.input = val),
});
const isGenerating = computed(() => chatStore.streaming);
const generationStatus = ref(ThinkingStatuses.End);

// Workflow state
const currentWorkflowTasks = computed(() => workflowStore.tasks);

// Sub-Agent Tasks State
const subAgentTaskMap = ref(new Map());

function getTasksByMessage(msg) {
  const idx = messages.value.indexOf(msg);
  if (idx === -1) return [];
  return Array.from(subAgentTaskMap.value.values()).filter(
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

// Computed properties
const lastAssistantMessage = computed(() => {
  for (let i = messages.value.length - 1; i >= 0; i--) {
    if (messages.value[i]?.role === ChatRoles.Assistant)
      return messages.value[i];
  }
  return null;
});

const formatTime = (time) => {
  return new Date(time).toLocaleTimeString();
};

const modeOptions = [
  { label: "对话 (Chat)", value: ChatModes.Chat },
  { label: "计划 (Plan)", value: ChatModes.Plan },
  { label: "构建 (Build)", value: ChatModes.Build },
];

// UI State
const showCreateModal = ref(false);
const newSessionName = ref("");
const selectedMode = ref(ChatModes.Chat);
const selectedAgentId = ref(null);
const showDetailModal = ref(false);
const showPromptModal = ref(false);
const currentViewPrompt = ref("");
const viewType = ref("text");
const viewMetadata = ref({});
const toolInvocations = ref([]);
const sessionSearchQuery = ref("");
const searchedSessions = ref([]);
const searchingSessions = ref(false);

const projectOptions = computed(() => {
  return projects.value.map((p) => ({
    label: p.name,
    value: p.id,
  }));
});

const agentOptions = computed(() => {
  return agents.value.map((a) => ({
    label: a.name,
    value: a.id,
  }));
});

const showPromptModalFn = (item) => {
  viewType.value = "text";
  viewMetadata.value = {};

  switch (item.type) {
    case "tool":
      {
        currentViewPrompt.value = item.toolOutput || "";
        const toolName = item.toolName;

        if (toolName === "list_files") {
          viewType.value = "tree";
        } else if (toolName === "diff_file") {
          viewType.value = "diff";
          viewMetadata.value = { language: "diff" };
        } else if (toolName === "read_file" || toolName === "read_file_range") {
          viewType.value = "code";
          // Try to get language from path in arguments
          try {
            const args = JSON.parse(item.toolArguments || "{}");
            const path = args.path || "";
            const ext = path.split(".").pop();
            viewMetadata.value = { path, language: ext };
          } catch (e) {}
        } else {
          // Check if it's JSON
          try {
            JSON.parse(currentViewPrompt.value);
            viewType.value = "code";
            viewMetadata.value = { language: "json" };
          } catch (e) {}
        }
      }
      break;
    case "message":
      {
        currentViewPrompt.value = item.prompt || "";
        if (currentViewPrompt.value) {
          try {
            const parsed = JSON.parse(currentViewPrompt.value);
            currentViewPrompt.value = JSON.stringify(parsed, null, 2);
            viewType.value = "code";
            viewMetadata.value = { language: "json" };
          } catch (e) {}
        }
      }
      break;
  }

  showPromptModal.value = true;
};

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

function getThinkingStatus(item) {
  if (item !== lastAssistantMessage.value) return ThinkingStatuses.End;
  if (generationStatus.value === ThinkingStatuses.Cancel)
    return ThinkingStatuses.Cancel;
  if (generationStatus.value === ThinkingStatuses.Error)
    return ThinkingStatuses.Error;
  const parsed = parseThinkContent(item?.content);
  if (isGenerating.value) {
    if (parsed.isThinkingOpen) {
      return parsed.think ? ThinkingStatuses.Thinking : ThinkingStatuses.Start;
    }
    return ThinkingStatuses.End;
  }
  return ThinkingStatuses.End;
}

const isSearchingSessions = computed(() => {
  return String(sessionSearchQuery.value || "").trim() !== "";
});

const displaySessions = computed(() => {
  return isSearchingSessions.value ? searchedSessions.value : sessions.value;
});

// Event Source
let eventSource = null;

watch(currentProjectId, (newVal) => {
  if (newVal) {
    chatStore.fetchSessions(newVal);
  }
});

watch(currentSessionId, async (newVal) => {
  if (newVal) {
    // Sync agent from session
    const sess = sessions.value.find((s) => s.id === newVal);
    if (sess && sess.agentId) {
      currentChatAgentId.value = sess.agentId;
    }
  }
});

// Methods
async function handleClear() {
  if (!currentSessionId.value) return;
  dialog.warning({
    title: "清空上下文",
    content: "确定要清空当前会话的所有消息吗？这不会删除会话本身。",
    positiveText: "确定",
    negativeText: "取消",
    onPositiveClick: async () => {
      try {
        await chatStore.clearHistory();
        message.success("上下文已清空");
      } catch (e) {
        message.error("清空失败: " + e.message);
      }
    },
  });
}

async function handleStop() {
  if (!currentSessionId.value || !isGenerating.value) return;
  try {
    await api.abortSession(currentSessionId.value);
    message.info("已请求停止生成");
    chatStore.setStreaming(false);
    generationStatus.value = ThinkingStatuses.Cancel;
  } catch (e) {
    message.error("停止失败: " + e.message);
  }
}

// Chat Logic
async function handleSend(val) {
  const content = typeof val === "string" ? val : inputText.value;

  if (isGenerating.value) return;
  if (!content || !content.trim()) return;

  if (!currentChatAgentId.value) {
    message.error("请先选择一个智能体");
    return;
  }

  chatStore.input = "";
  chatStore.setStreaming(true);
  generationStatus.value = ThinkingStatuses.Start;

  // Add User Message
  const sendData = {
    role: ChatRoles.User,
    content: content,
    createdAt: new Date().toISOString(),
  };
  chatStore.addMessage(sendData);

  // Add Assistant Placeholder
  const aiMsgIndex =
    chatStore.addMessage({
      role: ChatRoles.Assistant,
      content: "",
      createdAt: new Date().toISOString(),
    }) - 1;

  // Use Fetch for stream instead of EventSource for POST
  try {
    const response = await fetch(SSE.EventsUrl, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        sessionId: currentSessionId.value || 0, // 0 for temporary session if not selected
        message: content,
        agentId: currentChatAgentId.value || 0,
        mode: currentChatMode.value,
      }),
    });

    if (!response.ok) throw new Error(response.statusText);

    const reader = response.body.getReader();
    const decoder = new TextDecoder();
    let buffer = "";

    while (true) {
      const { done, value } = await reader.read();
      if (done) break;

      buffer += decoder.decode(value, { stream: true });
      const parts = buffer.split("\n\n");
      buffer = parts.pop() || "";

      for (const line of parts) {
        const trimmedLine = line.trim();
        if (!trimmedLine) continue;

        if (trimmedLine.startsWith("data: ")) {
          const dataStr = trimmedLine.slice(6);
          try {
            const data = JSON.parse(dataStr);
            handleSSEEvent(data, aiMsgIndex);
          } catch (e) {
            console.error("Error parsing SSE event", e);
          }
        }
      }
    }
  } catch (e) {
    message.error("发送失败: " + e.message);
    chatStore.messages[aiMsgIndex].content = "[Error: " + e.message + "]";
    generationStatus.value = ThinkingStatuses.Error;
  } finally {
    chatStore.setStreaming(false);
    generationStatus.value = ThinkingStatuses.End;
  }
}

function handleSSEEvent(data, aiMsgIndex) {
  if (data.type === "chunk") {
    chatStore.messages[aiMsgIndex].content += data.content;
  } else if (data.type === "workflow_start") {
    const { goal, tasks } = data.extra || {};
    workflowStore.tasks = (tasks || []).map((t) => ({
      ...t,
      status: "pending",
    }));
  } else if (data.type === "task_status") {
    const { taskId, status, output } = data.extra || {};
    workflowStore.updateTaskStatus(taskId, status, output);
  } else if (data.type === "tool_call") {
    const {
      stage,
      taskId,
      content,
      agentName,
      query,
      status,
      error,
      result,
      depth,
      parentTaskId,
    } = data.extra || {};

    if (
      stage === "subagent_start" ||
      data.extra?.eventType === "subagent_start"
    ) {
      const task = {
        taskId,
        agentName,
        query,
        status: status || "pending",
        depth: depth || 0,
        parentTaskId,
        chunks: [],
        children: [],
        messageIndex: aiMsgIndex, // Link to the message that triggered this task
      };
      subAgentTaskMap.value.set(taskId, task);

      if (parentTaskId) {
        const parentTask = subAgentTaskMap.value.get(parentTaskId);
        if (parentTask) {
          parentTask.children.push(task);
        }
      }
    } else if (
      stage === "subagent_chunk" ||
      data.extra?.eventType === "subagent_chunk"
    ) {
      const task = subAgentTaskMap.value.get(taskId);
      if (task) {
        task.chunks.push(content);
        task.status = "running";
      }
    } else if (
      stage === "subagent_done" ||
      data.extra?.eventType === "subagent_done"
    ) {
      const task = subAgentTaskMap.value.get(taskId);
      if (task) {
        task.status = status;
        task.result = result;
        task.error = error;
      }
    } else {
      // Standard tool call
      chatStore.addMessage({
        role: ChatRoles.Tool,
        toolName: data.extra.name,
        toolArguments: data.extra.arguments,
        content: "",
        collapsed: true,
      });
    }
  } else if (data.type === "error") {
    message.error(data.content);
    generationStatus.value = ThinkingStatuses.Error;
  }
}

// Lifecycle
onMounted(() => {
  projectStore.fetchProjects();
  agentStore.fetchAll();
  chatStore.initWS();
});
</script>

<style scoped>
.debug-logs {
  font-family: "Fira Code", monospace;
  font-size: 11px;
}
.log-item {
  margin-bottom: 12px;
  border-bottom: 1px solid #f0f0f0;
  padding-bottom: 8px;
}
.log-meta {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 4px;
}
.log-time {
  color: #999;
}
.log-data {
  margin: 0;
  white-space: pre-wrap;
  word-break: break-all;
  color: #444;
  background: #f9f9f9;
  padding: 4px;
  border-radius: 2px;
}
.empty-logs {
  text-align: center;
  color: #ccc;
  padding: 40px 0;
}

.chat-container {
  display: flex;
  height: 100%;
  background-color: #fff;
}

.chat-sidebar {
  width: 260px;
  border-right: 1px solid #eee;
  display: flex;
  flex-direction: column;
  padding: 10px;
}

.sidebar-header {
  margin-bottom: 10px;
}

.session-search {
  margin-bottom: 10px;
}

.session-list {
  flex: 1;
  overflow-y: auto;
}

.session-item {
  padding: 8px;
  cursor: pointer;
  border-radius: 4px;
  margin-bottom: 4px;
}

.session-item:hover {
  background-color: #f5f5f5;
}

.session-item.active {
  background-color: #e6f7ff;
  color: #1890ff;
}

.session-name {
  font-weight: 500;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.session-time {
  font-size: 12px;
  color: #999;
}

.chat-main {
  flex: 1;
  display: flex;
  flex-direction: column;
  padding: 10px;
  width: calc(100% - 260px);
}

.chat-header-bar {
  display: flex;
  justify-content: flex-end;
  padding-bottom: 10px;
  border-bottom: 1px solid #eee;
}

.chat-main-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  gap: 10px;
}

.header-controls {
  display: flex;
  gap: 10px;
}

.messages-area {
  flex: 1;
  overflow-y: auto;
  padding: 0; /* BubbleList handles padding */
  background-color: transparent; /* BubbleList handles background */
  margin-bottom: 0;
}

.message-item {
  margin-bottom: 16px;
}

.message-role {
  margin-bottom: 4px;
  display: flex;
  align-items: center;
  gap: 8px;
}

.message-time {
  margin-left: 8px;
  font-size: 12px;
  color: #ccc;
}

.content-pre {
  white-space: pre-wrap;
  word-break: break-word;
  background-color: #fff;
  padding: 10px;
  border-radius: 4px;
  border: 1px solid #eee;
  font-family: inherit;
  margin: 0;
}

.input-area {
  width: 100%;
  display: flex;
  gap: 10px;
}

:deep(.el-sender-wrap) {
  width: calc(100% - 100px);
}

.send-btn {
  height: auto;
}

.tool-pre {
  margin: 0;
  padding: 8px 10px;
  background: rgba(0, 0, 0, 0.05);
  border-radius: 6px;
  white-space: pre-wrap;
  word-break: break-word;
  font-family: monospace;
  font-size: 12px;
}

.tool-call-bubble {
  background-color: #f0f0f0;
  border-radius: 8px;
  padding: 8px;
  border-left: 4px solid #1890ff;
}

.message-footer {
  display: flex;
  gap: 8px;
}

.token-usage {
  font-size: 12px;
  color: #999;
}

:deep(.el-bubble-content) {
  --bubble-content-max-width: 100% !important;
}

.tool-header {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: bold;
  font-size: 12px;
  color: #666;
  margin-bottom: 4px;
}

.tool-args {
  margin: 0;
  font-size: 11px;
  color: #333;
  background: #fff;
  padding: 4px;
  border-radius: 4px;
  overflow: auto;
}
</style>
