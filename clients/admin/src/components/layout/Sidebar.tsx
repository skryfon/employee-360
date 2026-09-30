import { NavLink, useNavigate } from 'react-router-dom'
import { useAuthStore } from '../../stores/authStore'
import { useUiStore } from '../../stores/uiStore'
import { useLogoutMutation } from '../../features/auth/queries/authMutations'
import { NAV_ITEMS } from './navItems'

const FOCUS = 'focus:outline-none focus-visible:ring-2 focus-visible:ring-slate-900 focus-visible:ring-offset-2'

export function Sidebar() {
  const user = useAuthStore((s) => s.user)
  const collapsed = useUiStore((s) => s.sidebarCollapsed)
  const mobileOpen = useUiStore((s) => s.sidebarMobileOpen)
  const toggle = useUiStore((s) => s.toggleSidebar)
  const setMobileOpen = useUiStore((s) => s.setSidebarMobileOpen)
  const navigate = useNavigate()
  const logout = useLogoutMutation()

  // Labels stay visible in the mobile drawer even if the desktop rail is collapsed.
  const labelClass = collapsed ? 'md:hidden' : ''

  return (
    <>
      {mobileOpen && (
        <div
          data-testid="sidebar-backdrop"
          className="fixed inset-0 z-30 bg-slate-900/50 md:hidden"
          onClick={() => setMobileOpen(false)}
        />
      )}
      <aside
        aria-label="Sidebar"
        data-collapsed={collapsed}
        className={`fixed inset-y-0 left-0 z-40 flex w-56 flex-col border-r border-slate-200 bg-white transition-transform md:static md:translate-x-0 md:transition-none ${
          collapsed ? 'md:w-16' : 'md:w-56'
        } ${mobileOpen ? 'translate-x-0' : '-translate-x-full'}`}
      >
        <div className="flex h-12 items-center justify-between border-b border-slate-200 px-4">
          <span className={`text-base font-semibold text-slate-900 ${labelClass}`}>Employee360 Admin</span>
          <button
            type="button"
            onClick={toggle}
            aria-label={collapsed ? 'Expand sidebar' : 'Collapse sidebar'}
            aria-expanded={!collapsed}
            className={`hidden h-8 w-8 items-center justify-center rounded-sm text-sm text-slate-900 hover:bg-slate-100 md:flex ${FOCUS}`}
          >
            {collapsed ? '»' : '«'}
          </button>
        </div>
        <nav aria-label="Main" className="flex flex-1 flex-col gap-1 p-2">
          {NAV_ITEMS.map((item) => (
            <NavLink
              key={item.to}
              to={item.to}
              title={item.label}
              onClick={() => setMobileOpen(false)}
              className={({ isActive }) =>
                `flex h-9 items-center gap-2 rounded-sm px-3 text-sm ${FOCUS} ${
                  isActive ? 'bg-slate-200 font-semibold text-slate-900' : 'text-slate-900 hover:bg-slate-100'
                }`
              }
            >
              <span aria-hidden="true" className={collapsed ? 'hidden w-6 text-xs font-semibold md:inline' : 'hidden'}>
                {item.short}
              </span>
              <span className={labelClass}>{item.label}</span>
            </NavLink>
          ))}
        </nav>
        <div className="flex flex-col gap-2 border-t border-slate-200 p-2">
          <span className={`truncate px-3 text-xs text-slate-600 ${labelClass}`}>{user?.email}</span>
          <button
            type="button"
            onClick={() => logout.mutate(undefined, { onSettled: () => navigate('/login', { replace: true }) })}
            className={`h-9 rounded-sm border border-slate-300 bg-white px-3 text-sm font-medium text-slate-800 hover:bg-slate-100 ${FOCUS}`}
          >
            <span className={labelClass}>Sign out</span>
            <span aria-hidden="true" className={collapsed ? 'hidden md:inline' : 'hidden'}>
              Out
            </span>
          </button>
        </div>
      </aside>
    </>
  )
}
