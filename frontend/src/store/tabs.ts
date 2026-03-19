import { create } from 'zustand'

export type TabType = 'terminal' | 'files'

export interface Tab {
  id: string
  agentId: number
  hostname: string
  type: TabType
}

interface TabStore {
  tabs: Tab[]
  activeId: string | null
  open: (agentId: number, hostname: string, type: TabType) => void
  close: (id: string) => void
  setActive: (id: string) => void
}

export const useTabStore = create<TabStore>((set) => ({
  tabs: [],
  activeId: null,
  open: (agentId, hostname, type) => {
    const id = crypto.randomUUID()
    set(s => ({ tabs: [...s.tabs, { id, agentId, hostname, type }], activeId: id }))
  },
  close: (id) => set(s => {
    const tabs = s.tabs.filter(t => t.id !== id)
    const idx = s.tabs.findIndex(t => t.id === id)
    const activeId = s.activeId === id
      ? (tabs[Math.max(0, idx - 1)]?.id ?? null)
      : s.activeId
    return { tabs, activeId }
  }),
  setActive: (id) => set({ activeId: id }),
}))
