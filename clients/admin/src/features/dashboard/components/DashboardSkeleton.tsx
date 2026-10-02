import { Card } from '@employee360/ui'

function Bar({ className }: { className: string }) {
  return <div className={`animate-pulse rounded-sm bg-slate-200 ${className}`} />
}

export function DashboardSkeleton() {
  return (
    <div role="status" aria-label="Loading dashboard" className="flex flex-col gap-6">
      <div className="grid grid-cols-2 gap-4 lg:grid-cols-3">
        {Array.from({ length: 6 }, (_, i) => (
          <Card key={i}>
            <Bar className="h-3 w-20" />
            <Bar className="mt-3 h-6 w-12" />
          </Card>
        ))}
      </div>
    </div>
  )
}
