import { useHealth } from '@employee360/api-client'
import { useAuthStore } from '../stores/authStore'

export default function DashboardPage() {
  const user = useAuthStore((s) => s.user)
  const { data, isPending, isError, error } = useHealth()

  const displayName = [user?.firstName, user?.lastName].filter(Boolean).join(' ') || user?.email || 'Employee'

  return (
    <div className="mx-auto max-w-4xl space-y-6">
      <div className="rounded-lg bg-white p-6 shadow-sm border border-slate-200">
        <h1 className="text-2xl font-bold text-slate-900">Welcome, {displayName}</h1>
        <p className="mt-1 text-sm text-slate-600">You are logged into the Employee360 employee portal.</p>
      </div>

      <div className="rounded-lg bg-white p-6 shadow-sm border border-slate-200">
        <h2 className="text-lg font-semibold text-slate-900">System Status</h2>
        {isPending && <p className="mt-2 text-sm text-slate-500">Checking API health…</p>}
        {isError && (
          <p className="mt-2 text-sm text-red-600">
            API health check failed: {error instanceof Error ? error.message : 'Unknown error'}
          </p>
        )}
        {data && (
          <dl className="mt-4 grid grid-cols-1 gap-4 sm:grid-cols-3 text-sm text-slate-700">
            <div className="rounded-md bg-slate-50 p-4 border border-slate-100">
              <dt className="font-medium text-slate-500">Status</dt>
              <dd className="mt-1 font-mono font-semibold text-slate-900">{data.status}</dd>
            </div>
            <div className="rounded-md bg-slate-50 p-4 border border-slate-100">
              <dt className="font-medium text-slate-500">Application</dt>
              <dd className="mt-1 font-mono font-semibold text-slate-900">{data.app}</dd>
            </div>
            <div className="rounded-md bg-slate-50 p-4 border border-slate-100">
              <dt className="font-medium text-slate-500">Database</dt>
              <dd className="mt-1 font-mono font-semibold text-slate-900">{data.database}</dd>
            </div>
          </dl>
        )}
      </div>
    </div>
  )
}
