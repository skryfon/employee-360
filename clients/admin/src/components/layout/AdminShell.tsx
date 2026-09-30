import { Outlet } from 'react-router-dom'
import { useUiStore } from '../../stores/uiStore'
import { Sidebar } from './Sidebar'

export function AdminShell() {
  const setMobileOpen = useUiStore((s) => s.setSidebarMobileOpen)
  const mobileOpen = useUiStore((s) => s.sidebarMobileOpen)
  return (
    <div className="flex min-h-screen bg-slate-50">
      <Sidebar />
      <div className="flex min-w-0 flex-1 flex-col">
        <header className="flex h-12 items-center border-b border-slate-200 bg-white px-4 md:hidden">
          <button
            type="button"
            onClick={() => setMobileOpen(true)}
            aria-label="Open menu"
            aria-expanded={mobileOpen}
            className="h-9 rounded-sm border border-slate-300 bg-white px-4 text-sm font-medium text-slate-800 hover:bg-slate-100 focus:outline-none focus-visible:ring-2 focus-visible:ring-slate-900 focus-visible:ring-offset-2"
          >
            Menu
          </button>
          <span className="ml-4 text-base font-semibold text-slate-900">Employee360 Admin</span>
        </header>
        <main className="p-6">
          <Outlet />
        </main>
      </div>
    </div>
  )
}
