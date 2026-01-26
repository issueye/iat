import { defineStore } from 'pinia'
import { getAgents, getModels, getModes } from '@/api'

export const useAgentStore = defineStore('agent', {
  state: () => ({
    agents: [],
    models: [],
    modes: [],
    currentAgent: null,
    loading: false
  }),
  actions: {
    async fetchAll() {
      this.loading = true
      try {
        const [agentsRes, modelsRes, modesRes] = await Promise.all([
          getAgents(),
          getModels(),
          getModes()
        ])
        this.agents = agentsRes || []
        this.models = modelsRes || []
        this.modes = modesRes || []
      } catch (error) {
        console.error('Failed to fetch agent data:', error)
      } finally {
        this.loading = false
      }
    },
    setCurrentAgent(agent) {
      this.currentAgent = agent
    }
  },
  persist: true
})
