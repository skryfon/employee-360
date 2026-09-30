import { Outlet, useNavigate } from 'react-router-dom'
import { useAuthStore } from '../../stores/authStore'
import { useUiStore } from '../../stores/uiStore'
import { useLogoutMutation } from '../../features/auth/queries/authMutations'
import { Sidebar } from './Sidebar'

const FOCUS = 'focus:outline-none focus-visible:ring-2 focus-visible:ring-slate-900 focus-visible:ring-offset-2'

export function AdminShell() {
  const user = useAuthStore((s) => s.user)
  const setMobileOpen = useUiStore((s) => s.setSidebarMobileOpen)
  const mobileOpen = useUiStore((s) => s.sidebarMobileOpen)
  const navigate = useNavigate()
  const logout = useLogoutMutation()
  return (
    <div className="flex min-h-screen bg-slate-50">
      <Sidebar />
      <div className="flex min-w-0 flex-1 flex-col">
        <header className="flex h-12 items-center justify-between gap-4 border-b border-slate-200 bg-white px-4">
          <div className="flex items-center gap-4">
            <button
              type="button"
              onClick={() => setMobileOpen(true)}
              aria-label="Open menu"
              aria-expanded={mobileOpen}
              className={`h-9 rounded-sm border border-slate-300 bg-white px-4 text-sm font-medium text-slate-800 hover:bg-slate-100 md:hidden ${FOCUS}`}
            >
              Menu
            </button>
            <span className="text-base font-semibold text-slate-900 md:hidden">Employee360 Admin</span>
          </div>
          <div className="ml-auto flex items-center gap-4">
            <span className="hidden truncate text-sm text-slate-600 sm:inline">{user?.email}</span>
            <button
              type="button"
              onClick={() => logout.mutate(undefined, { onSettled: () => navigate('/login', { replace: true }) })}
              className={`h-9 rounded-sm border border-slate-300 bg-white px-4 text-sm font-medium text-slate-800 hover:bg-slate-100 ${FOCUS}`}
            >
              Sign out
            </button>
          </div>
        </header>
        <main className="p-6">
          <Outlet />
        </main>
      </div>
    </div>
  )
}
