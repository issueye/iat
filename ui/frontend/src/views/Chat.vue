<template>
  <div class="chat-container">
    <!-- Left Sidebar: Sessions -->
    <div class="chat-sidebar">
      <div class="sidebar-header">
        <div style="display: flex; gap: 8px">
          <n-select
            v-model:value="currentProjectId"
            :options="projectOptions"
            placeholder="选择项目"
            size="small"
            style="flex: 1"
          />
          <n-button size="small" @click="refreshProjects">
            <template #icon>
              <n-icon><RefreshOutline /></n-icon>
            </template>
          </n-button>
        </div>
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
          <div class="session-info">
            <div class="session-name">{{ session.name }}</div>
            <div class="session-time">
              {{ new Date(session.createdAt).toLocaleString() }}
            </div>
          </div>
          <n-button
            v-if="currentSessionId === session.id"
            size="tiny"
            quaternary
            circle
            class="delete-session-btn"
            @click.stop="handleDeleteSession(session)"
          >
            <template #icon>
              <n-icon><TrashOutline /></n-icon>
            </template>
          </n-button>
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
          :auto-scroll="messages.length >= 2"
        >
          <template #avatar="{ item }">
            <n-avatar
              size="medium"
              round
              :color="item.role === 'user' ? '#e7f5ff' : '#f3f0ff'"
              :text-color="item.role === 'user' ? '#228be6' : '#7950f2'"
            >
              {{ item.role === "user" ? "U" : "AI" }}
            </n-avatar>
          </template>
          <template #header="{ item }">
            <ChatItemHeader :item="item" />
          </template>
          <template #content="{ item }">
            <ChatItemContent
              :item="item"
              :messages="messages"
              :taskMap="subAgentTaskMap"
            />
          </template>
          <template #footer="{ item }">
            <ChatItemFooter :item="item" />
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
          try {
            await api.createSession(
              newSessionName || '',
              currentProjectId,
              currentChatAgentId || 0,
            );
            await chatStore.fetchSessions(currentProjectId);
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
        placeholder="会话名称 (可选，留空由 AI 生成)"
        autofocus
      />
    </n-modal>
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
  InformationCircleOutline,
  StopCircleOutline,
  RefreshOutline,
} from "@vicons/ionicons5";
import { api } from "../api";
import { useAgentStore } from "../store/agent";
import { useChatStore } from "../store/chat";
import { useProjectStore } from "../store/project";
import { useWorkflowStore } from "../store/workflow";
import WorkflowCanvas from "../components/workflow/WorkflowCanvas.vue";
import ChatItemHeader from "./components/ChatItemHeader.vue";
import ChatItemFooter from "./components/ChatItemFooter.vue";
import ChatItemContent from "./components/ChatItemContent.vue";

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

// State from stores
const projects = computed(() => projectStore.projects);
const sessions = computed(() => chatStore.sessions);
const agents = computed(() => agentStore.agents);
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

const modeOptions = computed(() => {
  return agentStore.modes.map((m) => ({
    label: `${m.name} (${m.key})`,
    value: m.key,
    id: m.id,
  }));
});

// Filter agents by selected mode
const agentOptions = computed(() => {
  const mode = agentStore.modes.find((m) => m.key === currentChatMode.value);
  if (!mode) return [];

  return agents.value
    .filter((a) => (a.modes || []).some((m) => m.id === mode.id))
    .map((a) => ({
      label: a.name,
      value: a.id,
    }));
});

// Watch mode change to auto-select first available agent
watch(currentChatMode, (newMode) => {
  const filtered = agentOptions.value;
  if (filtered.length > 0) {
    // If current agent not in filtered list, pick first one
    if (!filtered.find((a) => a.value === currentChatAgentId.value)) {
      currentChatAgentId.value = filtered[0].value;
    }
  } else {
    currentChatAgentId.value = null;
  }
});

// Watch agent change to sync mode (if current agent doesn't support current mode)
watch(currentChatAgentId, (newAgentId) => {
  if (!newAgentId) return;
  const agent = agents.value.find((a) => a.id === newAgentId);
  if (agent && agent.modes && agent.modes.length > 0) {
    const supportsCurrentMode = agent.modes.some(
      (m) => m.key === currentChatMode.value,
    );
    if (!supportsCurrentMode) {
      // Switch to first supported mode of this agent
      currentChatMode.value = agent.modes[0].key;
    }
  }
});

// UI State
const showCreateModal = ref(false);
const newSessionName = ref("");
const sessionSearchQuery = ref("");
const searchedSessions = ref([]);

const projectOptions = computed(() => {
  return projects.value.map((p) => ({
    label: p.name,
    value: p.id,
  }));
});

const isSearchingSessions = computed(() => {
  return String(sessionSearchQuery.value || "").trim() !== "";
});

const displaySessions = computed(() => {
  return isSearchingSessions.value ? searchedSessions.value : sessions.value;
});

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
async function refreshProjects() {
  try {
    await projectStore.fetchProjects();
    message.success("项目列表已刷新");
  } catch (e) {
    message.error("刷新项目失败: " + e.message);
  }
}

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

