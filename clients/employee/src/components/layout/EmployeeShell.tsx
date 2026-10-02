import { Outlet, useNavigate } from 'react-router-dom'
import { BrandMark, UserBadge, pickPrimaryRole } from '@employee360/ui'
import { useAuthStore } from '../../stores/authStore'
import { useLogoutMutation } from '../../features/auth/queries/authMutations'

const FOCUS = 'focus:outline-none focus-visible:ring-2 focus-visible:ring-slate-900 focus-visible:ring-offset-2'

export function EmployeeShell() {
  const user = useAuthStore((s) => s.user)
  const navigate = useNavigate()
  const logout = useLogoutMutation()

  return (
    <div className="flex min-h-dvh flex-col bg-slate-50">
      <header className="sticky top-0 z-20 border-b border-slate-200 bg-white">
       <div className="mx-auto flex h-14 w-full max-w-6xl items-center justify-between gap-4 px-4 sm:px-6 lg:px-8">
        <BrandMark name="Employee360" subtitle="Employee Portal" compactOnMobile />
        <div className="ml-auto flex min-w-0 items-center gap-3">
          <UserBadge email={user?.email} role={pickPrimaryRole(user?.roles)} />
          <span aria-hidden="true" className="h-6 w-px shrink-0 bg-slate-200" />
          <button
            type="button"
            onClick={() => logout.mutate(undefined, { onSettled: () => navigate('/login', { replace: true }) })}
            className={`h-9 shrink-0 rounded-sm border border-slate-300 bg-white px-3 text-sm font-medium text-slate-700 hover:bg-slate-100 ${FOCUS}`}
          >
            Sign out
          </button>
        </div>
       </div>
      </header>
      <main className="mx-auto w-full max-w-6xl flex-1 px-4 py-6 sm:px-6 lg:px-8">
        <Outlet />
      </main>
    </div>
  )
}
