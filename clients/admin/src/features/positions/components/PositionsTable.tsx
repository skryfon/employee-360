import { getErrorCode, getErrorMessage } from '@employee360/api-client'
import { useState } from 'react'
import { useDeletePositionMutation, type Position } from '../queries/positionQueries'
import { ConfirmDeletePositionModal } from './ConfirmDeletePositionModal'
import { PositionStatusBadge } from './PositionStatusBadge'
import { PositionFormModal } from './PositionFormModal'
import { useToast } from '../../../hooks/useToast'

const CELL =
  'px-4 py-3 max-sm:flex max-sm:items-start max-sm:justify-between max-sm:gap-4 max-sm:p-0 max-sm:before:shrink-0 max-sm:before:text-xs max-sm:before:font-medium max-sm:before:text-slate-600 max-sm:before:content-[attr(data-label)]'

const fmt = (iso?: string) => (iso ? new Date(iso).toLocaleDateString() : '-')

const secondaryBtn =
  'h-11 md:h-8 rounded-sm border border-slate-300 bg-white px-3 text-xs font-medium text-slate-800 hover:bg-slate-100 focus:outline-none focus-visible:ring-2 focus-visible:ring-slate-900 focus-visible:ring-offset-2'
const dangerBtn =
  'h-11 md:h-8 rounded-sm border border-red-300 bg-white px-3 text-xs font-medium text-red-600 hover:bg-red-50 focus:outline-none focus-visible:ring-2 focus-visible:ring-red-600 focus-visible:ring-offset-2'

export const IN_USE_MESSAGE =
  'This position is still assigned to users or pending invitations. Reassign them before deleting it.'

export function PositionsTable({ positions, filtered = false }: { positions: Position[]; filtered?: boolean }) {
  const toast = useToast()
  const del = useDeletePositionMutation()
  const [editing, setEditing] = useState<Position | null>(null)
  const [deleting, setDeleting] = useState<Position | null>(null)

  if (positions.length === 0) {
    return (
      <div className="flex flex-col items-center gap-1 px-4 py-8 text-center">
        <p className="text-sm font-medium text-slate-900">{filtered ? 'No matching positions' : 'No positions yet'}</p>
        <p className="text-sm text-slate-600">
          {filtered ? 'Try a different status filter.' : 'Create a position to organise your people.'}
        </p>
      </div>
    )
  }

  // The backend returns CONFLICT for any 409; on delete that can only mean "in use".
  const deleteError = del.isError
    ? getErrorCode(del.error) === 'CONFLICT'
      ? IN_USE_MESSAGE
      : getErrorMessage(del.error, 'Could not delete position.')
    : undefined

  return (
    <>
      <div className="overflow-x-auto">
        <table className="w-full text-left max-sm:block sm:min-w-[46rem]">
          <thead className="bg-slate-100 text-xs font-medium text-slate-600 max-sm:sr-only">
            <tr>
              <th scope="col" className="px-4 py-3">Name</th>
              <th scope="col" className="px-4 py-3">Description</th>
              <th scope="col" className="px-4 py-3">Status</th>
              <th scope="col" className="px-4 py-3">Updated</th>
              <th scope="col" className="px-4 py-3 text-right">Actions</th>
            </tr>
          </thead>
          <tbody className="max-sm:block">
            {positions.map((d) => (
              <tr
                key={d.id}
                className="border-t border-slate-200 text-sm text-slate-900 hover:bg-slate-50 max-sm:flex max-sm:flex-col max-sm:gap-2 max-sm:p-4"
              >
                <td className="px-4 py-3 font-medium max-sm:p-0 max-sm:break-words">{d.name}</td>
                <td className={`${CELL} text-slate-600 sm:max-w-md sm:break-words`} data-label="Description">
                  {d.description || '-'}
                </td>
                <td className={CELL} data-label="Status">
                  <PositionStatusBadge active={d.is_active} />
                </td>
                <td className={CELL} data-label="Updated">{fmt(d.updated_at)}</td>
                <td className="px-4 py-3 max-sm:p-0 max-sm:pt-1">
                  <div className="flex gap-2 sm:justify-end max-sm:[&>button]:flex-1">
                    <button type="button" className={secondaryBtn} aria-label={`Edit position ${d.name}`} onClick={() => setEditing(d)}>
                      Edit
                    </button>
                    <button
                      type="button"
                      className={dangerBtn}
                      aria-label={`Delete position ${d.name}`}
                      onClick={() => {
                        del.reset()
                        setDeleting(d)
                      }}
                    >
                      Delete
                    </button>
                  </div>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
      {editing && <PositionFormModal position={editing} onClose={() => setEditing(null)} />}
      {deleting && (
        <ConfirmDeletePositionModal
          name={deleting.name ?? ''}
          loading={del.isPending}
          error={deleteError}
          onCancel={() => setDeleting(null)}
          onConfirm={() =>
            del.mutate(deleting.id as string, {
              onSuccess: () => {
                toast.success(`Position "${deleting.name}" deleted.`)
                setDeleting(null)
              },
              onError: (err) => {
                if (getErrorCode(err) !== 'CONFLICT') toast.error(getErrorMessage(err, 'Could not delete position.'))
              },
            })
          }
        />
      )}
    </>
  )
}
