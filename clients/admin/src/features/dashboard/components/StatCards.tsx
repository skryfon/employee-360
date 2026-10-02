import { Card } from '@employee360/ui'

export interface Stat {
  label: string
  value: number | undefined
}

export function StatCards({ stats }: { stats: Stat[] }) {
  return (
    <dl className="grid grid-cols-2 gap-4 lg:grid-cols-3">
      {stats.map((s) => (
        <Card key={s.label} className="min-w-0">
          <dt className="text-xs font-medium text-slate-600">{s.label}</dt>
          <dd className="mt-2 text-xl font-semibold text-slate-900">{s.value ?? 0}</dd>
        </Card>
      ))}
    </dl>
  )
}
