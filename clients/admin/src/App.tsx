import { Navigate, Route, Routes } from 'react-router-dom'
import { authRoutes } from './features/auth/routes'
import { RequireAuth } from './components/layout/RequireAuth'
import { AdminShell } from './components/layout/AdminShell'
import { invitationRoutes } from './features/invitations/routes'
import { AppToaster } from './components/feedback/AppToaster'
import { dashboardRoutes } from './features/dashboard/routes'

export default function App() {
  return (
    <>
      <Routes>
        {authRoutes}
        <Route element={<RequireAuth />}>
          <Route element={<AdminShell />}>
            {dashboardRoutes}
            {invitationRoutes}
          </Route>
        </Route>
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
      <AppToaster />
    </>
  )
}
