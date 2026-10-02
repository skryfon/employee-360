import type { AdminDashboard } from '@employee360/api-client'
import { StatCards } from './StatCards'

/** Sections shared by both dashboards; role-specific blocks are passed as `before`. */
export function DashboardBody({ data, before }: { data: AdminDashboard; before?: React.ReactNode }) {
  return (
    <div className="flex flex-col gap-6">
      {before}
      <StatCards
        stats={[
          { label: 'Total users', value: data.users?.total },
          { label: 'Active users', value: data.users?.active },
          { label: 'Inactive users', value: data.users?.inactive },
          { label: 'Pending invited', value: data.users?.pending_invited },
          { label: 'Departments', value: data.departments },
          { label: 'Positions', value: data.positions },
        ]}
      />
    </div>
  )
}
