import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { useAuthStore } from '../store/auth'
import Login from '../pages/Login'
import Dashboard from '../pages/Dashboard'

function RequireAuth({ children }: { children: React.ReactNode }) {
  const token = useAuthStore((s) => s.token)
  return token ? <>{children}</> : <Navigate to="/login" replace />
}

export default function Router() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/login" element={<Login />} />
        <Route path="/*" element={<RequireAuth><Dashboard /></RequireAuth>} />
      </Routes>
    </BrowserRouter>
  )
}