async function handleDeleteSession(session) {
  dialog.warning({
    title: "删除会话",
    content: `确定要删除会话 "${session.name}" 吗？此操作无法撤销。`,
    positiveText: "删除",
    negativeText: "取消",
    onPositiveClick: async () => {
      try {
        await api.deleteSession(session.id);
        message.success("会话已删除");
        if (currentSessionId.value === session.id) {
          currentSessionId.value = null;
          chatStore.messages = [];
        }
        await chatStore.fetchSessions(currentProjectId.value);
      } catch (e) {
        message.error("删除失败: " + e.message);
      }
    },
  });
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

  // Auto-create session if not selected
  if (!currentSessionId.value) {
    try {
      const sess = await api.createSession(
        "", // Empty name
        currentProjectId.value,
        currentChatAgentId.value,
      );
      await chatStore.fetchSessions(currentProjectId.value);
      currentSessionId.value = sess.id;
    } catch (e) {
      message.error("创建会话失败: " + e.message);
      return;
    }
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
    if (chatStore.messages[aiMsgIndex]) {
      chatStore.messages[aiMsgIndex].content = "[Error: " + e.message + "]";
    }
    generationStatus.value = ThinkingStatuses.Error;
  } finally {
    chatStore.setStreaming(false);
    generationStatus.value = ThinkingStatuses.End;
  }
}

function handleSSEEvent(data, aiMsgIndex) {
  const msg = chatStore.messages[aiMsgIndex];
  if (!msg) return;

  if (data.type === "chunk") {
    msg.content += data.content || "";
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
:deep(.el-bubble-content-wrapper .el-bubble-footer) {
  margin-top: var(--base-gap-sm);
}

:deep(.el-bubble-content-wrapper .el-bubble-content-filled) {
  padding: 0px;
  width: 100%;
}

.chat-container {
  display: flex;
  height: 100%;
  background-color: var(--color-grey-light);
  overflow: hidden;
}

/* Sidebar Styling */
.chat-sidebar {
  width: var(--sidebar-width);
  background-color: var(--color-light);
  border-right: 1px solid var(--color-grey-light);
  display: flex;
  flex-shrink: 0;
  flex-direction: column;
  transition: all 0.3s;
  box-shadow: 2px 0 8px rgba(0, 0, 0, 0.02);
}

.sidebar-header {
  padding: var(--base-padding-sm) var(--base-padding);
  border-bottom: 1px solid var(--color-grey-light);
}

.session-search {
  padding: var(--base-padding-sm) var(--base-padding);
}

.session-list {
  flex: 1;
  overflow-y: auto;
  padding: var(--base-padding);
}

.session-list::-webkit-scrollbar {
  width: 4px;
}

.session-list::-webkit-scrollbar-thumb {
  background: #dee2e6;
  border-radius: 4px;
}

.session-item {
  padding: var(--base-padding-sm) var(--base-padding);
  cursor: pointer;
  border-radius: var(--base-radius);
  margin-bottom: var(--base-gap-sm);
  transition: all 0.2s;
  border: 1px solid transparent;
}

.session-item:hover {
  background-color: var(--color-grey-light);
}

.session-item.active {
  background-color: #e7f5ff;
  border-color: #a5d8ff;
}

.session-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.session-info {
  flex: 1;
  min-width: 0;
}

.delete-session-btn {
  opacity: 0;
  transition: opacity 0.2s;
  margin-left: 8px;
}

.session-item:hover .delete-session-btn {
  opacity: 1;
}

.session-name {
  font-size: 14px;
  font-weight: 500;
  color: var(--color-text);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.session-time {
  font-size: 11px;
  color: var(--color-grey-text);
  margin-top: 4px;
}

.empty-sessions {
  text-align: center;
  color: #adb5bd;
  padding: 40px 0;
  font-size: 13px;
}

/* Main Content Styling */
.chat-main {
  flex: calc(100% - var(--sidebar-width));
  display: flex;
  flex-direction: column;
  background-color: #ffffff;
}

.chat-header-bar {
  padding: var(--base-padding);
  width: calc(100% - 27px);
  display: flex;
  align-items: center;
  justify-content: space-between;
  border-bottom: 1px solid #e9ecef;
  background-color: rgba(255, 255, 255, 0.8);
  backdrop-filter: blur(8px);
  z-index: 10;
}

.header-controls {
  display: flex;
  align-items: center;
  gap: 12px;
}

.chat-main-content {
  width: calc(100% - 1px);
  height: calc(100% - 45px);
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.messages-area {
  flex: 1;
  padding: var(--base-padding);
  overflow-y: auto;
}

.input-area {
  padding: var(--base-padding);
  background: linear-gradient(to top, #ffffff 80%, rgba(255, 255, 255, 0));
}

/* Bubble Overrides */
:deep(.n-avatar) {
  background-color: #e7f5ff;
  color: #228be6;
  font-weight: bold;
}

:deep(.bubble-list-item) {
  margin-bottom: 24px !important;
}

/* Sender Override */
:deep(.n-input) {
  border-radius: 12px;
}
</style>
