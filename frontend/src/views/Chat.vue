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
          placeholder="Select Project"
          :options="projectOptions"
          @update:value="handleProjectChange"
        />
      </div>
      
      <n-button block dashed style="margin-bottom: 12px" @click="handleCreateSession" :disabled="!currentProjectId">
        + New Chat
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
              <n-button size="tiny" text type="error" @click.stop="handleDeleteSession(session.id)">
                <template #icon><n-icon><TrashOutline /></n-icon></template>
              </n-button>
            </div>
          </n-list-item>
        </n-list>
        <div v-if="sessions.length === 0 && currentProjectId" style="text-align: center; color: #666; margin-top: 20px;">
          No sessions
        </div>
      </n-scrollbar>
    </n-layout-sider>

    <!-- Chat Area -->
    <n-layout-content content-style="display: flex; flex-direction: column; height: 100%;">
      <div v-if="!currentSessionId" class="empty-state">
        <n-empty description="Select a project and session to start chatting" />
      </div>
      
      <template v-else>
        <!-- Messages Area -->
        <div class="messages-container">
          <n-scrollbar ref="scrollbarRef">
            <div class="message-list">
              <div v-for="(msg, index) in messages" :key="index" :class="['message-row', msg.role]">
                <div class="message-avatar">
                  <n-avatar size="small" :style="{ backgroundColor: msg.role === 'user' ? '#18a058' : '#2080f0' }">
                    {{ msg.role === 'user' ? 'U' : 'AI' }}
                  </n-avatar>
                </div>
                <div class="message-content">
                  <div class="message-bubble">
                    {{ msg.content }}
                  </div>
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
            placeholder="Type a message..."
            @keydown.enter.prevent="handleSend"
          />
          <n-button type="primary" style="margin-left: 12px" @click="handleSend" :disabled="!inputText.trim()">
            Send
          </n-button>
        </div>
      </template>
    </n-layout-content>
  </n-layout>
</template>

<script setup>
import { ref, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useMessage, useDialog, NIcon } from 'naive-ui'
import { TrashOutline } from '@vicons/ionicons5'
import { ListProjects, ListSessions, CreateSession, DeleteSession } from '../../wailsjs/go/main/App'

const route = useRoute()
const router = useRouter()
const message = useMessage()
const dialog = useDialog()

// State
const projects = ref([])
const sessions = ref([])
const currentProjectId = ref(null)
const currentSessionId = ref(null)
const messages = ref([]) // Temporary local state for messages
const inputText = ref('')

// Computeds
const projectOptions = ref([])

// Methods
async function loadProjects() {
  try {
    const res = await ListProjects()
    if (res.code === 200) {
      projects.value = res.data || []
      projectOptions.value = projects.value.map(p => ({ label: p.name, value: p.id }))
      
      // Auto select first project if none selected
      if (!currentProjectId.value && projects.value.length > 0) {
        currentProjectId.value = projects.value[0].id
        await loadSessions(currentProjectId.value)
      }
    }
  } catch (e) {
    message.error('Failed to load projects')
  }
}

async function loadSessions(pid) {
  try {
    const res = await ListSessions(pid)
    if (res.code === 200) {
      sessions.value = res.data || []
      // If route has sessionId, verify it belongs to this project
      if (route.params.sessionId) {
        const sid = parseInt(route.params.sessionId)
        const exists = sessions.value.find(s => s.id === sid)
        if (exists) {
          currentSessionId.value = sid
        } else {
          // If session not found in this project, clear it
          currentSessionId.value = null
        }
      }
    }
  } catch (e) {
    message.error('Failed to load sessions')
  }
}

async function handleProjectChange(pid) {
  currentProjectId.value = pid
  currentSessionId.value = null
  router.push({ name: 'Chat' }) // Clear sessionId param
  await loadSessions(pid)
}

async function handleCreateSession() {
  try {
    const name = `New Chat ${sessions.value.length + 1}`
    const res = await CreateSession(currentProjectId.value, name)
    if (res.code === 200) {
      await loadSessions(currentProjectId.value)
      // Select the new session (assuming it's the last one or we need ID from create response... 
      // Current CreateSession doesn't return ID, we might need to fix backend if we want auto-select.
      // For now, just reload.)
    } else {
      message.error(res.msg)
    }
  } catch (e) {
    message.error('Failed to create session')
  }
}

function handleSelectSession(sid) {
  currentSessionId.value = sid
  router.push({ name: 'Chat', params: { sessionId: sid } })
  // TODO: Load history messages for this session
  messages.value = [] // Clear previous messages
}

async function handleDeleteSession(sid) {
  dialog.warning({
    title: 'Delete Chat',
    content: 'Are you sure?',
    positiveText: 'Confirm',
    negativeText: 'Cancel',
    onPositiveClick: async () => {
      try {
        const res = await DeleteSession(sid)
        if (res.code === 200) {
          if (currentSessionId.value === sid) {
            currentSessionId.value = null
            router.push({ name: 'Chat' })
          }
          await loadSessions(currentProjectId.value)
        }
      } catch (e) {
        message.error('Failed to delete')
      }
    }
  })
}

function handleSend() {
  if (!inputText.value.trim()) return
  
  // Add user message
  messages.value.push({
    role: 'user',
    content: inputText.value
  })
  
  const prompt = inputText.value
  inputText.value = ''
  
  // Simulate AI response (Placeholder for next phase)
  setTimeout(() => {
    messages.value.push({
      role: 'assistant',
      content: `Echo: ${prompt} (AI logic not connected yet)`
    })
  }, 500)
}

onMounted(() => {
  loadProjects()
})
</script>

<style scoped>
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
  background-color: #f5f5f5; /* Light bg for contrast */
}

/* Dark mode override if needed, but simplistic for now */
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
  max-width: 70%;
}

.message-bubble {
  padding: 10px 14px;
  border-radius: 8px;
  background-color: #fff;
  box-shadow: 0 1px 2px rgba(0,0,0,0.1);
  white-space: pre-wrap;
}

.message-row.user .message-bubble {
  background-color: #e7f5ee; /* Light green */
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
  border-top: 1px solid rgba(255,255,255,0.1);
  display: flex;
  align-items: flex-end;
  background-color: var(--n-color);
}
</style>
