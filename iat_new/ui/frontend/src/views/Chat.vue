<script setup>
import { ref, onMounted, onUnmounted, computed, watch, nextTick } from "vue";
import { useRoute, useRouter } from "vue-router";
import { useMessage, useDialog, NIcon } from "naive-ui";
import {
  TrashOutline,
  ArchiveOutline,
  ContractOutline,
  InformationCircleOutline,
  ChevronForwardOutline,
  ChevronDownOutline,
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
    if (messages.value[i]?.role === ChatRoles.Assistant) return messages.value[i];
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
    return projects.value.map(p => ({
        label: p.name,
        value: p.id
    }));
});

const agentOptions = computed(() => {
    return agents.value.map(a => ({
        label: a.name,
        value: a.id
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
  if (generationStatus.value === ThinkingStatuses.Cancel) return ThinkingStatuses.Cancel;
  if (generationStatus.value === ThinkingStatuses.Error) return ThinkingStatuses.Error;
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

// Methods
async function loadProjects() {
    try {
        const data = await api.listProjects();
        projects.value = data || [];
        if (!currentProjectId.value && projects.value.length > 0) {
            currentProjectId.value = projects.value[0].id;
            // TODO: loadSessions(currentProjectId.value);
        }
    } catch (e) {
        message.error("加载项目失败: " + e.message);
    }
}

async function loadAgents() {
    // TODO: Implement api.listAgents
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
        createdAt: new Date().toISOString()
    });

    // Add Assistant Placeholder
    const aiMsgIndex = messages.value.push({
        role: ChatRoles.Assistant,
        content: "",
        createdAt: new Date().toISOString()
    }) - 1;

    // Use Fetch for stream instead of EventSource for POST
    try {
        const response = await fetch(SSE.EventsUrl, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                sessionId: currentSessionId.value || 0, // 0 for temporary session if not selected
                message: content,
                agentId: currentChatAgentId.value || 0,
                mode: currentChatMode.value
            })
        });

        if (!response.ok) throw new Error(response.statusText);

        const reader = response.body.getReader();
        const decoder = new TextDecoder();
        
        while (true) {
            const { done, value } = await reader.read();
            if (done) break;
            
            const chunk = decoder.decode(value);
            const lines = chunk.split('\n\n');
            
            for (const line of lines) {
                if (line.startsWith('data: ')) {
                    const dataStr = line.slice(6);
                    try {
                        const data = JSON.parse(dataStr);
                        handleSSEEvent(data, aiMsgIndex);
                    } catch (e) {
                        console.error('Error parsing SSE event', e);
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
    if (data.type === 'chunk') {
        messages.value[aiMsgIndex].content += data.content;
    } else if (data.type === 'tool_call') {
        // Handle tool call UI
        messages.value.push({
            role: ChatRoles.Tool,
            toolName: data.extra.name,
            toolArguments: data.extra.arguments,
            content: "",
            collapsed: true
        });
    } else if (data.type === 'error') {
        message.error(data.content);
        generationStatus.value = ThinkingStatuses.Error;
    }
}

// Lifecycle
onMounted(() => {
    loadProjects();
    // loadAgents();
});

</script>

<style scoped>
/* Reuse existing styles */
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
