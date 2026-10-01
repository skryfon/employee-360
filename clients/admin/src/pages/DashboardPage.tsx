import { useHealth } from '@employee360/api-client'
import { Card, PageHeader } from '@employee360/ui'

function DashboardPage() {
  const { data, isPending, isError, error } = useHealth()

  return (
    <div className="mx-auto flex w-full max-w-6xl flex-col gap-6">
      <PageHeader title="Employee360 Admin" description="Manage your organization from here." />
      <Card aria-label="System status" className="max-w-md">
        <h2 className="text-base font-semibold text-slate-900">System status</h2>
        {isPending && <p className="mt-4 text-sm text-slate-600">Checking API health…</p>}

        {isError && (
          <p className="mt-4 text-sm text-red-700">
            API health check failed: {error instanceof Error ? error.message : 'Unknown error'}
          </p>
        )}

        {data && (
          <dl className="mt-4 space-y-2 text-sm text-slate-900">
            <div className="flex items-center justify-between gap-4">
              <dt className="font-medium text-slate-600">Status</dt>
              <dd className="font-mono">{data.status}</dd>
            </div>
            <div className="flex items-center justify-between gap-4 border-t border-slate-200 pt-2">
              <dt className="font-medium text-slate-600">App</dt>
              <dd className="font-mono">{data.app}</dd>
            </div>
            <div className="flex items-center justify-between gap-4 border-t border-slate-200 pt-2">
              <dt className="font-medium text-slate-600">Database</dt>
              <dd className="font-mono">{data.database}</dd>
            </div>
          </dl>
        )}
      </Card>
    </div>
  )
}

export default DashboardPage
