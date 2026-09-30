import { Link, Outlet, useNavigate } from 'react-router-dom'
import { useAuthStore } from '../../stores/authStore'
import { useLogoutMutation } from '../../features/auth/queries/authMutations'

export function AdminShell() {
  const user = useAuthStore((s) => s.user)
  const navigate = useNavigate()
  const logout = useLogoutMutation()
  return (
    <div className="min-h-screen bg-slate-50">
      <header className="flex h-12 items-center justify-between border-b border-slate-200 bg-white px-4">
        <div className="flex items-center gap-6">
          <span className="text-base font-semibold text-slate-900">Employee360 Admin</span>
          <Link to="/invitations" className="text-sm text-slate-900 underline hover:text-slate-700">
            Invitations
          </Link>
        </div>
        <div className="flex items-center gap-4">
          <span className="text-sm text-slate-600">{user?.email}</span>
          <button
            type="button"
            onClick={() => logout.mutate(undefined, { onSettled: () => navigate('/login', { replace: true }) })}
            className="h-9 rounded-sm border border-slate-300 bg-white px-4 text-sm font-medium text-slate-800 hover:bg-slate-100 focus:outline-none focus-visible:ring-2 focus-visible:ring-slate-900 focus-visible:ring-offset-2"
          >
            Sign out
          </button>
        </div>
      </header>
      <main className="p-6">
        <Outlet />
      </main>
    </div>
  )
}
