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

      <div class="messages-area">
        <div v-for="(msg, index) in messages" :key="index" class="message-item">
          <div class="message-role">
            <n-tag
              size="small"
              :type="msg.role === 'user' ? 'primary' : 'success'"
            >
              {{ msg.role }}
            </n-tag>
            <span class="message-time">{{
              new Date(msg.createdAt).toLocaleTimeString()
            }}</span>
          </div>
          <div class="message-content">
            <pre class="content-pre">{{ msg.content }}</pre>

            <!-- Tool Calls -->
            <div v-if="msg.role === 'tool'" class="tool-call">
              <div class="tool-header">Tool: {{ msg.toolName }}</div>
              <pre class="tool-args">{{ msg.toolArguments }}</pre>
            </div>
          </div>
        </div>
      </div>

      <div class="input-area">
        <n-input
          v-model:value="inputText"
          type="textarea"
          :autosize="{ minRows: 2, maxRows: 6 }"
          placeholder="输入消息... (Ctrl+Enter 发送)"
          @keydown.ctrl.enter="handleSend(inputText)"
        />
        <n-button
          type="primary"
          class="send-btn"
          @click="handleSend(inputText)"
          :loading="isGenerating"
          :disabled="!inputText.trim() && !isGenerating"
        >
          <template #icon>
            <n-icon><ChevronForwardOutline /></n-icon>
          </template>
          发送
        </n-button>
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
  </div>
</template>

<script setup>
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

// State
const projects = ref([]);
const sessions = ref([]);
const agents = ref([]);
const currentProjectId = ref(null);
const currentSessionId = ref(null);
const currentChatMode = ref(ChatModes.Chat);
const currentChatAgentId = ref(null);
const messages = ref([]);
const inputText = ref("");
const isGenerating = ref(false);
const generationStatus = ref(ThinkingStatuses.End);

// Computed properties
const lastAssistantMessage = computed(() => {
  for (let i = messages.value.length - 1; i >= 0; i--) {
    if (messages.value[i]?.role === ChatRoles.Assistant)
      return messages.value[i];
  }
  return null;
});

const modeOptions = [
  { label: "对话 (Chat)", value: ChatModes.Chat },
  { label: "计划 (Plan)", value: ChatModes.Plan },
  { label: "构建 (Build)", value: ChatModes.Build },
];

const totalTokenUsage = computed(() => {
  return messages.value.reduce((sum, m) => {
    const n = Number(m?.tokenUsage || 0);
    return sum + (Number.isFinite(n) ? n : 0);
  }, 0);
});

// UI State
const showCreateModal = ref(false);
const newSessionName = ref("");
const selectedMode = ref(ChatModes.Chat);
const selectedAgentId = ref(null);
const showDetailModal = ref(false);
const showPromptModal = ref(false);
const currentViewPrompt = ref("");
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
    loadSessions(newVal);
    currentSessionId.value = null;
    messages.value = [];
  }
});

watch(currentSessionId, async (newVal) => {
  if (newVal) {
    try {
      const msgs = await api.getSessionMessages(newVal);
      // Convert backend message format to UI format
      messages.value = (msgs || []).map((m) => ({
        role: m.role,
        content: m.content,
        createdAt: m.createdAt,
        // Handle tool calls if needed
      }));
    } catch (e) {
      message.error("加载消息失败: " + e.message);
    }
  } else {
    messages.value = [];
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
        await api.clearSessionMessages(currentSessionId.value);
        messages.value = [];
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
    isGenerating.value = false;
    generationStatus.value = ThinkingStatuses.Cancel;
  } catch (e) {
    message.error("停止失败: " + e.message);
  }
}

async function loadProjects() {
  try {
    const data = await api.listProjects();
    projects.value = data || [];
    if (!currentProjectId.value && projects.value.length > 0) {
      currentProjectId.value = projects.value[0].id;
    }
  } catch (e) {
    message.error("加载项目失败: " + e.message);
  }
}

async function loadAgents() {
  try {
    const data = await api.listAgents();
    agents.value = data || [];
  } catch (e) {
    message.error("加载智能体失败: " + e.message);
  }
}

async function loadSessions(projectId) {
  if (!projectId) return;
  try {
    const data = await api.listSessions(projectId);
    sessions.value = data || [];
  } catch (e) {
    message.error("加载会话失败: " + e.message);
  }
}

// Chat Logic
async function handleSend(content) {
  if (isGenerating.value) return;
  if (!content.trim()) return;

  inputText.value = "";
  isGenerating.value = true;
  generationStatus.value = ThinkingStatuses.Start;

  // Add User Message
  messages.value.push({
    role: ChatRoles.User,
    content: content,
    createdAt: new Date().toISOString(),
  });

  // Add Assistant Placeholder
  const aiMsgIndex =
    messages.value.push({
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

    while (true) {
      const { done, value } = await reader.read();
      if (done) break;

      const chunk = decoder.decode(value);
      const lines = chunk.split("\n\n");

      for (const line of lines) {
        if (line.startsWith("data: ")) {
          const dataStr = line.slice(6);
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
    messages.value[aiMsgIndex].content = "[Error: " + e.message + "]";
    generationStatus.value = ThinkingStatuses.Error;
  } finally {
    isGenerating.value = false;
    generationStatus.value = ThinkingStatuses.End;
  }
}

function handleSSEEvent(data, aiMsgIndex) {
  if (data.type === "chunk") {
    messages.value[aiMsgIndex].content += data.content;
  } else if (data.type === "tool_call") {
    // Handle tool call UI
    messages.value.push({
      role: ChatRoles.Tool,
      toolName: data.extra.name,
      toolArguments: data.extra.arguments,
      content: "",
      collapsed: true,
    });
  } else if (data.type === "error") {
    message.error(data.content);
    generationStatus.value = ThinkingStatuses.Error;
  }
}

// Lifecycle
onMounted(() => {
  loadProjects();
  loadAgents();
});
</script>

<style scoped>
.chat-container {
  display: flex;
  height: calc(100vh - 40px); /* Adjust based on layout */
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
}

.chat-header-bar {
  display: flex;
  justify-content: flex-end;
  padding-bottom: 10px;
  border-bottom: 1px solid #eee;
}

.header-controls {
  display: flex;
  gap: 10px;
}

.messages-area {
  flex: 1;
  overflow-y: auto;
  padding: 20px;
  background-color: #fafafa;
  margin-bottom: 10px;
  border-radius: 4px;
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
  display: flex;
  gap: 10px;
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
</style>
