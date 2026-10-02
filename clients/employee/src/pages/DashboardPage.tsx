import { useHealth } from '@employee360/api-client'
import { Card, PageContainer, PageHeader } from '@employee360/ui'
import { useAuthStore } from '../stores/authStore'

export default function DashboardPage() {
  const user = useAuthStore((s) => s.user)
  const { data, isPending, isError, error } = useHealth()

  const displayName = [user?.firstName, user?.lastName].filter(Boolean).join(' ') || user?.email || 'Employee'

  return (
    <PageContainer>
      <PageHeader title={`Welcome, ${displayName}`} description="You are logged into the Employee360 employee portal." />

      <Card>
        <h2 className="text-base font-semibold text-slate-900">System Status</h2>
        {isPending && <p className="mt-2 text-sm text-slate-600">Checking API health…</p>}
        {isError && (
          <p className="mt-2 text-sm text-red-700">
            API health check failed: {error instanceof Error ? error.message : 'Unknown error'}
          </p>
        )}
        {data && (
          <dl className="mt-4 grid grid-cols-1 gap-4 sm:grid-cols-3 text-sm text-slate-900">
            <div className="rounded-sm bg-slate-100 p-4 border border-slate-200">
              <dt className="font-medium text-slate-600">Status</dt>
              <dd className="mt-1 font-mono font-semibold text-slate-900">{data.status}</dd>
            </div>
            <div className="rounded-sm bg-slate-100 p-4 border border-slate-200">
              <dt className="font-medium text-slate-600">Application</dt>
              <dd className="mt-1 font-mono font-semibold text-slate-900">{data.app}</dd>
            </div>
            <div className="rounded-sm bg-slate-100 p-4 border border-slate-200">
              <dt className="font-medium text-slate-600">Database</dt>
              <dd className="mt-1 font-mono font-semibold text-slate-900">{data.database}</dd>
            </div>
          </dl>
        )}
      </Card>
    </PageContainer>
  )
}
