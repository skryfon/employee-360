import { Card } from '@employee360/ui'
import type { TenantInfo } from '@employee360/api-client'

const DATE = new Intl.DateTimeFormat('en-GB', { day: 'numeric', month: 'short', year: 'numeric' })

/** "2026-10-01T..." -> "1 Oct 2026"; "-" when missing or unparseable. */
export function formatCreated(iso?: string): string {
  if (!iso) return '-'
  const d = new Date(iso)
  return Number.isNaN(d.getTime()) ? '-' : DATE.format(d)
}

function StatusBadge({ active }: { active: boolean }) {
  return (
    <span
      className={`inline-flex h-6 items-center gap-1.5 rounded-sm border px-2 text-xs font-medium ${
        active ? 'border-green-300 bg-green-50 text-green-700' : 'border-slate-300 bg-slate-100 text-slate-600'
      }`}
    >
      <span aria-hidden="true" className={`h-1.5 w-1.5 rounded-full ${active ? 'bg-green-700' : 'bg-slate-600'}`} />
      {active ? 'Active' : 'Inactive'}
    </span>
  )
}

function Row({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <div className="flex min-w-0 flex-col gap-1 py-3 sm:py-0">
      <dt className="text-xs font-medium text-slate-600">{label}</dt>
      <dd className="min-w-0 text-sm text-slate-900">{children}</dd>
    </div>
  )
}

export function OrganisationCard({ tenant }: { tenant: TenantInfo | undefined }) {
  return (
    <Card role="region" aria-labelledby="org-title" padded={false}>
      <div className="border-b border-slate-200 bg-slate-50 px-4 py-3 sm:px-6">
        <h2 id="org-title" className="text-base font-semibold text-slate-900">Organisation</h2>
      </div>
      <dl className="grid grid-cols-1 divide-y divide-slate-200 px-4 sm:grid-cols-3 sm:gap-6 sm:divide-y-0 sm:px-6 sm:py-5">
        <Row label="Name">
          <span className="break-words font-semibold">{tenant?.name ?? '-'}</span>
        </Row>
        <Row label="Status">
          <StatusBadge active={Boolean(tenant?.is_active)} />
        </Row>
        <Row label="Created">{formatCreated(tenant?.created_at)}</Row>
      </dl>
    </Card>
  )
}
