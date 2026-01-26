import { defineStore } from 'pinia'
import { getWorkflows, getWorkflowTasks } from '@/api'

export const useWorkflowStore = defineStore('workflow', {
  state: () => ({
    workflows: [],
    currentWorkflowId: null,
    tasks: [],
    loading: false
  }),
  getters: {
    currentWorkflow: (state) => state.workflows.find(w => w.id === state.currentWorkflowId)
  },
  actions: {
    async fetchWorkflows(sessionId) {
      this.loading = true
      try {
        const res = await getWorkflows(sessionId)
        this.workflows = res || []
      } catch (error) {
        console.error('Failed to fetch workflows:', error)
      } finally {
        this.loading = false
      }
    },
    async fetchTasks(workflowId) {
      this.loading = true
      try {
        const res = await getWorkflowTasks(workflowId)
        this.tasks = res || []
        this.currentWorkflowId = workflowId
      } catch (error) {
        console.error('Failed to fetch tasks:', error)
      } finally {
        this.loading = false
      }
    },
    updateTaskStatus(taskId, status, output) {
      const task = this.tasks.find(t => t.id === taskId)
      if (task) {
        task.status = status
        if (output) {
          task.output = typeof output === 'string' ? output : JSON.stringify(output, null, 2)
        }
      }
    }
  }
})
