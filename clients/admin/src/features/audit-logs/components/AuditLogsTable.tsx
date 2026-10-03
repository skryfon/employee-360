import { Fragment, useState } from 'react'
import { actionLabel, entityTypeLabel } from '../schemas/auditLogSchemas'
import type { AuditLogEntry } from '../queries/auditLogQueries'

const secondaryBtn =
  'h-11 md:h-8 rounded-sm border border-slate-300 bg-white px-3 text-xs font-medium text-slate-800 hover:bg-slate-100 focus:outline-none focus-visible:ring-2 focus-visible:ring-slate-900 focus-visible:ring-offset-2'

const fmtTime = (iso?: string) => {
  if (!iso) return '-'
  const d = new Date(iso)
  return Number.isNaN(d.getTime()) ? '-' : d.toLocaleString(undefined, { dateStyle: 'medium', timeStyle: 'medium' })
}

function Actor({ actor }: { actor: AuditLogEntry['actor'] }) {
  if (!actor) return <span className="text-slate-600">System</span>
  return (
    <div className="flex flex-col break-words">
      <span className="font-medium">{actor.name || actor.email}</span>
      {actor.name && <span className="text-xs text-slate-600">{actor.email}</span>}
    </div>
  )
}

/** Metadata is rendered strictly as text (React-escaped) inside a <pre>. */
function Metadata({ metadata }: { metadata: AuditLogEntry['metadata'] }) {
  const empty = !metadata || Object.keys(metadata).length === 0
  if (empty) return <p className="text-sm text-slate-600">No additional details.</p>
  return (
    <pre className="max-h-72 overflow-auto whitespace-pre-wrap break-words rounded-sm border border-slate-200 bg-white p-3 text-xs text-slate-900">
      {JSON.stringify(metadata, null, 2)}
    </pre>
  )
}

export function AuditLogsTable({ entries, filtered = false }: { entries: AuditLogEntry[]; filtered?: boolean }) {
  const [open, setOpen] = useState<Set<string>>(new Set())
  const toggle = (id: string) =>
    setOpen((prev) => {
      const next = new Set(prev)
      if (!next.delete(id)) next.add(id)
      return next
    })

  if (entries.length === 0) {
    return (
      <div className="flex flex-col items-center gap-1 px-4 py-8 text-center">
        <p className="text-sm font-medium text-slate-900">{filtered ? 'No matching entries' : 'No audit entries yet'}</p>
        <p className="text-sm text-slate-600">
          {filtered ? 'Try different filters or a wider date range.' : 'Administrative changes will appear here.'}
        </p>
      </div>
    )
  }

  return (
    <div className="overflow-x-auto">
      <table className="w-full min-w-[52rem] text-left">
        <thead className="bg-slate-100 text-xs font-medium text-slate-600">
          <tr>
            <th scope="col" className="px-4 py-3">Time</th>
            <th scope="col" className="px-4 py-3">Actor</th>
            <th scope="col" className="px-4 py-3">Action</th>
            <th scope="col" className="px-4 py-3">Entity</th>
            <th scope="col" className="px-4 py-3 text-right">Details</th>
          </tr>
        </thead>
        <tbody>
          {entries.map((e, i) => {
            const id = e.id ?? String(i)
            const isOpen = open.has(id)
            const panelId = `audit-details-${id}`
            return (
              <Fragment key={id}>
                <tr className="border-t border-slate-200 align-top text-sm text-slate-900 hover:bg-slate-50">
                  <td className="whitespace-nowrap px-4 py-3">{fmtTime(e.created_at)}</td>
                  <td className="px-4 py-3"><Actor actor={e.actor} /></td>
                  <td className="px-4 py-3">
                    <div className="flex flex-col">
                      <span className="font-medium" title={e.action}>{actionLabel(e.action)}</span>
                      <code className="break-all text-xs text-slate-600">{e.action}</code>
                    </div>
                  </td>
                  <td className="px-4 py-3">
                    <div className="flex flex-col">
                      <span>{entityTypeLabel(e.entity_type)}</span>
                      {e.entity_id && (
                        <code className="text-xs text-slate-600" title={e.entity_id}>{e.entity_id.slice(0, 8)}</code>
                      )}
                    </div>
                  </td>
                  <td className="px-4 py-3 text-right">
                    <button
                      type="button"
                      className={secondaryBtn}
                      aria-expanded={isOpen}
                      aria-controls={panelId}
                      aria-label={`${isOpen ? 'Hide' : 'Show'} details for ${actionLabel(e.action)}`}
                      onClick={() => toggle(id)}
                    >
                      {isOpen ? 'Hide' : 'Details'}
                    </button>
                  </td>
                </tr>
                {isOpen && (
                  <tr id={panelId} className="bg-slate-50">
                    <td colSpan={5} className="px-4 pb-4 pt-1">
                      <Metadata metadata={e.metadata} />
                    </td>
                  </tr>
                )}
              </Fragment>
            )
          })}
        </tbody>
      </table>
    </div>
  )
}
