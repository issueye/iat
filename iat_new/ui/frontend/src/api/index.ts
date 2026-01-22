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
  // Note: Chat stream is handled via EventSource or fetch stream manually, not simple axios get
};
