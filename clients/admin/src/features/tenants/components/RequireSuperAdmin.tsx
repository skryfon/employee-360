import { Navigate, Outlet } from 'react-router-dom'
import { useAuthStore } from '../../../stores/authStore'

/** UX guard only; the backend enforces 403 for non-super_admin callers. */
export function RequireSuperAdmin() {
  const isSuperAdmin = useAuthStore((s) => Boolean(s.user?.roles.includes('super_admin')))
  if (!isSuperAdmin) return <Navigate to="/" replace />
  return <Outlet />
}
