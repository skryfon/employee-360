import { Outlet, useNavigate } from 'react-router-dom'
import { useAuthStore } from '../../stores/authStore'
import { useUiStore } from '../../stores/uiStore'
import { useLogoutMutation } from '../../features/auth/queries/authMutations'
import { BrandMark, UserBadge } from '@employee360/ui'
import { Sidebar } from './Sidebar'

const FOCUS = 'focus:outline-none focus-visible:ring-2 focus-visible:ring-slate-900 focus-visible:ring-offset-2'

export function AdminShell() {
  const user = useAuthStore((s) => s.user)
  const setMobileOpen = useUiStore((s) => s.setSidebarMobileOpen)
  const mobileOpen = useUiStore((s) => s.sidebarMobileOpen)
  const navigate = useNavigate()
  const logout = useLogoutMutation()
  return (
    <div className="flex min-h-dvh bg-slate-50">
      <Sidebar />
      <div className="flex min-w-0 flex-1 flex-col">
        <header className="sticky flex top-0 z-20 h-14 items-center justify-between gap-2 border-b border-slate-200 bg-white px-3 sm:gap-4 sm:px-4">
          <div className="flex items-center gap-2 sm:gap-4">
            <button
              type="button"
              onClick={() => setMobileOpen(true)}
              aria-label="Open menu"
              aria-expanded={mobileOpen}
              className={`inline-flex h-11 w-11 items-center justify-center rounded-sm border border-slate-300 bg-white text-slate-800 hover:bg-slate-100 md:hidden ${FOCUS}`}
            >
              <svg aria-hidden="true" viewBox="0 0 24 24" width={20} height={20} fill="none" stroke="currentColor" strokeWidth={2} strokeLinecap="round">
                <path d="M4 7h16M4 12h16M4 17h16" />
              </svg>
            </button>
            <span className="md:hidden">
              <BrandMark name="Employee360" subtitle="Admin" compactOnMobile />
            </span>
          </div>
          <div className="ml-auto flex min-w-0 items-center gap-2 sm:gap-4">
            <UserBadge email={user?.email} />
            <button
              type="button"
              onClick={() => logout.mutate(undefined, { onSettled: () => navigate('/login', { replace: true }) })}
              className={`h-11 md:h-9 rounded-sm border border-slate-300 bg-white px-4 text-sm font-medium text-slate-800 hover:bg-slate-100 ${FOCUS}`}
            >
              Sign out
            </button>
          </div>
        </header>
        <main className="min-w-0 flex-1 px-4 py-6 sm:px-6 lg:px-8">
          <Outlet />
        </main>
      </div>
    </div>
  )
}
