<template>
  <n-layout
    has-sider
    style="height: 100%; padding: 20px; background-color: #f5f5f5"
  >
    <!-- Session List Sider -->
    <n-layout-sider
      bordered
      :width="260"
      content-style="padding: 12px; display: flex; flex-direction: column;"
    >
      <div style="margin-bottom: 12px">
        <n-select
          v-model:value="currentProjectId"
          filterable
          placeholder="选择项目"
          :options="projectOptions"
          @update:value="handleProjectChange"
        />
      </div>

      <div style="margin-bottom: 12px">
        <n-input
          v-model:value="sessionSearchQuery"
          clearable
          placeholder="按项目名称搜索会话"
        />
      </div>

      <n-button
        block
        dashed
        style="margin-bottom: 12px"
        @click="showCreateModal = true"
        :disabled="!currentProjectId"
      >
        + 新建会话
      </n-button>

      <n-scrollbar>
        <n-list hoverable clickable>
          <n-list-item
            v-for="session in displaySessions"
            :key="session.id"
            :class="{ 'active-session': currentSessionId === session.id }"
            @click="
              isSearchingSessions
                ? handleSelectSearchedSession(session)
                : handleSelectSession(session.id)
            "
          >
            <div class="session-item">
              <n-icon v-if="session.compressed" size="16" style="color: #18a058">
                <ArchiveOutline />
              </n-icon>
              <span class="session-name">
                {{
                  isSearchingSessions && session.projectName
                    ? `${session.projectName} / ${session.name}`
                    : session.name
                }}
              </span>
              <n-button
                size="tiny"
                text
                type="error"
                @click.stop="handleDeleteSession(session.id)"
              >
                <template #icon
                  ><n-icon><TrashOutline /></n-icon
                ></template>
              </n-button>
            </div>
          </n-list-item>
        </n-list>
        <div
          v-if="displaySessions.length === 0 && (currentProjectId || isSearchingSessions)"
          style="text-align: center; color: #666; margin-top: 20px"
        >
          {{ isSearchingSessions ? "未找到会话" : "暂无会话" }}
        </div>
      </n-scrollbar>
    </n-layout-sider>

    <!-- Chat Area -->
    <n-layout-content
      content-style="display: flex; flex-direction: column; height: 100%;"
    >
      <div v-if="!currentSessionId" class="empty-state">
        <n-empty description="选择一个项目和会话以开始聊天" />
      </div>

      <template v-else>
        <div
          style="
            display: flex;
            flex-direction: column;
            height: calc(100% - 40px);
            padding: 20px;
          "
        >
          <BubbleList
            :list="messages"
            style="flex: 1; margin-bottom: 20px; overflow: hidden"
          >
            <template #content="{ item }">
              <div v-if="item.role === 'tool'">
                <div
                  style="
                    display: flex;
                    gap: 8px;
                    align-items: center;
                    justify-content: space-between;
                  "
                >
                  <div style="display: flex; gap: 8px; align-items: center">
                    <span style="font-weight: 600">工具：{{ item.toolName }}</span>
                    <span v-if="item.toolOk === false" style="color: #d03050"
                      >失败</span
                    >
                    <span v-else-if="item.toolOk === true" style="color: #18a058"
                      >成功</span
                    >
                  </div>
                  <n-button
                    size="tiny"
                    text
                    @click="item.collapsed = !item.collapsed"
                  >
                    {{ item.collapsed ? "展开" : "收起" }}
                  </n-button>
                </div>
                <div v-if="!item.collapsed" style="margin-top: 8px">
                  <div style="font-size: 12px; color: #999; margin-bottom: 4px">
                    参数
                  </div>
                  <pre class="tool-pre">{{ item.toolArguments }}</pre>
                  <template v-if="item.toolOutput !== ''">
                    <div
                      style="
                        font-size: 12px;
                        color: #999;
                        margin: 8px 0 4px;
                      "
                    >
                      返回
                    </div>
                    <pre class="tool-pre">{{ item.toolOutput }}</pre>
                  </template>
                </div>
              </div>
              <div v-else-if="!item.isMarkdown">
                {{ item.content }}
              </div>
              <XMarkdown
                v-else
                :markdown="item.content"
                default-theme-mode="light"
                style="text-align: left"
                :code-x-props="{ enableCodeLineNumber: true }"
              />
            </template>
            <template #footer="{ item }">
              <div style="margin-top: 4px; font-size: 12px; color: #999">
                <span v-if="formatTime(item.createdAt)">{{
                  formatTime(item.createdAt)
                }}</span>
                <span v-if="item.tokenUsage">
                  {{ formatTime(item.createdAt) ? " · " : "" }}Tokens:
                  {{ item.tokenUsage }}
                </span>
                <n-button
                  v-if="item.role === 'assistant' && item.prompt"
                  size="tiny"
                  text
                  type="primary"
                  style="margin-left: 8px"
                  @click="handleViewPrompt(item.prompt)"
                >
                  查看 Prompt
                </n-button>
              </div>
            </template>
          </BubbleList>

          <div style="display: flex; align-items: center; margin-bottom: 8px">
            <n-select
              v-model:value="currentChatMode"
              :options="modeOptions"
              size="small"
              style="width: 120px"
              placeholder="模式"
            />
            <n-select
              v-model:value="currentChatAgentId"
              :options="agentOptions"
              size="small"
              style="width: 180px; margin-left: 8px"
              placeholder="选择 Agent"
              clearable
            />
            <span style="font-size: 12px; color: #999; margin-left: 8px"
              >(模式与 Agent 共同决定对话逻辑)</span
            >
            <span style="font-size: 12px; color: #999; margin-left: 12px"
              >累计 Tokens: {{ totalTokenUsage }}</span
            >
            <n-button
              size="small"
              secondary
              style="margin-left: auto"
              @click="openSessionDetail"
            >
              <template #icon
                ><n-icon><InformationCircleOutline /></n-icon
              ></template>
              详情
            </n-button>
            <n-button
              size="small"
              secondary
              style="margin-left: 8px"
              @click="handleCompressSession"
            >
              <template #icon
                ><n-icon><ContractOutline /></n-icon
              ></template>
              压缩
            </n-button>
            <n-button
              size="small"
              type="error"
              secondary
              style="margin-left: 8px"
              @click="handleClearSession"
            >
              清空会话
            </n-button>
            <n-button
              size="small"
              type="warning"
              secondary
              style="margin-left: 8px"
              :disabled="!isGenerating"
              @click="handleTerminateSession"
            >
              终止生成
            </n-button>
          </div>

          <Sender
            v-model="inputText"
            @submit="handleSend"
            :loading="isGenerating"
          />
        </div>
      </template>
    </n-layout-content>

    <!-- Create Session Modal -->
    <n-modal
      v-model:show="showCreateModal"
      preset="dialog"
      title="新建会话"
      style="width: 400px"
    >
      <n-form>
        <n-form-item label="会话名称">
          <n-input v-model:value="newSessionName" placeholder="我的会话" />
        </n-form-item>
        <n-form-item label="模式">
          <n-select v-model:value="selectedMode" :options="modeOptions" />
        </n-form-item>
        <n-form-item label="选择 Agent">
          <n-select v-model:value="selectedAgentId" :options="agentOptions" placeholder="默认不选择" clearable />
        </n-form-item>
      </n-form>
      <template #action>
        <n-button @click="showCreateModal = false">取消</n-button>
        <n-button type="primary" @click="confirmCreateSession">创建</n-button>
      </template>
    </n-modal>

    <n-modal
      v-model:show="showDetailModal"
      preset="dialog"
      title="会话详情"
      :block-scroll="true"
      :style="{ width: '760px', maxWidth: 'calc(100vw - 32px)' }"
    >
      <div class="session-detail-scroll">
        <div v-if="toolInvocations.length === 0" style="color: #999">
          暂无工具调用记录
        </div>
        <n-collapse v-else accordion>
          <n-collapse-item
            v-for="it in toolInvocations"
            :key="it.id"
            :title="`${it.name}${it.hasResult ? (it.ok ? '（成功）' : '（失败）') : ''}`"
          >
            <div style="font-size: 12px; color: #999; margin-bottom: 6px">
              {{ it.createdAt }}
            </div>
            <div style="font-size: 12px; color: #999; margin-bottom: 4px">
              请求参数
            </div>
            <pre class="tool-pre">{{ it.arguments }}</pre>
            <template v-if="it.hasResult">
              <div style="font-size: 12px; color: #999; margin: 10px 0 4px">
                返回结果
              </div>
              <pre class="tool-pre">{{ it.output }}</pre>
            </template>
          </n-collapse-item>
        </n-collapse>
      </div>
    </n-modal>

    <n-modal
      v-model:show="showPromptModal"
      preset="dialog"
      title="完整发送内容 (Prompt)"
      :style="{ width: '800px', maxWidth: 'calc(100vw - 32px)' }"
    >
      <div style="max-height: 600px; overflow: auto">
        <pre
          style="
            background: #f5f5f5;
            padding: 12px;
            border-radius: 4px;
            font-size: 12px;
            white-space: pre-wrap;
            word-break: break-all;
          "
          >{{ currentViewPrompt }}</pre
        >
      </div>
    </n-modal>
  </n-layout>
