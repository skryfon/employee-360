import { useHealth } from '@employee360/api-client'

function App() {
  const { data, isPending, isError, error } = useHealth()

  return (
    <div className="flex min-h-screen items-center justify-center bg-slate-50">
      <div className="w-full max-w-sm rounded-lg bg-white p-8 text-center shadow">
        <h1 className="text-3xl font-semibold text-slate-900">Employee360 Employee Portal</h1>

        {isPending && <p className="mt-4 text-sm text-slate-500">Checking API health…</p>}

        {isError && (
          <p className="mt-4 text-sm text-red-600">
            API health check failed: {error instanceof Error ? error.message : 'Unknown error'}
          </p>
        )}

        {data && (
          <dl className="mt-6 space-y-2 text-left text-sm text-slate-700">
            <div className="flex items-center justify-between gap-4">
              <dt className="font-medium text-slate-500">Status</dt>
              <dd className="font-mono">{data.status}</dd>
            </div>
            <div className="flex items-center justify-between gap-4">
              <dt className="font-medium text-slate-500">App</dt>
              <dd className="font-mono">{data.app}</dd>
            </div>
            <div className="flex items-center justify-between gap-4">
              <dt className="font-medium text-slate-500">Database</dt>
              <dd className="font-mono">{data.database}</dd>
            </div>
          </dl>
        )}
      </div>
    </div>
  )
}

export default App
