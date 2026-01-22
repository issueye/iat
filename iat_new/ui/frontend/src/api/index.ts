import axios from 'axios';

const API_BASE_URL = 'http://localhost:8080/api';

const client = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
});

export const api = {
  checkHealth: async () => {
    const resp = await client.get('/health');
    return resp.data;
  },
  listProjects: async () => {
      const resp = await client.get('/projects');
      return resp.data;
  },
  createProject: async (name: string, description: string, path: string) => {
      const resp = await client.post('/projects', { name, description, path });
      return resp.data;
  },
  
  // Agent Methods
  listAgents: async () => {
    const resp = await client.get('/agents');
    return resp.data;
  },

  // Session Methods
  listSessions: async (projectId: number) => {
    const resp = await client.get(`/sessions?projectId=${projectId}`);
    return resp.data;
  },
  createSession: async (name: string, projectId: number, agentId: number) => {
    const resp = await client.post('/sessions', { name, projectId, agentId });
    return resp.data;
  },
  updateSession: async (id: number, name: string) => {
    const resp = await client.put(`/sessions/${id}`, { name });
    return resp.data;
  },
  deleteSession: async (id: number) => {
    const resp = await client.delete(`/sessions/${id}`);
    return resp.data;
  },
  getSessionMessages: async (id: number) => {
    const resp = await client.get(`/sessions/${id}/messages`);
    return resp.data;
  },

  // Note: Chat stream is handled via EventSource or fetch stream manually
};
