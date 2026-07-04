import api from './client'

export const authAPI = {
  login: (data: { email: string; password: string; device_name: string; device_type: string; device_fingerprint: string }) =>
    api.post('/auth/login', data),
  register: (data: { username: string; email: string; password: string; device_name: string; device_type: string; device_fingerprint: string }) =>
    api.post('/auth/register', data),
  refresh: (refreshToken: string) =>
    api.post('/auth/refresh', { refresh_token: refreshToken }),
}

export const userAPI = {
  getProfile: () => api.get('/users/me'),
}

export const deviceAPI = {
  list: (params?: { page?: number; page_size?: number }) =>
    api.get('/devices', { params }),
  get: (id: string) => api.get(`/devices/${id}`),
  update: (id: string, data: { device_name?: string }) =>
    api.put(`/devices/${id}`, data),
  delete: (id: string) => api.delete(`/devices/${id}`),
}

export const tunnelAPI = {
  list: (params?: { page?: number; page_size?: number; status?: string }) =>
    api.get('/tunnels', { params }),
  create: (data: { name: string; protocol: string; local_host: string; local_port: number; remote_port?: number; domain?: string }) =>
    api.post('/tunnels', data),
  get: (id: string) => api.get(`/tunnels/${id}`),
  update: (id: string, data: Record<string, unknown>) =>
    api.put(`/tunnels/${id}`, data),
  delete: (id: string) => api.delete(`/tunnels/${id}`),
  start: (id: string) => api.post(`/tunnels/${id}/start`),
  stop: (id: string) => api.post(`/tunnels/${id}/stop`),
  stats: (id: string) => api.get(`/tunnels/${id}/stats`),
}

export const trafficAPI = {
  overview: () => api.get('/traffic/overview'),
  byTunnel: (id: string) => api.get(`/traffic/tunnel/${id}`),
}

export const adminAPI = {
  listUsers: (params?: { page?: number; page_size?: number }) =>
    api.get('/admin/users', { params }),
  updateUser: (id: string, data: Record<string, unknown>) =>
    api.put(`/admin/users/${id}`, data),
  systemStats: () => api.get('/admin/system/stats'),
  auditLogs: (params?: { page?: number; page_size?: number }) =>
    api.get('/admin/audit-logs', { params }),
}

export const configAPI = {
  export: () => api.get('/config/export'),
  import: (config: string) => api.post('/config/import', { config }),
}
