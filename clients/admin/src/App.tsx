import { Navigate, Route, Routes } from 'react-router-dom'
import { authRoutes } from './features/auth/routes'
import { RequireAuth } from './components/layout/RequireAuth'
import { AdminShell } from './components/layout/AdminShell'
import DashboardPage from './pages/DashboardPage'

export default function App() {
  return (
    <Routes>
      {authRoutes}
      <Route element={<RequireAuth />}>
        <Route element={<AdminShell />}>
          <Route path="/" element={<DashboardPage />} />
        </Route>
      </Route>
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  )
}
