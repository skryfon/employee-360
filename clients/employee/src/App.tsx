import { Navigate, Route, Routes } from 'react-router-dom'
import { authRoutes } from './features/auth/routes'
import { RequireAuth } from './components/layout/RequireAuth'
import { EmployeeShell } from './components/layout/EmployeeShell'
import DashboardPage from './pages/DashboardPage'

export default function App() {
  return (
    <Routes>
      {authRoutes}
      <Route element={<RequireAuth />}>
        <Route element={<EmployeeShell />}>
          <Route path="/" element={<DashboardPage />} />
        </Route>
      </Route>
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  )
}
