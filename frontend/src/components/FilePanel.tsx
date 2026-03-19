import { useRef, useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { filesApi, type FileEntry } from '../api/files'
import { Folder, File, Download, Trash2, Upload, ChevronRight, Home } from 'lucide-react'

function formatSize(bytes: number) {
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  if (bytes < 1024 * 1024 * 1024) return `${(bytes / 1024 / 1024).toFixed(1)} MB`
  return `${(bytes / 1024 / 1024 / 1024).toFixed(1)} GB`
}

function formatTime(unix: number) {
  const d = new Date(unix * 1000)
  return d.toLocaleDateString() + ' ' + d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
}

export function FilePanel({ agentId }: { agentId: number }) {
  const qc = useQueryClient()
  const [path, setPath] = useState('/')
  const uploadRef = useRef<HTMLInputElement>(null)

  const { data, isLoading, error } = useQuery({
    queryKey: ['files', agentId, path],
    queryFn: () => filesApi.list(agentId, path),
  })

  const deleteMutation = useMutation({
    mutationFn: (p: string) => filesApi.delete(agentId, p),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['files', agentId, path] }),
  })

  const uploadMutation = useMutation({
    mutationFn: (file: File) => filesApi.upload(agentId, path, file),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['files', agentId, path] }),
  })

  const handleDownload = async (entry: FileEntry) => {
    const fullPath = path.endsWith('/') ? path + entry.name : path + '/' + entry.name
    const blob = await filesApi.download(agentId, fullPath)
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url; a.download = entry.name; a.click()
    URL.revokeObjectURL(url)
  }

  const parts = path.split('/').filter(Boolean)
  const crumbs = [{ label: 'root', path: '/' }, ...parts.map((p, i) => ({
    label: p, path: '/' + parts.slice(0, i + 1).join('/')
  }))]

  return (
    <div style={{ display: 'flex', flexDirection: 'column', height: '100%', background: 'var(--bg)' }}>
      {/* Path bar */}
      <div className="flex items-center gap-2 px-4 py-2" style={{ borderBottom: '1px solid var(--border)', background: 'var(--surface)', flexShrink: 0 }}>
        <nav className="flex items-center gap-1 text-sm mono flex-1 overflow-x-auto">
          {crumbs.map((c, i) => (
            <span key={c.path} className="flex items-center gap-1 shrink-0">
              {i > 0 && <ChevronRight size={11} style={{ color: 'var(--border)' }} />}
              <button
                onClick={() => setPath(c.path)}
                className="cursor-pointer transition-colors flex items-center gap-1"
                style={{ color: i === crumbs.length - 1 ? 'var(--text)' : 'var(--muted)' }}
                onMouseEnter={e => (e.currentTarget.style.color = 'var(--text)')}
                onMouseLeave={e => { if (i < crumbs.length - 1) e.currentTarget.style.color = 'var(--muted)' }}
              >
                {i === 0 ? <Home size={12} /> : c.label}
              </button>
            </span>
          ))}
        </nav>
        <input ref={uploadRef} type="file" className="hidden"
          onChange={e => { const f = e.target.files?.[0]; if (f) uploadMutation.mutate(f); e.target.value = '' }} />
        <button
          onClick={() => uploadRef.current?.click()}
          disabled={uploadMutation.isPending}
          className="flex items-center gap-1.5 px-3 py-1 rounded-lg text-xs font-medium cursor-pointer transition-all disabled:opacity-50 shrink-0"
          style={{ background: 'var(--surface2)', color: 'var(--text)', border: '1px solid var(--border)' }}
          onMouseEnter={e => { if (!uploadMutation.isPending) e.currentTarget.style.borderColor = 'var(--green)' }}
          onMouseLeave={e => (e.currentTarget.style.borderColor = 'var(--border)')}
        >
          <Upload size={12} />
          {uploadMutation.isPending ? '上传中...' : '上传'}
        </button>
      </div>

      {/* File list */}
      <div style={{ flex: 1, overflowY: 'auto' }}>
        {isLoading && <div className="px-5 py-10 text-center text-sm" style={{ color: 'var(--muted)' }}>加载中...</div>}
        {error && <div className="px-5 py-10 text-center text-sm" style={{ color: '#F87171' }}>加载失败</div>}
        {data && (
          <table className="w-full">
            <thead>
              <tr style={{ borderBottom: '1px solid var(--border)' }}>
                {['名称', '大小', '修改时间', ''].map(h => (
                  <th key={h} className="text-left px-4 py-2 text-xs font-medium" style={{ color: 'var(--muted)' }}>{h}</th>
                ))}
              </tr>
            </thead>
            <tbody>
              {path !== '/' && (
                <tr className="cursor-pointer transition-colors" style={{ borderBottom: '1px solid var(--border)' }}
                  onClick={() => setPath(path.split('/').slice(0, -1).join('/') || '/')}
                  onMouseEnter={e => (e.currentTarget.style.background = 'var(--surface)')}
                  onMouseLeave={e => (e.currentTarget.style.background = 'transparent')}>
                  <td className="px-4 py-2.5 text-sm mono" style={{ color: 'var(--muted)' }}>
                    <span className="flex items-center gap-2"><Folder size={13} />.. </span>
                  </td>
                  <td /><td /><td />
                </tr>
              )}
              {(data.entries ?? []).map((e, i) => {
                const fullPath = path.endsWith('/') ? path + e.name : path + '/' + e.name
                const isLast = i === (data.entries ?? []).length - 1
                return (
                  <tr key={e.name} className="transition-colors"
                    style={{ borderBottom: isLast ? 'none' : '1px solid var(--border)' }}
                    onMouseEnter={ev => (ev.currentTarget.style.background = 'var(--surface)')}
                    onMouseLeave={ev => (ev.currentTarget.style.background = 'transparent')}>
                    <td className="px-4 py-2.5">
                      <span className={`flex items-center gap-2 text-sm mono ${e.is_dir ? 'cursor-pointer' : ''}`}
                        style={{ color: e.is_dir ? 'var(--green)' : 'var(--text)' }}
                        onClick={() => { if (e.is_dir) setPath(fullPath) }}>
                        {e.is_dir
                          ? <Folder size={13} style={{ color: 'var(--green)' }} />
                          : <File size={13} style={{ color: 'var(--muted)' }} />}
                        {e.name}
                      </span>
                    </td>
                    <td className="px-4 py-2.5 text-xs mono" style={{ color: 'var(--muted)' }}>{e.is_dir ? '—' : formatSize(e.size)}</td>
                    <td className="px-4 py-2.5 text-xs mono" style={{ color: 'var(--muted)' }}>{formatTime(e.mod_time)}</td>
                    <td className="px-4 py-2.5">
                      <div className="flex items-center gap-1">
                        {!e.is_dir && (
                          <button onClick={() => handleDownload(e)} title="下载"
                            className="p-1 rounded cursor-pointer transition-all" style={{ color: 'var(--muted)' }}
                            onMouseEnter={ev => { ev.currentTarget.style.color = 'var(--text)'; ev.currentTarget.style.background = 'var(--surface2)' }}
                            onMouseLeave={ev => { ev.currentTarget.style.color = 'var(--muted)'; ev.currentTarget.style.background = 'transparent' }}>
                            <Download size={13} />
                          </button>
                        )}
                        <button onClick={() => { if (confirm(`删除 ${e.name}？`)) deleteMutation.mutate(fullPath) }} title="删除"
                          className="p-1 rounded cursor-pointer transition-all" style={{ color: 'var(--muted)' }}
                          onMouseEnter={ev => { ev.currentTarget.style.color = '#F87171'; ev.currentTarget.style.background = 'rgba(239,68,68,0.1)' }}
                          onMouseLeave={ev => { ev.currentTarget.style.color = 'var(--muted)'; ev.currentTarget.style.background = 'transparent' }}>
                          <Trash2 size={13} />
                        </button>
                      </div>
                    </td>
                  </tr>
                )
              })}
            </tbody>
          </table>
        )}
      </div>
    </div>
  )
}
