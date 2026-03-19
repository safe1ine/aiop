import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { authApi } from '../api/agents'
import { useAuthStore } from '../store/auth'

export default function Login() {
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)
  const login = useAuthStore((s) => s.login)
  const navigate = useNavigate()

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')
    setLoading(true)
    try {
      const { token } = await authApi.login(username, password)
      login(token)
      navigate('/')
    } catch {
      setError('用户名或密码错误')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="min-h-screen flex items-center justify-center" style={{ background: 'var(--bg)' }}>
      <div className="w-full max-w-sm px-4">
        {/* Logo */}
        <div className="text-center mb-10">
          <div className="inline-flex items-center gap-2 mb-2">
            <div className="w-8 h-8 rounded-lg flex items-center justify-center" style={{ background: 'var(--green)' }}>
              <svg width="16" height="16" viewBox="0 0 16 16" fill="none">
                <rect x="2" y="2" width="5" height="5" rx="1" fill="white" />
                <rect x="9" y="2" width="5" height="5" rx="1" fill="white" opacity="0.6" />
                <rect x="2" y="9" width="5" height="5" rx="1" fill="white" opacity="0.6" />
                <rect x="9" y="9" width="5" height="5" rx="1" fill="white" />
              </svg>
            </div>
            <span className="text-xl font-semibold tracking-tight mono">aipo</span>
          </div>
          <p style={{ color: 'var(--muted)', fontSize: '13px' }}>Linux 远程管理</p>
        </div>

        <form onSubmit={handleSubmit} className="space-y-3">
          {error && (
            <div className="rounded-lg px-4 py-3 text-sm" style={{ background: 'rgba(239,68,68,0.1)', color: '#F87171', border: '1px solid rgba(239,68,68,0.2)' }}>
              {error}
            </div>
          )}
          <div>
            <input
              className="w-full rounded-lg px-4 py-3 text-sm outline-none transition-all"
              style={{ background: 'var(--surface)', border: '1px solid var(--border)', color: 'var(--text)' }}
              onFocus={e => e.target.style.borderColor = 'var(--green)'}
              onBlur={e => e.target.style.borderColor = 'var(--border)'}
              placeholder="用户名"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              autoFocus
            />
          </div>
          <div>
            <input
              type="password"
              className="w-full rounded-lg px-4 py-3 text-sm outline-none transition-all"
              style={{ background: 'var(--surface)', border: '1px solid var(--border)', color: 'var(--text)' }}
              onFocus={e => e.target.style.borderColor = 'var(--green)'}
              onBlur={e => e.target.style.borderColor = 'var(--border)'}
              placeholder="密码"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
            />
          </div>
          <button
            type="submit"
            disabled={loading || !username || !password}
            className="w-full rounded-lg py-3 text-sm font-medium transition-all cursor-pointer"
            style={{ background: loading || !username || !password ? 'var(--surface2)' : 'var(--green)', color: loading || !username || !password ? 'var(--muted)' : '#000' }}
          >
            {loading ? '登录中...' : '登录'}
          </button>
        </form>
      </div>
    </div>
  )
}
