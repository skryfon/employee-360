import { Card, Skeleton, SkeletonRegion } from '@employee360/ui'

/** Mirrors DashboardBody: optional Organisation card (super admin) + six stat cards. */
export function DashboardSkeleton({ withOrganisation = false }: { withOrganisation?: boolean }) {
  return (
    <SkeletonRegion label="Loading dashboard…" className="flex flex-col gap-6">
      {withOrganisation && (
        <Card padded={false}>
          <div className="border-b border-slate-200 px-4 py-3 sm:px-6">
            <Skeleton className="h-5 w-28" />
          </div>
          <div className="grid grid-cols-1 gap-4 px-4 py-4 sm:grid-cols-3 sm:gap-6 sm:px-6 sm:py-5">
            {Array.from({ length: 3 }, (_, i) => (
              <div key={i} className="flex flex-col gap-2">
                <Skeleton className="h-3 w-16" />
                <Skeleton className="h-4 w-32" />
              </div>
            ))}
          </div>
        </Card>
      )}
      <div className="grid grid-cols-2 gap-4 lg:grid-cols-3">
        {Array.from({ length: 6 }, (_, i) => (
          <Card key={i}>
            <Skeleton className="h-3 w-20" />
            <Skeleton className="mt-2 h-7 w-12" />
          </Card>
        ))}
      </div>
    </SkeletonRegion>
  )
}
