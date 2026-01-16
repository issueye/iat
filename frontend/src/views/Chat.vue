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
              <div v-if="!item.isMarkdown">
                {{ item.content }}
              </div>
              <XMarkdown
                v-else
                :markdown="item.content"
                default-theme-mode="light"
                style="text-align: left"
              />
            </template>
          </BubbleList>
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
import { ref, onMounted, onUnmounted } from "vue";
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
const isGenerating = ref(false);

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

async function handleSend(payload) {
  const content = typeof payload === "string" ? payload : inputText.value;
  if (!content || !content.trim()) return;

  inputText.value = "";

  // Optimistic UI update
  messages.value.push({ role: "user", content: content });

  // Placeholder for AI response
  const aiMsgIndex =
    messages.value.push({
      role: "assistant",
      content: "",
      isMarkdown: true,
    }) - 1;

  isGenerating.value = true;

  try {
    const res = await SendMessage(currentSessionId.value, content);
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
        if (data.delta) {
          console.log("data.delta", data.delta);

          // Find last assistant message and append
          const lastMsg = messages.value[messages.value.length - 1];
          if (lastMsg && lastMsg.role === "assistant") {
            lastMsg.content += data.delta;
          } else {
            messages.value.push({
              role: "assistant",
              content: data.delta,
              isMarkdown: true,
            });
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
</style>
