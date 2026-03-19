import { useEffect, useRef } from 'react'
import { Terminal } from '@xterm/xterm'
import { FitAddon } from '@xterm/addon-fit'
import { useAuthStore } from '../store/auth'
import '@xterm/xterm/css/xterm.css'

interface Props {
  agentId: number
  isActive: boolean
}

export function TerminalPanel({ agentId, isActive }: Props) {
  const token = useAuthStore(s => s.token)
  const containerRef = useRef<HTMLDivElement>(null)
  const fitAddonRef = useRef<FitAddon | null>(null)

  useEffect(() => {
    if (!containerRef.current || !token) return

    const term = new Terminal({
      cursorBlink: true,
      fontSize: 13,
      fontFamily: '"Fira Code", "Cascadia Code", Menlo, monospace',
      theme: {
        background: '#020617', foreground: '#F8FAFC',
        cursor: '#22C55E', cursorAccent: '#020617',
        selectionBackground: 'rgba(34,197,94,0.2)',
        black: '#0F172A', brightBlack: '#1E293B',
        red: '#F87171', brightRed: '#FCA5A5',
        green: '#22C55E', brightGreen: '#4ADE80',
        yellow: '#FBBF24', brightYellow: '#FCD34D',
        blue: '#60A5FA', brightBlue: '#93C5FD',
        magenta: '#C084FC', brightMagenta: '#D8B4FE',
        cyan: '#22D3EE', brightCyan: '#67E8F9',
        white: '#CBD5E1', brightWhite: '#F8FAFC',
      },
    })
    const fitAddon = new FitAddon()
    fitAddonRef.current = fitAddon
    term.loadAddon(fitAddon)
    term.open(containerRef.current)
    fitAddon.fit()

    const wsUrl = `${location.protocol === 'https:' ? 'wss' : 'ws'}://${location.host}/api/v1/agents/${agentId}/terminal?token=${token}`
    const ws = new WebSocket(wsUrl)
    ws.binaryType = 'arraybuffer'
    ws.onopen = () => ws.send(JSON.stringify({ cols: term.cols, rows: term.rows }))
    ws.onmessage = e => term.write(e.data instanceof ArrayBuffer ? new Uint8Array(e.data) : e.data)
    ws.onclose = () => term.write('\r\n\x1b[31m[连接已断开]\x1b[0m\r\n')
    term.onData(data => ws.readyState === WebSocket.OPEN && ws.send(data))

    const handleResize = () => {
      fitAddon.fit()
      if (ws.readyState === WebSocket.OPEN) ws.send(JSON.stringify({ cols: term.cols, rows: term.rows }))
    }
    window.addEventListener('resize', handleResize)

    return () => {
      window.removeEventListener('resize', handleResize)
      ws.close()
      term.dispose()
    }
  }, [agentId, token])

  useEffect(() => {
    if (isActive) {
      const t = setTimeout(() => fitAddonRef.current?.fit(), 50)
      return () => clearTimeout(t)
    }
  }, [isActive])

  return <div ref={containerRef} style={{ width: '100%', height: '100%', padding: '8px 4px 4px 8px' }} />
}
