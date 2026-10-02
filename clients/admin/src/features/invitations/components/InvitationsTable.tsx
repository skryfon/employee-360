import { useState } from 'react'
import { getErrorMessage, type InvitationListItem } from '@employee360/api-client'
import { useResendInvitationMutation, useRevokeInvitationMutation } from '../queries/invitationQueries'
import { StatusBadge } from './StatusBadge'
import { ConfirmRevokeModal } from './ConfirmRevokeModal'
import { useToast } from '../../../hooks/useToast'

const CELL =
  "px-4 py-3 max-sm:flex max-sm:items-center max-sm:justify-between max-sm:p-0 max-sm:before:text-xs max-sm:before:font-medium max-sm:before:text-slate-600 max-sm:before:content-[attr(data-label)]"

const fmt = (iso?: string) => (iso ? new Date(iso).toLocaleDateString() : '-')

const secondaryBtn =
  'h-11 md:h-8 rounded-sm border border-slate-300 bg-white px-3 text-xs font-medium text-slate-800 hover:bg-slate-100 focus:outline-none focus-visible:ring-2 focus-visible:ring-slate-900 focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:bg-slate-100 disabled:text-slate-500'
const dangerBtn =
  'h-11 md:h-8 rounded-sm border border-red-300 bg-white px-3 text-xs font-medium text-red-600 hover:bg-red-50 focus:outline-none focus-visible:ring-2 focus-visible:ring-red-600 focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:text-slate-500 disabled:border-slate-200'

export function InvitationsTable({ invitations, filtered = false }: { invitations: InvitationListItem[]; filtered?: boolean }) {
  const toast = useToast()
  const resend = useResendInvitationMutation()
  const revoke = useRevokeInvitationMutation()
  const [revoking, setRevoking] = useState<InvitationListItem | null>(null)

  if (invitations.length === 0) {
    return (
      <div className="flex flex-col items-center gap-1 px-4 py-8 text-center">
        <p className="text-sm font-medium text-slate-900">
          {filtered ? 'No invitations match your filters' : 'No invitations yet'}
        </p>
        <p className="text-sm text-slate-600">
          {filtered ? 'Try a different status or search term.' : 'Create an invitation to onboard your first user.'}
        </p>
      </div>
    )
  }

  return (
    <>
      <div className="overflow-x-auto">
      <table className="w-full text-left max-sm:block sm:min-w-[64rem]">
        <thead className="bg-slate-100 text-xs font-medium text-slate-600 max-sm:sr-only">
          <tr>
            <th scope="col" className="px-4 py-3">Email</th>
            <th scope="col" className="px-4 py-3">Role</th>
            <th scope="col" className="px-4 py-3">Invited by</th>
            <th scope="col" className="px-4 py-3">Invited on</th>
            <th scope="col" className="px-4 py-3">Expires on</th>
            <th scope="col" className="px-4 py-3">Status</th>
            <th scope="col" className="px-4 py-3 text-right">Actions</th>
          </tr>
        </thead>
        <tbody className="max-sm:block">
          {invitations.map((inv) => {
            const pending = inv.status === 'pending'
            return (
              <tr key={inv.id} className="border-t border-slate-200 text-sm text-slate-900 hover:bg-slate-50 max-sm:flex max-sm:flex-col max-sm:gap-2 max-sm:p-4">
                <td className="px-4 py-3 max-sm:p-0 max-sm:break-all max-sm:font-medium">{inv.email}</td>
                <td className={CELL} data-label="Role">{inv.role ?? '-'}</td>
                <td className={CELL} data-label="Invited by">{inv.invited_by?.name || inv.invited_by?.email || '-'}</td>
                <td className={CELL} data-label="Invited on">{fmt(inv.invited_on ?? inv.created_at)}</td>
                <td className={CELL} data-label="Expires on">{fmt(inv.expires_at)}</td>
                <td className={CELL} data-label="Status">
                  <StatusBadge status={inv.status} />
                </td>
                <td className="px-4 py-3 max-sm:p-0 max-sm:pt-1 max-sm:empty:hidden">
                  {pending && inv.id && (
                    <div className="flex gap-2 sm:justify-end max-sm:[&>button]:flex-1">
                      <button
                        type="button"
                        className={secondaryBtn}
                        disabled={resend.isPending}
                        onClick={() =>
                          resend.mutate(inv.id as string, {
                            onSuccess: () => toast.success(`Invitation resent to ${inv.email}.`),
                            onError: (err) => toast.error(getErrorMessage(err, 'Could not resend invitation.')),
                          })
                        }
                        aria-label={`Resend invitation to ${inv.email}`}
                      >
                        Resend
                      </button>
                      <button
                        type="button"
                        className={dangerBtn}
                        onClick={() => {
                          revoke.reset()
                          setRevoking(inv)
                        }}
                        aria-label={`Revoke invitation for ${inv.email}`}
                      >
                        Revoke
                      </button>
                    </div>
                  )}
                </td>
              </tr>
            )
          })}
        </tbody>
      </table>
      </div>
      {revoking && (
        <ConfirmRevokeModal
          email={revoking.email ?? ''}
          loading={revoke.isPending}
          error={revoke.isError ? getErrorMessage(revoke.error, 'Could not revoke invitation.') : undefined}
          onCancel={() => setRevoking(null)}
          onConfirm={() =>
            revoke.mutate(revoking.id as string, {
              onSuccess: () => {
                toast.success(`Invitation for ${revoking.email} revoked.`)
                setRevoking(null)
              },
              onError: (err) => toast.error(getErrorMessage(err, 'Could not revoke invitation.')),
            })
          }
        />
      )}
    </>
  )
}
