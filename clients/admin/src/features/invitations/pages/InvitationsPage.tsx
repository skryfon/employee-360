import { useState } from 'react'
import { getErrorMessage } from '@employee360/api-client'
import { useInvitationsQuery } from '../queries/invitationQueries'
import { InviteUserForm } from '../components/InviteUserForm'
import { InvitationsTable } from '../components/InvitationsTable'
import { InlineAlert } from '../../auth/components/InlineAlert'

export default function InvitationsPage() {
  const [page, setPage] = useState(1)
  const { data, isPending, isError, error } = useInvitationsQuery(page)
  const totalPages = data?.meta?.total_pages ?? 1

  return (
    <div className="mx-auto flex max-w-4xl flex-col gap-6">
      <h1 className="text-xl font-semibold text-slate-900">Invitations</h1>
      <InviteUserForm />
      <section className="rounded-sm border border-slate-200 bg-white" aria-label="Invitations list">
        {isPending && <p className="p-4 text-sm text-slate-600">Loading invitations...</p>}
        {isError && (
          <div className="p-4">
            <InlineAlert tone="error">{getErrorMessage(error, 'Could not load invitations.')}</InlineAlert>
          </div>
        )}
        {data && <InvitationsTable invitations={data.data} />}
        {totalPages > 1 && (
          <div className="flex items-center justify-end gap-2 border-t border-slate-200 p-2 text-xs text-slate-600">
            <button
              type="button"
              disabled={page <= 1}
              onClick={() => setPage((p) => p - 1)}
              className="h-8 rounded-sm border border-slate-300 bg-white px-3 font-medium text-slate-800 hover:bg-slate-100 disabled:cursor-not-allowed disabled:bg-slate-100 disabled:text-slate-500"
            >
              Previous
            </button>
            <span>
              Page {page} of {totalPages}
            </span>
            <button
              type="button"
              disabled={page >= totalPages}
              onClick={() => setPage((p) => p + 1)}
              className="h-8 rounded-sm border border-slate-300 bg-white px-3 font-medium text-slate-800 hover:bg-slate-100 disabled:cursor-not-allowed disabled:bg-slate-100 disabled:text-slate-500"
            >
              Next
            </button>
          </div>
        )}
      </section>
    </div>
  )
}
