import { Outlet, useNavigate } from 'react-router-dom'
import { useAuthStore } from '../../stores/authStore'
import { useLogoutMutation } from '../../features/auth/queries/authMutations'

const FOCUS = 'focus:outline-none focus-visible:ring-2 focus-visible:ring-slate-900 focus-visible:ring-offset-2'

export function EmployeeShell() {
  const user = useAuthStore((s) => s.user)
  const navigate = useNavigate()
  const logout = useLogoutMutation()

  return (
    <div className="flex min-h-screen flex-col bg-slate-50">
      <header className="flex h-14 items-center justify-between border-b border-slate-200 bg-white px-6">
        <div className="flex items-center gap-4">
          <span className="text-lg font-semibold text-slate-900">Employee360 Portal</span>
        </div>
        <div className="ml-auto flex items-center gap-4">
          <span className="text-sm text-slate-600">{user?.email}</span>
          <button
            type="button"
            onClick={() => logout.mutate(undefined, { onSettled: () => navigate('/login', { replace: true }) })}
            className={`h-9 rounded-sm border border-slate-300 bg-white px-4 text-sm font-medium text-slate-800 hover:bg-slate-100 ${FOCUS}`}
          >
            Sign out
          </button>
        </div>
      </header>
      <main className="flex-1 p-6">
        <Outlet />
      </main>
    </div>
  )
}
