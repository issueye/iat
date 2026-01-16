<template>
  <n-layout has-sider style="height: 100%">
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
        <!-- Messages Area -->
        <div class="messages-container">
          <n-scrollbar ref="scrollbarRef">
            <div class="message-list">
              <div
                v-for="(msg, index) in messages"
                :key="index"
                :class="['message-row', msg.role]"
              >
                <div class="message-avatar">
                  <n-avatar
                    size="small"
                    :style="{
                      backgroundColor:
                        msg.role === 'user' ? '#18a058' : '#2080f0',
                    }"
                  >
                    {{ msg.role === "user" ? "U" : "AI" }}
                  </n-avatar>
                </div>
                <div class="message-content">
                  <div class="message-bubble" v-if="msg.role === 'user'">
                    {{ msg.content }}
                  </div>
                  <div
                    class="message-bubble markdown-body"
                    v-else
                    v-html="renderMarkdown(msg.content)"
                  ></div>
                </div>
              </div>
            </div>
          </n-scrollbar>
        </div>

        <!-- Input Area -->
        <div class="input-area">
          <n-input
            v-model:value="inputText"
            type="textarea"
            :autosize="{ minRows: 1, maxRows: 5 }"
            placeholder="输入消息..."
            @keydown.enter.prevent="handleSend"
          />
          <n-button
            type="primary"
            style="margin-left: 12px"
            @click="handleSend"
            :disabled="!inputText.trim()"
          >
            发送
          </n-button>
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
import { ref, onMounted, watch, onUnmounted, nextTick } from "vue";
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
} from "../../wailsjs/go/main/App";
import { renderMarkdown } from "../utils/markdown";

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
const messages = ref([]);
const inputText = ref("");
const scrollbarRef = ref(null);

// Create Modal State
const showCreateModal = ref(false);
const newSessionName = ref("");
const selectedAgentId = ref(null);

// SSE
let eventSource = null;

// Computeds
const projectOptions = ref([]);
const agentOptions = ref([]);

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
      selectedAgentId.value = null;
      await loadSessions(currentProjectId.value);
    } else {
      message.error(res.msg);
    }
  } catch (e) {
    message.error("创建失败: " + e);
  }
}

function handleSelectSession(sid) {
  currentSessionId.value = sid;
  router.push({ name: "Chat", params: { sessionId: sid } });
  messages.value = []; // TODO: Load history
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

function scrollToBottom() {
  nextTick(() => {
    if (scrollbarRef.value) {
      scrollbarRef.value.scrollTo({ top: 99999, behavior: "smooth" });
    }
  });
}

async function handleSend() {
  if (!inputText.value.trim()) return;

  const content = inputText.value;
  inputText.value = "";

  // Optimistic UI update
  messages.value.push({ role: "user", content: content });
  scrollToBottom();

  // Placeholder for AI response
  const aiMsgIndex =
    messages.value.push({ role: "assistant", content: "" }) - 1;

  try {
    const res = await SendMessage(currentSessionId.value, content);
    if (res.code !== 200) {
      message.error(res.msg);
      messages.value[aiMsgIndex].content = "[错误: " + res.msg + "]";
    }
  } catch (e) {
    message.error("发送失败: " + e);
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
        if (data.delta) {
          // Find last assistant message and append
          // Ideally we should have message ID, but for now assuming last message is the one streaming
          const lastMsg = messages.value[messages.value.length - 1];
          if (lastMsg && lastMsg.role === "assistant") {
            lastMsg.content += data.delta;
          } else {
            // If for some reason last msg is not assistant (should not happen with our logic), append new
            messages.value.push({ role: "assistant", content: data.delta });
          }
          scrollToBottom();
        }
        if (data.error) {
          message.error("AI 错误: " + data.error);
        }
      }
    } catch (e) {
      console.error("SSE Parse Error", e);
    }
  };
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
/* Same styles as before */
.active-session {
  background-color: rgba(24, 160, 88, 0.1);
  border-radius: 4px;
}
.session-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 4px 8px;
}
.session-name {
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: 180px;
}

.empty-state {
  height: 100%;
  display: flex;
  justify-content: center;
  align-items: center;
}

.messages-container {
  flex: 1;
  padding: 20px;
  overflow: hidden;
  background-color: #f5f5f5;
}

@media (prefers-color-scheme: dark) {
  .messages-container {
    background-color: #101014;
  }
}

.message-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.message-row {
  display: flex;
  gap: 12px;
}

.message-row.user {
  flex-direction: row-reverse;
}

.message-content {
  max-width: 80%;
}

.message-bubble {
  padding: 10px 14px;
  border-radius: 8px;
  background-color: #fff;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.1);
  white-space: pre-wrap;
  word-wrap: break-word;
}

.message-row.user .message-bubble {
  background-color: #e7f5ee;
  color: #18a058;
}

@media (prefers-color-scheme: dark) {
  .message-bubble {
    background-color: #26262a;
  }
  .message-row.user .message-bubble {
    background-color: #163524;
  }
}

.input-area {
  padding: 16px;
  border-top: 1px solid rgba(255, 255, 255, 0.1);
  display: flex;
  align-items: flex-end;
  background-color: var(--n-color);
}

/* Markdown Styles */
:deep(.markdown-body) {
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Helvetica, Arial,
    sans-serif;
}
:deep(.markdown-body pre) {
  background-color: #282c34;
  border-radius: 6px;
  padding: 12px;
  overflow: auto;
  color: #abb2bf;
}
:deep(.markdown-body code) {
  font-family: "SFMono-Regular", Consolas, "Liberation Mono", Menlo, Courier,
    monospace;
}
:deep(.markdown-body p) {
  margin-bottom: 8px;
}
:deep(.markdown-body p:last-child) {
  margin-bottom: 0;
}
</style>
