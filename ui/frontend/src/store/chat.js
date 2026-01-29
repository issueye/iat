import { defineStore } from 'pinia'
import { getSessions, getMessages, deleteMessages } from '@/api'
import { WSClient } from '@/utils/websocket'
import { useDebugStore } from './debug'
import { useWorkflowStore } from './workflow'

export const useChatStore = defineStore('chat', {
  state: () => ({
    sessions: [],
    currentSessionId: null,
    messages: [],
    streaming: false,
    input: '',
    loading: false,
    ws: null
  }),
  getters: {
    currentSession: (state) => state.sessions.find(s => s.id === state.currentSessionId)
  },
  actions: {
    initWS() {
      if (this.ws) return
      const debugStore = useDebugStore()
      const workflowStore = useWorkflowStore()
      const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
      const host = window.location.host || 'localhost:8080'
      this.ws = new WSClient(`${protocol}//${host}/api/ws`)
      this.ws.connect()
      
      this.ws.on('*', (msg) => {
        // Handle global real-time updates
        console.log('Global WS Msg:', msg)
        debugStore.addLog(msg.action || msg.type, msg.payload || msg)

        // Handle session updates
        if (msg.action === 'session_updated') {
          const { sessionId, name } = msg.payload
          const session = this.sessions.find(s => s.id === sessionId)
          if (session) {
            session.name = name
          }
        }

        // Handle task status updates
        if (msg.action === 'task_status') {
          const { taskId, status, output } = msg.payload || {}
          workflowStore.updateTaskStatus(taskId, status, output)
        }
      })
    },
    async fetchSessions(projectId) {
      try {
        const res = await getSessions(projectId)
        this.sessions = res || []
      } catch (error) {
        console.error('Failed to fetch sessions:', error)
      }
    },
    async fetchMessages(sessionId) {
      this.loading = true
      try {
        const res = await getMessages(sessionId)
        this.messages = res || []
        this.currentSessionId = sessionId
      } catch (error) {
        console.error('Failed to fetch messages:', error)
      } finally {
        this.loading = false
      }
    },
    async clearHistory() {
      if (!this.currentSessionId) return
      try {
        await deleteMessages(this.currentSessionId)
        this.messages = []
      } catch (error) {
        console.error('Failed to clear history:', error)
      }
    },
    addMessage(msg) {
      return this.messages.push(msg)
    },
    removeMessage(id) {
      const idx = this.messages.findIndex(m => m.id === id)
      if (idx !== -1) {
        this.messages.splice(idx, 1)
      }
    },
    updateLastMessage(content) {
      if (this.messages.length > 0) {
        const last = this.messages[this.messages.length - 1]
        if (last.role === 'assistant') {
          last.content += content
        }
      }
    },
    setStreaming(val) {
      this.streaming = val
    }
  },
  persist: {
    paths: ['currentSessionId']
  }
})
