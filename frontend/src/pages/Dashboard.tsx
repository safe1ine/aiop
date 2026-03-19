import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Terminal, FolderOpen, Trash2, Plus, X, Copy, Check, LogOut, Server } from 'lucide-react'
import { agentsApi, type Agent } from '../api/agents'
import { useAuthStore } from '../store/auth'
import { useTabStore } from '../store/tabs'
import { TerminalPanel } from '../components/TerminalPanel'
import { FilePanel } from '../components/FilePanel'

function CopyButton({ text }: { text: string }) {
  const [copied, setCopied] = useState(false)
  return (
    <button onClick={() => { navigator.clipboard.writeText(text); setCopied(true); setTimeout(() => setCopied(false), 2000) }}
      className="cursor-pointer transition-colors" style={{ color: copied ? 'var(--green)' : 'var(--muted)' }} title="复制">
      {copied ? <Check size={13} /> : <Copy size={13} />}
    </button>
  )
}

export default function Dashboard() {
  const { data: agentsRaw, isLoading } = useQuery({ queryKey: ['agents'], queryFn: agentsApi.list, refetchInterval: 5000 })
  const agents = agentsRaw ?? []
  const [showInstall, setShowInstall] = useState(false)
  const { data: enrollData } = useQuery({ queryKey: ['enroll-token'], queryFn: agentsApi.getEnrollToken, enabled: showInstall })
  const qc = useQueryClient()
  const logout = useAuthStore(s => s.logout)
  const { tabs, activeId, open, close, setActive } = useTabStore()

  const deleteMutation = useMutation({
    mutationFn: (id: number) => agentsApi.delete(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['agents'] }),
  })

  const wsOrigin = location.origin.replace(/^http/, 'ws')
  const installCmd = enrollData
    ? `sudo AIPO_SERVER_URL=${wsOrigin}/ws/agent \\\n     AIPO_ENROLL_TOKEN=${enrollData.enroll_token} \\\n     bash -c "$(curl -fsSL ${location.origin}/install.sh)"`
    : '加载中...'

  const online = agents.filter((a: Agent) => a.status === 'online').length
  const activeTab = tabs.find(t => t.id === activeId)

  return (
    <div style={{ display: 'flex', height: '100vh', background: 'var(--bg)', overflow: 'hidden' }}>
      {/* Sidebar */}
      <aside style={{ width: 220, flexShrink: 0, display: 'flex', flexDirection: 'column', borderRight: '1px solid var(--border)', background: 'var(--surface)' }}>
        {/* Logo */}
        <div className="px-4 py-3 flex items-center justify-between" style={{ borderBottom: '1px solid var(--border)' }}>
          <div className="flex items-center gap-2">
            <div className="w-5 h-5 rounded flex items-center justify-center" style={{ background: 'var(--green)' }}>
              <svg width="10" height="10" viewBox="0 0 16 16" fill="none">
                <rect x="2" y="2" width="5" height="5" rx="1" fill="black" />
                <rect x="9" y="2" width="5" height="5" rx="1" fill="black" opacity="0.6" />
                <rect x="2" y="9" width="5" height="5" rx="1" fill="black" opacity="0.6" />
                <rect x="9" y="9" width="5" height="5" rx="1" fill="black" />
              </svg>
            </div>
            <span className="font-semibold mono text-sm">aipo</span>
          </div>
          <span className="text-xs mono px-1.5 py-0.5 rounded" style={{ background: 'rgba(34,197,94,0.1)', color: 'var(--green)', border: '1px solid rgba(34,197,94,0.2)' }}>
            {online}
          </span>
        </div>

        {/* Agent list */}
        <div style={{ flex: 1, overflowY: 'auto' }}>
          {isLoading && <p className="px-4 py-3 text-xs" style={{ color: 'var(--muted)' }}>加载中...</p>}
          {!isLoading && agents.length === 0 && (
            <div className="px-4 py-6 text-center">
              <Server size={24} style={{ color: 'var(--surface2)', margin: '0 auto 8px' }} />
              <p className="text-xs" style={{ color: 'var(--muted)' }}>暂无 Agent</p>
            </div>
          )}
          {agents.map((a: Agent) => (
            <div key={a.id} className="px-3 py-2.5 transition-colors" style={{ borderBottom: '1px solid var(--border)' }}
              onMouseEnter={e => (e.currentTarget.style.background = 'var(--surface2)')}
              onMouseLeave={e => (e.currentTarget.style.background = 'transparent')}>
              <div className="flex items-center justify-between mb-1">
                <div className="flex items-center gap-1.5 min-w-0">
                  <span className="w-1.5 h-1.5 rounded-full shrink-0" style={{
                    background: a.status === 'online' ? 'var(--green)' : 'var(--muted)',
                    boxShadow: a.status === 'online' ? '0 0 5px var(--green)' : 'none'
                  }} />
                  <span className="text-xs font-medium mono truncate">{a.hostname || '-'}</span>
                </div>
                <button onClick={() => { if (confirm(`删除 ${a.hostname}？`)) deleteMutation.mutate(a.id) }}
                  className="p-0.5 rounded cursor-pointer transition-all shrink-0" style={{ color: 'var(--muted)' }}
                  onMouseEnter={e => { e.currentTarget.style.color = '#F87171' }}
                  onMouseLeave={e => { e.currentTarget.style.color = 'var(--muted)' }}>
                  <Trash2 size={11} />
                </button>
              </div>
              <div className="flex items-center justify-between">
                <span className="text-xs mono truncate" style={{ color: 'var(--muted)' }}>{a.ip || '-'}</span>
                <div className="flex items-center gap-0.5 shrink-0">
                  <button
                    onClick={() => open(a.id, a.hostname || `agent-${a.id}`, 'terminal')}
                    disabled={a.status !== 'online'}
                    title="终端"
                    className="p-1 rounded cursor-pointer transition-all disabled:opacity-25 disabled:cursor-not-allowed"
                    style={{ color: 'var(--muted)' }}
                    onMouseEnter={e => { if (a.status === 'online') { e.currentTarget.style.color = 'var(--green)'; e.currentTarget.style.background = 'rgba(34,197,94,0.1)' } }}
                    onMouseLeave={e => { e.currentTarget.style.color = 'var(--muted)'; e.currentTarget.style.background = 'transparent' }}>
                    <Terminal size={13} />
                  </button>
                  <button
                    onClick={() => open(a.id, a.hostname || `agent-${a.id}`, 'files')}
                    disabled={a.status !== 'online'}
                    title="文件"
                    className="p-1 rounded cursor-pointer transition-all disabled:opacity-25 disabled:cursor-not-allowed"
                    style={{ color: 'var(--muted)' }}
                    onMouseEnter={e => { if (a.status === 'online') { e.currentTarget.style.color = 'var(--green)'; e.currentTarget.style.background = 'rgba(34,197,94,0.1)' } }}
                    onMouseLeave={e => { e.currentTarget.style.color = 'var(--muted)'; e.currentTarget.style.background = 'transparent' }}>
                    <FolderOpen size={13} />
                  </button>
                </div>
              </div>
            </div>
          ))}
        </div>

        {/* Install + Logout */}
        <div className="px-3 py-2 space-y-1" style={{ borderTop: '1px solid var(--border)' }}>
          <button
            onClick={() => setShowInstall(!showInstall)}
            className="w-full flex items-center gap-2 px-3 py-1.5 rounded-lg text-xs font-medium cursor-pointer transition-all"
            style={{ background: showInstall ? 'var(--green)' : 'var(--surface2)', color: showInstall ? '#000' : 'var(--text)', border: '1px solid var(--border)' }}
          >
            <Plus size={12} />安装 Agent
          </button>
          <button onClick={logout}
            className="w-full flex items-center gap-2 px-3 py-1.5 rounded-lg text-xs cursor-pointer transition-all"
            style={{ color: 'var(--muted)' }}
            onMouseEnter={e => { e.currentTarget.style.color = 'var(--text)'; e.currentTarget.style.background = 'var(--surface2)' }}
            onMouseLeave={e => { e.currentTarget.style.color = 'var(--muted)'; e.currentTarget.style.background = 'transparent' }}>
            <LogOut size={12} />退出
          </button>
        </div>
      </aside>

      {/* Main */}
      <div style={{ flex: 1, display: 'flex', flexDirection: 'column', overflow: 'hidden' }}>
        {/* Tab bar */}
        {tabs.length > 0 && (
          <div className="flex items-center overflow-x-auto" style={{ borderBottom: '1px solid var(--border)', background: 'var(--surface)', flexShrink: 0, minHeight: 36 }}>
            {tabs.map(tab => (
              <div
                key={tab.id}
                onClick={() => setActive(tab.id)}
                className="flex items-center gap-1.5 px-3 cursor-pointer transition-all shrink-0"
                style={{
                  height: 36,
                  borderRight: '1px solid var(--border)',
                  background: tab.id === activeId ? 'var(--bg)' : 'transparent',
                  borderBottom: tab.id === activeId ? '1px solid var(--bg)' : '1px solid transparent',
                  marginBottom: tab.id === activeId ? -1 : 0,
                  color: tab.id === activeId ? 'var(--text)' : 'var(--muted)',
                }}
              >
                {tab.type === 'terminal'
                  ? <Terminal size={12} style={{ color: tab.id === activeId ? 'var(--green)' : 'var(--muted)' }} />
                  : <FolderOpen size={12} style={{ color: tab.id === activeId ? 'var(--green)' : 'var(--muted)' }} />
                }
                <span className="text-xs mono max-w-24 truncate">{tab.hostname}</span>
                <button
                  onClick={e => { e.stopPropagation(); close(tab.id) }}
                  className="p-0.5 rounded cursor-pointer transition-colors ml-1"
                  style={{ color: 'var(--muted)' }}
                  onMouseEnter={ev => (ev.currentTarget.style.color = 'var(--text)')}
                  onMouseLeave={ev => (ev.currentTarget.style.color = 'var(--muted)')}>
                  <X size={11} />
                </button>
              </div>
            ))}
          </div>
        )}

        {/* Panel content */}
        <div style={{ flex: 1, position: 'relative', overflow: 'hidden' }}>
          {tabs.length === 0 && (
            <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'center', height: '100%', flexDirection: 'column', gap: 12 }}>
              <Server size={36} style={{ color: 'var(--surface2)' }} />
              <p className="text-sm" style={{ color: 'var(--muted)' }}>从左侧选择主机，点击 <Terminal size={12} style={{ display: 'inline', verticalAlign: 'middle' }} /> 或 <FolderOpen size={12} style={{ display: 'inline', verticalAlign: 'middle' }} /> 打开面板</p>
            </div>
          )}

          {/* Terminal panels — always mounted, hidden when inactive */}
          {tabs.filter(t => t.type === 'terminal').map(tab => (
            <div key={tab.id} style={{ position: 'absolute', inset: 0, display: tab.id === activeId ? 'flex' : 'none' }}>
              <TerminalPanel agentId={tab.agentId} isActive={tab.id === activeId} />
            </div>
          ))}

          {/* File panels — mount/unmount on switch is fine */}
          {activeTab?.type === 'files' && (
            <div style={{ position: 'absolute', inset: 0 }}>
              <FilePanel agentId={activeTab.agentId} />
            </div>
          )}
        </div>
      </div>

      {/* Install modal */}
      {showInstall && (
        <div style={{ position: 'fixed', inset: 0, background: 'rgba(0,0,0,0.6)', display: 'flex', alignItems: 'center', justifyContent: 'center', zIndex: 50 }}
          onClick={() => setShowInstall(false)}>
          <div style={{ background: 'var(--surface)', border: '1px solid var(--border)', borderRadius: 12, padding: 20, width: 480, maxWidth: '90vw' }}
            onClick={e => e.stopPropagation()}>
            <div className="flex items-center justify-between mb-3">
              <p className="text-sm font-medium">在目标服务器上执行以下命令</p>
              <button onClick={() => setShowInstall(false)} className="cursor-pointer" style={{ color: 'var(--muted)' }}><X size={15} /></button>
            </div>
            <div className="relative rounded-lg p-4" style={{ background: 'var(--bg)', border: '1px solid var(--border)' }}>
              <pre className="mono text-xs leading-relaxed overflow-x-auto pr-6" style={{ color: 'var(--green)' }}>{installCmd}</pre>
              <div style={{ position: 'absolute', top: 10, right: 10 }}>
                <CopyButton text={installCmd.replace(/\\\n\s*/g, ' ')} />
              </div>
            </div>
            <p className="text-xs mt-2" style={{ color: 'var(--muted)' }}>Agent 连接后自动注册，无需手动操作。</p>
            <div className="mt-4 pt-3" style={{ borderTop: '1px solid var(--border)' }}>
              <p className="text-xs mb-2" style={{ color: 'var(--muted)' }}>卸载</p>
              <div className="relative rounded-lg p-3" style={{ background: 'var(--bg)', border: '1px solid var(--border)' }}>
                <pre className="mono text-xs" style={{ color: 'var(--muted)' }}>{`sudo bash -c "$(curl -fsSL ${location.origin}/uninstall.sh)"`}</pre>
                <div style={{ position: 'absolute', top: 8, right: 8 }}>
                  <CopyButton text={`sudo bash -c "$(curl -fsSL ${location.origin}/uninstall.sh)"`} />
                </div>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
