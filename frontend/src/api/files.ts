import api from './client'

export interface FileEntry {
  name: string
  is_dir: boolean
  size: number
  mod_time: number
  mode: string
}

export interface FileListResponse {
  path: string
  entries: FileEntry[]
}

export const filesApi = {
  list: (agentId: number, path: string) =>
    api.get<FileListResponse>(`/agents/${agentId}/files`, { params: { path } }).then((r) => r.data),

  download: (agentId: number, path: string) =>
    api.get(`/agents/${agentId}/files/download`, { params: { path }, responseType: 'blob' }).then((r) => r.data as Blob),

  upload: (agentId: number, path: string, file: File) => {
    const form = new FormData()
    form.append('file', file)
    return api.post(`/agents/${agentId}/files/upload`, form, { params: { path } })
  },

  delete: (agentId: number, path: string) =>
    api.delete(`/agents/${agentId}/files`, { params: { path } }),
}