</template>

<script setup>
import { ref, onMounted, onUnmounted, computed, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import { useMessage, useDialog, NIcon } from "naive-ui";
import {
  TrashOutline,
  ArchiveOutline,
  ContractOutline,
  InformationCircleOutline,
} from "@vicons/ionicons5";
import {
  ListProjects,
  ListSessions,
  CreateSession,
  DeleteSession,
  ListAgents,
  SendMessage,
  ListMessages,
  ClearSessionMessages,
  TerminateSession,
  CompressSession,
  ListToolInvocations,
  SearchSessionsByProjectName,
} from "../../wailsjs/go/main/App";

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
const currentChatMode = ref("chat");
const currentChatAgentId = ref(null);
const messages = ref([]);
const inputText = ref("");
const isGenerating = ref(false);

const modeOptions = [
  { label: "对话 (Chat)", value: "chat" },
  { label: "代码 (Code)", value: "code" },
  { label: "工具 (Tool)", value: "tool" },
];

const totalTokenUsage = computed(() => {
  return messages.value.reduce((sum, m) => {
    const n = Number(m?.tokenUsage || 0);
    return sum + (Number.isFinite(n) ? n : 0);
  }, 0);
});

// Create Modal State
const showCreateModal = ref(false);
const newSessionName = ref("");
const selectedMode = ref("chat");
const selectedAgentId = ref(null);
const showDetailModal = ref(false);
const showPromptModal = ref(false);
const currentViewPrompt = ref("");
const toolInvocations = ref([]);
const sessionSearchQuery = ref("");
const searchedSessions = ref([]);
const searchingSessions = ref(false);

const isSearchingSessions = computed(() => {
  return String(sessionSearchQuery.value || "").trim() !== "";
});

const displaySessions = computed(() => {
  return isSearchingSessions.value ? searchedSessions.value : sessions.value;
});

const prevOverflow = {
  html: "",
  body: "",
};

watch(showDetailModal, (visible) => {
  if (typeof document === "undefined") return;
  const html = document.documentElement;
  const body = document.body;
  if (!html || !body) return;

  if (visible) {
    prevOverflow.html = html.style.overflow;
    prevOverflow.body = body.style.overflow;
    html.style.overflow = "hidden";
    body.style.overflow = "hidden";
  } else {
    html.style.overflow = prevOverflow.html;
    body.style.overflow = prevOverflow.body;
  }
});

let sessionSearchTimer = null;
watch(sessionSearchQuery, (q) => {
  if (sessionSearchTimer) clearTimeout(sessionSearchTimer);
  const query = String(q || "").trim();
  if (query === "") {
    searchedSessions.value = [];
    searchingSessions.value = false;
    return;
  }
  sessionSearchTimer = setTimeout(async () => {
    searchingSessions.value = true;
    try {
      const res = await SearchSessionsByProjectName(query);
      if (res.code === 200) {
        searchedSessions.value = res.data || [];
      } else {
        searchedSessions.value = [];
      }
    } finally {
      searchingSessions.value = false;
    }
  }, 250);
});

const userAvatar = `data:image/svg+xml;utf8,${encodeURIComponent(
  `<svg xmlns="http://www.w3.org/2000/svg" width="64" height="64">
     <defs>
       <linearGradient id="g" x1="0" y1="0" x2="1" y2="1">
         <stop offset="0" stop-color="#36ad6a"/>
         <stop offset="1" stop-color="#18a058"/>
       </linearGradient>
     </defs>
     <rect rx="32" ry="32" width="64" height="64" fill="url(#g)"/>
     <text x="32" y="40" text-anchor="middle" font-size="28" fill="#fff" font-family="Arial, sans-serif">我</text>
   </svg>`
)}`;

const aiAvatar = `data:image/svg+xml;utf8,${encodeURIComponent(
  `<svg xmlns="http://www.w3.org/2000/svg" width="64" height="64">
     <defs>
       <linearGradient id="g2" x1="0" y1="0" x2="1" y2="1">
         <stop offset="0" stop-color="#2080f0"/>
         <stop offset="1" stop-color="#1d4ed8"/>
       </linearGradient>
     </defs>
     <rect rx="32" ry="32" width="64" height="64" fill="url(#g2)"/>
     <text x="32" y="40" text-anchor="middle" font-size="26" fill="#fff" font-family="Arial, sans-serif">AI</text>
   </svg>`
)}`;

function formatTime(input) {
  if (!input) return "";
  const d = input instanceof Date ? input : new Date(input);
  if (Number.isNaN(d.getTime())) return "";
  const yyyy = d.getFullYear();
  const mm = String(d.getMonth() + 1).padStart(2, "0");
  const dd = String(d.getDate()).padStart(2, "0");
  const hh = String(d.getHours()).padStart(2, "0");
  const mi = String(d.getMinutes()).padStart(2, "0");
  return `${yyyy}-${mm}-${dd} ${hh}:${mi}`;
}

function applyMessageMeta(item) {
  if (!item) return item;
  if (item.role === "user") {
    item.placement = "end";
    item.avatar = userAvatar;
  } else {
    item.placement = "start";
    item.avatar = aiAvatar;
  }
  return item;
}

// SSE
let eventSource = null;
const toolBubbleIndexByCallId = new Map();

// Computeds
const projectOptions = ref([]);
const agentOptions = ref([]);

function handleToolEvent(tool) {
  if (!tool || !tool.stage) return;

  const toolCallId = tool.toolCallId || "";
  const name = tool.name || "unknown";

  if (tool.stage === "call") {
    const idx =
      messages.value.push({
        role: "tool",
        toolName: name,
        toolArguments: tool.arguments || "",
        toolOutput: "",
        toolOk: null,
        collapsed: true,
        content: "",
        isMarkdown: false,
          createdAt: new Date().toISOString(),
          placement: "start",
          avatar: aiAvatar,
        variant: "outlined",
        shape: "corner",
      }) - 1;
    if (toolCallId) toolBubbleIndexByCallId.set(toolCallId, idx);
    return;
  }

  if (tool.stage === "result") {
    const ok = tool.ok !== false;
    const idx = toolCallId ? toolBubbleIndexByCallId.get(toolCallId) : null;
    if (idx != null && messages.value[idx]) {
      messages.value[idx].toolOk = ok;
      messages.value[idx].toolOutput = tool.output || "";
      return;
    }
    messages.value.push({
      role: "tool",
      toolName: name,
      toolArguments: tool.arguments || "",
      toolOutput: tool.output || "",
      toolOk: ok,
      collapsed: true,
      content: "",
      isMarkdown: false,
        createdAt: new Date().toISOString(),
        placement: "start",
        avatar: aiAvatar,
      variant: "outlined",
      shape: "corner",
    });
  }
}


// Methods
async function loadProjects() {
  try {
    const res = await ListProjects();
    if (res.code === 200) {
      projects.value = res.data || [];
      projectOptions.value = projects.value.map((p) => ({
        label: p.name,
        value: p.id,
      }));

      if (!currentProjectId.value && projects.value.length > 0) {
        currentProjectId.value = projects.value[0].id;
        await loadSessions(currentProjectId.value);
      }
    }
  } catch (e) {
    message.error("加载项目失败");
  }
}

async function loadAgents() {
  try {
    const res = await ListAgents();
    if (res.code === 200) {
      agents.value = res.data || [];
      agentOptions.value = agents.value.map((a) => ({
        label: a.name,
        value: a.id,
      }));

      // Reset to defaults
      currentChatAgentId.value = null;
      selectedAgentId.value = null;
    }
  } catch (e) {
    message.error("加载 Agent 失败");
  }
}

async function loadSessions(pid) {
  try {
    const res = await ListSessions(pid);
    if (res.code === 200) {
      sessions.value = res.data || [];
      if (route.params.sessionId) {
        const sid = parseInt(route.params.sessionId);
        const exists = sessions.value.find((s) => s.id === sid);
        if (exists) {
          currentSessionId.value = sid;
          loadHistory(sid);
        } else {
          currentSessionId.value = null;
        }
      }
    }
  } catch (e) {
    message.error("加载会话失败");
  }
}

async function handleProjectChange(pid) {
  currentProjectId.value = pid;
  currentSessionId.value = null;
  router.push({ name: "Chat" });
  await loadSessions(pid);
}

async function confirmCreateSession() {
  if (!newSessionName.value) {
    message.warning("名称为必填项");
    return;
  }
  try {
    const res = await CreateSession(
      currentProjectId.value,
      newSessionName.value,
      selectedAgentId.value || 0
    );
    if (res.code === 200) {
      showCreateModal.value = false;
      newSessionName.value = "";
      selectedMode.value = "chat";
      selectedAgentId.value = null;

      await loadSessions(currentProjectId.value);
    } else {
      message.error(res.msg);
    }
  } catch (e) {
    message.error("创建失败");
  }
}

async function loadHistory(sid) {
  try {
    const res = await ListMessages(sid);
    if (res.code === 200) {
      const history = res.data || [];
      messages.value = history
        .filter((m) => String(m?.content || "").trim() !== "")
        .map((msg) =>
          applyMessageMeta({
            role: msg.role,
            content: msg.content,
            prompt: msg.prompt,
            isMarkdown: msg.role === "assistant",
            tokenUsage:
              msg.role === "assistant" ? Number(msg.tokenCount || 0) : 0,
            createdAt: msg.createdAt,
          })
        );
      toolBubbleIndexByCallId.clear();
    }
  } catch (e) {
    message.error("加载历史记录失败");
  }
}

function handleSelectSession(sid) {
  currentSessionId.value = sid;
  currentChatMode.value = "chat"; // Default mode
  currentChatAgentId.value = null; // Reset temp agent selection when switching session
  router.push({ name: "Chat", params: { sessionId: sid } });
  loadHistory(sid);
  toolBubbleIndexByCallId.clear();
}

async function handleSelectSearchedSession(session) {
  const pid = session.projectId;
  if (!pid) return;
  sessionSearchQuery.value = "";
  searchedSessions.value = [];
  await handleProjectChange(pid);
  handleSelectSession(session.id);
}

async function handleDeleteSession(sid) {
  dialog.warning({
    title: "删除会话",
    content: "确定要删除吗？",
    positiveText: "确认",
    negativeText: "取消",
    onPositiveClick: async () => {
      try {
        const res = await DeleteSession(sid);
        if (res.code === 200) {
          if (currentSessionId.value === sid) {
            currentSessionId.value = null;
            router.push({ name: "Chat" });
          }
          await loadSessions(currentProjectId.value);
          if (isSearchingSessions.value) {
            const q = String(sessionSearchQuery.value || "").trim();
            if (q) {
              const sres = await SearchSessionsByProjectName(q);
              if (sres.code === 200) searchedSessions.value = sres.data || [];
            }
          }
        }
      } catch (e) {
        message.error("删除失败");
      }
    },
  });
}

async function handleSend(payload) {
  const content = typeof payload === "string" ? payload : inputText.value;
  if (!content || !content.trim()) return;

  inputText.value = "";

  // Optimistic UI update
  messages.value.push(
    applyMessageMeta({
      role: "user",
      content: content,
      isMarkdown: false,
      createdAt: new Date().toISOString(),
    })
  );

  // Placeholder for AI response
  const aiMsgIndex =
    messages.value.push({
      role: "assistant",
      content: "",
      isMarkdown: true,
      tokenUsage: 0,
      createdAt: new Date().toISOString(),
      placement: "start",
      avatar: aiAvatar,
    }) - 1;

  isGenerating.value = true;

  try {
    // Pass currentChatAgentId (if selected) or 0 (use session default)
    const agentId = currentChatAgentId.value || 0;
    const res = await SendMessage(currentSessionId.value, content, agentId);
    if (res.code !== 200) {
      message.error(res.msg);
      messages.value[aiMsgIndex].content = "[错误: " + res.msg + "]";
      isGenerating.value = false;
    }
  } catch (e) {
    message.error("发送失败: " + e);
    if (
      messages.value[aiMsgIndex] &&
      String(messages.value[aiMsgIndex].content || "").trim() === ""
    ) {
      messages.value[aiMsgIndex].content = "[错误: 发送失败]";
    }
    isGenerating.value = false;
  }
}

// SSE Setup
function initSSE() {
  eventSource = new EventSource("http://localhost:8080/events");
  eventSource.onmessage = (event) => {
    try {
      const data = JSON.parse(event.data);
      // Filter by current session
      if (data.sessionId === currentSessionId.value) {
        if (data.tool) {
          handleToolEvent(data.tool);
        }
        if (data.delta) {
          console.log("data.delta", data.delta);

          let appended = false;
          for (let i = messages.value.length - 1; i >= 0; i--) {
            if (messages.value[i]?.role === "assistant") {
              messages.value[i].content += data.delta;
              appended = true;
              break;
            }
          }
          if (!appended) {
            messages.value.push(
              applyMessageMeta({
                role: "assistant",
                content: data.delta,
                isMarkdown: true,
                tokenUsage: 0,
                createdAt: new Date().toISOString(),
              })
            );
          }
        }
        if (data.usage) {
          const usage = Number(data.usage || 0);
          for (let i = messages.value.length - 1; i >= 0; i--) {
            if (messages.value[i]?.role === "assistant") {
              messages.value[i].tokenUsage = Number.isFinite(usage) ? usage : 0;
              break;
            }
          }
        }
        if (data.done) {
          isGenerating.value = false;
          const last = messages.value[messages.value.length - 1];
          if (last?.role === "assistant" && String(last.content || "").trim() === "") {
            messages.value.pop();
          }
        }
        if (data.terminated) {
          message.warning("已终止生成");
          isGenerating.value = false;
          const last = messages.value[messages.value.length - 1];
          if (last?.role === "assistant" && String(last.content || "").trim() === "") {
            messages.value.pop();
          }
          return;
        }
        if (data.error) {
          const last = messages.value[messages.value.length - 1];
          if (
            last?.role === "assistant" &&
            String(last.content || "").trim() === ""
          ) {
            last.content = "[错误: " + data.error + "]";
          }
          message.error("AI 错误: " + data.error);
          isGenerating.value = false;
        }
      }
    } catch (e) {
      console.error("SSE Parse Error", e);
    }
  };
}

async function handleClearSession() {
  if (!currentSessionId.value) return;
  dialog.warning({
    title: "清空会话内容",
    content: "将删除该会话所有历史消息，且无法恢复。确定继续？",
    positiveText: "确认清空",
    negativeText: "取消",
    onPositiveClick: async () => {
      const res = await ClearSessionMessages(currentSessionId.value);
      if (res.code === 200) {
        messages.value = [];
        toolBubbleIndexByCallId.clear();
        isGenerating.value = false;
      } else {
        message.error(res.msg || "清空失败");
      }
    },
  });
}

async function handleTerminateSession() {
  if (!currentSessionId.value || !isGenerating.value) return;
  const res = await TerminateSession(currentSessionId.value);
  if (res.code === 200) {
    isGenerating.value = false;
  } else {
    message.error(res.msg || "终止失败");
  }
}

async function handleCompressSession() {
  if (!currentSessionId.value) return;
  dialog.warning({
    title: "压缩会话",
    content: "压缩会话会用摘要替换全部历史消息，以减少上下文长度。确定继续？",
    positiveText: "确认压缩",
    negativeText: "取消",
    onPositiveClick: async () => {
      const res = await CompressSession(currentSessionId.value);
      if (res.code === 200) {
        await loadSessions(currentProjectId.value);
        await loadHistory(currentSessionId.value);
      } else {
        message.error(res.msg || "压缩失败");
      }
    },
  });
}

async function openSessionDetail() {
  if (!currentSessionId.value) return;
  showDetailModal.value = true;
  const res = await ListToolInvocations(currentSessionId.value);
  if (res.code === 200) {
    toolInvocations.value = res.data || [];
  } else {
    message.error(res.msg || "加载详情失败");
  }
}

function handleViewPrompt(prompt) {
  try {
    const parsed = JSON.parse(prompt);
    currentViewPrompt.value = JSON.stringify(parsed, null, 2);
  } catch (e) {
    currentViewPrompt.value = prompt;
  }
  showPromptModal.value = true;
}

onMounted(() => {
  loadProjects();
  loadAgents();
  initSSE();
});

onUnmounted(() => {
  if (eventSource) {
    eventSource.close();
  }
  if (typeof document !== "undefined") {
    document.documentElement.style.overflow = prevOverflow.html;
    document.body.style.overflow = prevOverflow.body;
  }
});
</script>

<style scoped>
:deep(.el-bubble-content) {
  width: 100%;
  --bubble-content-max-width: 100% !important;
}

:deep(.typer-container) {
  text-align: left;
}

.active-session {
  background-color: rgba(24, 160, 88, 0.1);
  border-radius: 4px;
}
.session-item {
  display: flex;
  align-items: center;
  padding: 4px 8px;
  gap: 8px;
}
.session-name {
  flex: 1;
  min-width: 0;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.tool-pre {
  margin: 0;
  padding: 8px 10px;
  background: rgba(0, 0, 0, 0.05);
  border-radius: 6px;
  white-space: pre-wrap;
  word-break: break-word;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas,
    "Liberation Mono", "Courier New", monospace;
  font-size: 12px;
  line-height: 1.5;
}

.session-detail-scroll {
  max-height: calc(100vh - 260px);
  overflow: auto;
}

.empty-state {
  height: 100%;
  display: flex;
  justify-content: center;
  align-items: center;
}
</style>
