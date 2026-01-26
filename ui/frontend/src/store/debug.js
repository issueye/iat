import { defineStore } from 'pinia'

export const useDebugStore = defineStore('debug', {
  state: () => ({
    logs: [],
    enabled: true
  }),
  actions: {
    addLog(type, data) {
      if (!this.enabled) return
      this.logs.unshift({
        id: Date.now() + Math.random(),
        time: new Date().toLocaleTimeString(),
        type,
        data: typeof data === 'string' ? data : JSON.stringify(data, null, 2)
      })
      if (this.logs.length > 100) this.logs.pop()
    },
    clear() {
      this.logs = []
    }
  }
})
