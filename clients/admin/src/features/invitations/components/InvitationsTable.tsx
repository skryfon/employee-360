import { useState } from 'react'
import { getErrorMessage, type Invitation } from '@employee360/api-client'
import { useResendInvitationMutation, useRevokeInvitationMutation } from '../queries/invitationQueries'
import { StatusBadge } from './StatusBadge'
import { ConfirmRevokeModal } from './ConfirmRevokeModal'
import { useToast } from '../../../hooks/useToast'

const fmt = (iso?: string) => (iso ? new Date(iso).toLocaleDateString() : '-')

const secondaryBtn =
  'h-8 rounded-sm border border-slate-300 bg-white px-3 text-xs font-medium text-slate-800 hover:bg-slate-100 focus:outline-none focus-visible:ring-2 focus-visible:ring-slate-900 focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:bg-slate-100 disabled:text-slate-500'
const dangerBtn =
  'h-8 rounded-sm border border-red-300 bg-white px-3 text-xs font-medium text-red-600 hover:bg-red-50 focus:outline-none focus-visible:ring-2 focus-visible:ring-red-600 focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:text-slate-500 disabled:border-slate-200'

export function InvitationsTable({ invitations }: { invitations: Invitation[] }) {
  const toast = useToast()
  const resend = useResendInvitationMutation()
  const revoke = useRevokeInvitationMutation()
  const [revoking, setRevoking] = useState<Invitation | null>(null)

  if (invitations.length === 0) {
    return <p className="p-4 text-sm text-slate-600">No invitations yet.</p>
  }

  return (
    <>
      <table className="w-full text-left">
        <thead className="bg-slate-100 text-xs font-medium text-slate-600">
          <tr>
            <th className="px-4 py-2">Email</th>
            <th className="px-4 py-2">Status</th>
            <th className="px-4 py-2">Sent</th>
            <th className="px-4 py-2">Expires</th>
            <th className="px-4 py-2 text-right">Actions</th>
          </tr>
        </thead>
        <tbody>
          {invitations.map((inv) => {
            const pending = inv.status === 'pending'
            return (
              <tr key={inv.id} className="border-t border-slate-200 text-sm text-slate-900 hover:bg-slate-50">
                <td className="px-4 py-2">{inv.email}</td>
                <td className="px-4 py-2">
                  <StatusBadge status={inv.status} />
                </td>
                <td className="px-4 py-2">{fmt(inv.created_at)}</td>
                <td className="px-4 py-2">{fmt(inv.expires_at)}</td>
                <td className="px-4 py-2">
                  {pending && inv.id && (
                    <div className="flex justify-end gap-2">
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
