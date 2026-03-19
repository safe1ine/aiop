import { useEffect, useRef } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { Terminal } from '@xterm/xterm'
import { FitAddon } from '@xterm/addon-fit'
import { useAuthStore } from '../store/auth'
import { ArrowLeft, Terminal as TerminalIcon } from 'lucide-react'
import '@xterm/xterm/css/xterm.css'

export default function TerminalPage() {
  const { agentId } = useParams<{ agentId: string }>()
  const token = useAuthStore((s) => s.token)
  const navigate = useNavigate()
  const containerRef = useRef<HTMLDivElement>(null)
  const termRef = useRef<Terminal | null>(null)
  const wsRef = useRef<WebSocket | null>(null)

  useEffect(() => {
    if (!containerRef.current || !agentId || !token) return

    const term = new Terminal({
      cursorBlink: true,
      fontSize: 13,
      fontFamily: '"Fira Code", "Cascadia Code", Menlo, monospace',
      theme: {
        background: '#020617',
        foreground: '#F8FAFC',
        cursor: '#22C55E',
        cursorAccent: '#020617',
        selectionBackground: 'rgba(34,197,94,0.2)',
        black: '#0F172A',
        brightBlack: '#1E293B',
        red: '#F87171',
        brightRed: '#FCA5A5',
        green: '#22C55E',
        brightGreen: '#4ADE80',
        yellow: '#FBBF24',
        brightYellow: '#FCD34D',
        blue: '#60A5FA',
        brightBlue: '#93C5FD',
        magenta: '#C084FC',
        brightMagenta: '#D8B4FE',
        cyan: '#22D3EE',
        brightCyan: '#67E8F9',
        white: '#CBD5E1',
        brightWhite: '#F8FAFC',
      },
    })
    const fitAddon = new FitAddon()
    term.loadAddon(fitAddon)
    term.open(containerRef.current)
    fitAddon.fit()
    termRef.current = term

    const wsUrl = `${location.protocol === 'https:' ? 'wss' : 'ws'}://${location.host}/api/v1/agents/${agentId}/terminal?token=${token}`
    const ws = new WebSocket(wsUrl)
    wsRef.current = ws
    ws.binaryType = 'arraybuffer'

    ws.onopen = () => {
      ws.send(JSON.stringify({ cols: term.cols, rows: term.rows }))
    }

    ws.onmessage = (e) => {
      const data = e.data instanceof ArrayBuffer ? new Uint8Array(e.data) : e.data
      term.write(data)
    }

    ws.onclose = () => term.write('\r\n\x1b[31m[连接已断开]\x1b[0m\r\n')

    term.onData((data) => ws.readyState === WebSocket.OPEN && ws.send(data))

    const handleResize = () => {
      fitAddon.fit()
      if (ws.readyState === WebSocket.OPEN) {
        ws.send(JSON.stringify({ cols: term.cols, rows: term.rows }))
      }
    }
    window.addEventListener('resize', handleResize)

    return () => {
      window.removeEventListener('resize', handleResize)
      ws.close()
      term.dispose()
    }
  }, [agentId, token])

  return (
    <div className="h-screen flex flex-col" style={{ background: 'var(--bg)' }}>
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
          <div className="flex items-center gap-2">
            <TerminalIcon size={14} style={{ color: 'var(--green)' }} />
            <span className="text-sm mono" style={{ color: 'var(--muted)' }}>
              agent <span style={{ color: 'var(--text)' }}>#{agentId}</span>
            </span>
          </div>
          <div className="ml-auto flex items-center gap-1.5">
            <span className="w-1.5 h-1.5 rounded-full" style={{ background: 'var(--green)', boxShadow: '0 0 6px var(--green)' }} />
            <span className="text-xs mono" style={{ color: 'var(--green)' }}>connected</span>
          </div>
        </div>
      </header>
      <div ref={containerRef} className="flex-1" style={{ padding: '8px 4px 4px 8px' }} />
    </div>
  )
}
