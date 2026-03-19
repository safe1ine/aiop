import api from './client'

export interface Agent {
  id: number
  hostname: string
  ip: string
  os: string
  arch: string
  status: 'online' | 'offline'
  last_seen: string
  created_at: string
}

export const agentsApi = {
  list: () => api.get<Agent[]>('/agents').then((r) => r.data),
  delete: (id: number) => api.delete(`/agents/${id}`),
  getEnrollToken: () => api.get<{ enroll_token: string }>('/agents/enroll-token').then((r) => r.data),
}

export const authApi = {
  login: (username: string, password: string) =>
    api.post<{ token: string }>('/auth/login', { username, password }).then((r) => r.data),
}
