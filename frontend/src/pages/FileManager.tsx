import { useRef, useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { filesApi, type FileEntry } from '../api/files'
import { ArrowLeft, Folder, File, Download, Trash2, Upload, ChevronRight, Home } from 'lucide-react'

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

export default function FileManager() {
  const { agentId } = useParams<{ agentId: string }>()
  const id = Number(agentId)
  const navigate = useNavigate()
  const qc = useQueryClient()
  const [path, setPath] = useState('/')
  const uploadRef = useRef<HTMLInputElement>(null)

  const { data, isLoading, error } = useQuery({
    queryKey: ['files', id, path],
    queryFn: () => filesApi.list(id, path),
  })

  const deleteMutation = useMutation({
    mutationFn: (p: string) => filesApi.delete(id, p),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['files', id, path] }),
  })

  const uploadMutation = useMutation({
    mutationFn: (file: File) => filesApi.upload(id, path, file),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['files', id, path] }),
  })

  const handleDownload = async (entry: FileEntry) => {
    const fullPath = path.endsWith('/') ? path + entry.name : path + '/' + entry.name
    const blob = await filesApi.download(id, fullPath)
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = entry.name
    a.click()
    URL.revokeObjectURL(url)
  }

  const navigateTo = (entry: FileEntry) => {
    if (!entry.is_dir) return
    const next = path.endsWith('/') ? path + entry.name : path + '/' + entry.name
    setPath(next)
  }

  const parts = path.split('/').filter(Boolean)
  const crumbs = [{ label: 'root', path: '/' }]
  parts.forEach((p, i) => {
    crumbs.push({ label: p, path: '/' + parts.slice(0, i + 1).join('/') })
  })

  return (
    <div className="h-screen flex flex-col" style={{ background: 'var(--bg)' }}>
      {/* Header */}
      <header style={{ borderBottom: '1px solid var(--border)', background: 'var(--surface)' }}>
        <div className="h-12 px-4 flex items-center gap-3">
          <button
            onClick={() => navigate('/')}
            className="flex items-center gap-1.5 text-sm cursor-pointer transition-colors"
            style={{ color: 'var(--muted)' }}
            onMouseEnter={e => (e.currentTarget.style.color = 'var(--text)')}
            onMouseLeave={e => (e.currentTarget.style.color = 'var(--muted)')}
          >
            <ArrowLeft size={14} />
            返回
          </button>
          <span style={{ color: 'var(--border)' }}>|</span>

          {/* Breadcrumb */}
          <nav className="flex items-center gap-1 text-sm mono overflow-x-auto">
            {crumbs.map((c, i) => (
              <span key={c.path} className="flex items-center gap-1 shrink-0">
                {i === 0
                  ? <button
                      onClick={() => setPath('/')}
                      className="flex items-center gap-1 cursor-pointer transition-colors"
                      style={{ color: i === crumbs.length - 1 ? 'var(--text)' : 'var(--muted)' }}
                      onMouseEnter={e => (e.currentTarget.style.color = 'var(--text)')}
                      onMouseLeave={e => { if (i < crumbs.length - 1) e.currentTarget.style.color = 'var(--muted)' }}
                    >
                      <Home size={12} />
                    </button>
                  : <>
                      <ChevronRight size={12} style={{ color: 'var(--border)' }} />
                      <button
                        onClick={() => setPath(c.path)}
                        className="cursor-pointer transition-colors"
                        style={{ color: i === crumbs.length - 1 ? 'var(--text)' : 'var(--muted)' }}
                        onMouseEnter={e => (e.currentTarget.style.color = 'var(--text)')}
                        onMouseLeave={e => { if (i < crumbs.length - 1) e.currentTarget.style.color = 'var(--muted)' }}
                      >
                        {c.label}
                      </button>
                    </>
                }
              </span>
            ))}
          </nav>

          {/* Upload */}
          <div className="ml-auto">
            <input
              ref={uploadRef}
              type="file"
              className="hidden"
              onChange={(e) => {
                const file = e.target.files?.[0]
                if (file) uploadMutation.mutate(file)
                e.target.value = ''
              }}
            />
            <button
              onClick={() => uploadRef.current?.click()}
              disabled={uploadMutation.isPending}
              className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-sm font-medium cursor-pointer transition-all disabled:opacity-50"
              style={{ background: 'var(--surface2)', color: 'var(--text)', border: '1px solid var(--border)' }}
              onMouseEnter={e => { if (!uploadMutation.isPending) e.currentTarget.style.borderColor = 'var(--green)' }}
              onMouseLeave={e => (e.currentTarget.style.borderColor = 'var(--border)')}
            >
              <Upload size={13} />
              {uploadMutation.isPending ? '上传中...' : '上传'}
            </button>
          </div>
        </div>
      </header>

      {/* Content */}
      <main className="flex-1 overflow-auto">
        {isLoading && (
          <div className="px-5 py-12 text-center text-sm" style={{ color: 'var(--muted)' }}>加载中...</div>
        )}
        {error && (
          <div className="px-5 py-12 text-center text-sm" style={{ color: '#F87171' }}>加载失败</div>
        )}
        {data && (
          <table className="w-full">
            <thead>
              <tr style={{ borderBottom: '1px solid var(--border)' }}>
                {['名称', '大小', '修改时间', '操作'].map(h => (
                  <th key={h} className="text-left px-5 py-3 text-xs font-medium" style={{ color: 'var(--muted)' }}>{h}</th>
                ))}
              </tr>
            </thead>
            <tbody>
              {path !== '/' && (
                <tr
                  className="cursor-pointer transition-colors"
                  style={{ borderBottom: '1px solid var(--border)' }}
                  onClick={() => setPath(path.split('/').slice(0, -1).join('/') || '/')}
                  onMouseEnter={e => (e.currentTarget.style.background = 'var(--surface)')}
                  onMouseLeave={e => (e.currentTarget.style.background = 'transparent')}
                >
                  <td className="px-5 py-3 text-sm mono" style={{ color: 'var(--muted)' }}>
                    <span className="flex items-center gap-2">
                      <Folder size={14} style={{ color: 'var(--muted)' }} />
                      ..
                    </span>
                  </td>
                  <td /><td /><td />
                </tr>
              )}
              {(data.entries ?? []).map((e, i) => {
                const fullPath = path.endsWith('/') ? path + e.name : path + '/' + e.name
                const isLast = i === (data.entries ?? []).length - 1
                return (
                  <tr
                    key={e.name}
                    className="transition-colors"
                    style={{ borderBottom: isLast ? 'none' : '1px solid var(--border)' }}
                    onMouseEnter={ev => (ev.currentTarget.style.background = 'var(--surface)')}
                    onMouseLeave={ev => (ev.currentTarget.style.background = 'transparent')}
                  >
                    <td className="px-5 py-3">
                      <span
                        className={`flex items-center gap-2 text-sm mono ${e.is_dir ? 'cursor-pointer' : ''}`}
                        style={{ color: e.is_dir ? 'var(--green)' : 'var(--text)' }}
                        onClick={() => navigateTo(e)}
                      >
                        {e.is_dir
                          ? <Folder size={14} style={{ color: 'var(--green)' }} />
                          : <File size={14} style={{ color: 'var(--muted)' }} />
                        }
                        {e.name}
                      </span>
                    </td>
                    <td className="px-5 py-3 text-sm mono" style={{ color: 'var(--muted)' }}>
                      {e.is_dir ? '—' : formatSize(e.size)}
                    </td>
                    <td className="px-5 py-3 text-xs mono" style={{ color: 'var(--muted)' }}>
                      {formatTime(e.mod_time)}
                    </td>
                    <td className="px-5 py-3">
                      <div className="flex items-center gap-1">
                        {!e.is_dir && (
                          <button
                            onClick={() => handleDownload(e)}
                            title="下载"
                            className="p-1.5 rounded-lg cursor-pointer transition-all"
                            style={{ color: 'var(--muted)' }}
                            onMouseEnter={e => { e.currentTarget.style.color = 'var(--text)'; e.currentTarget.style.background = 'var(--surface2)' }}
                            onMouseLeave={e => { e.currentTarget.style.color = 'var(--muted)'; e.currentTarget.style.background = 'transparent' }}
                          >
                            <Download size={14} />
                          </button>
                        )}
                        <button
                          onClick={() => { if (confirm(`删除 ${e.name}？`)) deleteMutation.mutate(fullPath) }}
                          title="删除"
                          className="p-1.5 rounded-lg cursor-pointer transition-all"
                          style={{ color: 'var(--muted)' }}
                          onMouseEnter={ev => { ev.currentTarget.style.color = '#F87171'; ev.currentTarget.style.background = 'rgba(239,68,68,0.1)' }}
                          onMouseLeave={ev => { ev.currentTarget.style.color = 'var(--muted)'; ev.currentTarget.style.background = 'transparent' }}
                        >
                          <Trash2 size={14} />
                        </button>
                      </div>
                    </td>
                  </tr>
                )
              })}
            </tbody>
          </table>
        )}
      </main>
    </div>
  )
}
