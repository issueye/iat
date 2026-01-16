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
            v-for="session in sessions"
            :key="session.id"
            :class="{ 'active-session': currentSessionId === session.id }"
            @click="handleSelectSession(session.id)"
          >
            <div class="session-item">
              <span class="session-name">{{ session.name }}</span>
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
          v-if="sessions.length === 0 && currentProjectId"
          style="text-align: center; color: #666; margin-top: 20px"
        >
          暂无会话
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
              <div
                v-if="item.tokenUsage"
                style="margin-top: 4px; font-size: 12px; color: #999"
              >
                Tokens: {{ item.tokenUsage }}
              </div>
            </template>
          </BubbleList>

          <div style="display: flex; align-items: center; margin-bottom: 8px">
            <n-select
              v-model:value="currentChatAgentId"
              :options="agentOptions"
              size="small"
              style="width: 200px"
              placeholder="使用默认 Agent"
              clearable
            />
            <span style="font-size: 12px; color: #999; margin-left: 8px"
              >(选择此轮对话使用的 Agent)</span
            >
            <span style="font-size: 12px; color: #999; margin-left: 12px"
              >累计 Tokens: {{ totalTokenUsage }}</span
            >
            <n-button
              size="small"
              type="error"
              secondary
              style="margin-left: auto"
              @click="handleClearSession"
            >
              清空会话
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
        <n-form-item label="选择 Agent">
          <n-select v-model:value="selectedAgentId" :options="agentOptions" />
        </n-form-item>
      </n-form>
      <template #action>
        <n-button @click="showCreateModal = false">取消</n-button>
        <n-button type="primary" @click="confirmCreateSession">创建</n-button>
      </template>
    </n-modal>
  </n-layout>
</template>

<script setup>
import { ref, onMounted, onUnmounted, computed } from "vue";
import { useRoute, useRouter } from "vue-router";
import { useMessage, useDialog, NIcon } from "naive-ui";
import { TrashOutline } from "@vicons/ionicons5";
import {
  ListProjects,
  ListSessions,
  CreateSession,
  DeleteSession,
  ListAgents,
  SendMessage,
  ListMessages,
  ClearSessionMessages,
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
const currentChatAgentId = ref(null);
const messages = ref([]);
const inputText = ref("");
const isGenerating = ref(false);
const totalTokenUsage = computed(() => {
  return messages.value.reduce((sum, m) => {
    const n = Number(m?.tokenUsage || 0);
    return sum + (Number.isFinite(n) ? n : 0);
  }, 0);
});

// Create Modal State
const showCreateModal = ref(false);
const newSessionName = ref("");
const selectedAgentId = ref(null);

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

      // Default to "Chat" agent if available
      const chatAgent = agents.value.find(
        (a) => a.name === "Chat" && a.type === "builtin"
      );
      if (chatAgent) {
        selectedAgentId.value = chatAgent.id;
      } else if (agents.value.length > 0) {
        selectedAgentId.value = agents.value[0].id;
      }
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
  if (!newSessionName.value || !selectedAgentId.value) {
    message.warning("名称和 Agent 为必填项");
    return;
  }
  try {
    const res = await CreateSession(
      currentProjectId.value,
      newSessionName.value,
      selectedAgentId.value
    );
    if (res.code === 200) {
      showCreateModal.value = false;
      newSessionName.value = "";

      // Reset selected agent to default
      const chatAgent = agents.value.find(
        (a) => a.name === "Chat" && a.type === "builtin"
      );
      if (chatAgent) {
        selectedAgentId.value = chatAgent.id;
      } else if (agents.value.length > 0) {
        selectedAgentId.value = agents.value[0].id;
      } else {
        selectedAgentId.value = null;
      }

      await loadSessions(currentProjectId.value);
    } else {
      message.error(res.msg);
    }
  } catch (e) {
    message.error("创建失败: " + e);
  }
}

async function loadHistory(sid) {
  try {
    const res = await ListMessages(sid);
    if (res.code === 200) {
      const history = res.data || [];
      messages.value = history.map((msg) => ({
        role: msg.role,
        content: msg.content,
        isMarkdown: msg.role === "assistant",
        tokenUsage: msg.role === "assistant" ? Number(msg.tokenCount || 0) : 0,
      }));
      toolBubbleIndexByCallId.clear();
    }
  } catch (e) {
    message.error("加载历史记录失败");
  }
}

function handleSelectSession(sid) {
  currentSessionId.value = sid;
  currentChatAgentId.value = null; // Reset temp agent selection when switching session
  router.push({ name: "Chat", params: { sessionId: sid } });
  loadHistory(sid);
  toolBubbleIndexByCallId.clear();
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
  messages.value.push({ role: "user", content: content, isMarkdown: false });

  // Placeholder for AI response
  const aiMsgIndex =
    messages.value.push({
      role: "assistant",
      content: "",
      isMarkdown: true,
      tokenUsage: 0,
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
            messages.value.push({
              role: "assistant",
              content: data.delta,
              isMarkdown: true,
              tokenUsage: 0,
            });
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
        }
        if (data.error) {
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

onMounted(() => {
  loadProjects();
  loadAgents();
  initSSE();
});

onUnmounted(() => {
  if (eventSource) {
    eventSource.close();
  }
});
</script>

<style scoped>
:deep(.el-bubble-content) {
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

.empty-state {
  height: 100%;
  display: flex;
  justify-content: center;
  align-items: center;
}
</style>
