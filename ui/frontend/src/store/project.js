import { defineStore } from 'pinia'
import { getProjects } from '@/api'

export const useProjectStore = defineStore('project', {
  state: () => ({
    projects: [],
    currentProjectId: null,
    loading: false
  }),
  getters: {
    currentProject: (state) => state.projects.find(p => p.id === state.currentProjectId)
  },
  actions: {
    async fetchProjects() {
      this.loading = true
      try {
        const res = await getProjects()
        this.projects = res || []
      } catch (error) {
        console.error('Failed to fetch projects:', error)
      } finally {
        this.loading = false
      }
    },
    setCurrentProject(id) {
      this.currentProjectId = id
    }
  },
  persist: true
})
