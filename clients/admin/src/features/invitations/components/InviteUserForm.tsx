import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { getErrorMessage } from '@employee360/api-client'
import { inviteUserSchema, type InviteUserFormValues } from '../schemas/invitationSchemas'
import { useInviteUserMutation, useRolesQuery } from '../queries/invitationQueries'
import { FormField, SubmitButton } from '@employee360/ui'
import { useToast } from '../../../hooks/useToast'

const EMPTY: InviteUserFormValues = {
  email: '',
  roleId: '',
  firstName: '',
  lastName: '',
  departmentId: '',
  positionId: '',
}

export function InviteUserForm({ onSuccess, onCancel }: { onSuccess?: () => void; onCancel?: () => void }) {
  const toast = useToast()
  const mutation = useInviteUserMutation()
  const roles = useRolesQuery()
  const roleOptions = roles.data ?? []
  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<InviteUserFormValues>({ resolver: zodResolver(inviteUserSchema), defaultValues: EMPTY })

  const onSubmit = (v: InviteUserFormValues) =>
    mutation.mutate(
      {
        email: v.email,
        role_id: v.roleId,
        first_name: v.firstName || undefined,
        last_name: v.lastName || undefined,
        department_id: v.departmentId || undefined,
        position_id: v.positionId || undefined,
      },
      {
        onSuccess: () => {
          toast.success(`Invitation sent to ${v.email}.`)
          reset(EMPTY)
          onSuccess?.()
        },
        onError: (err) => toast.error(getErrorMessage(err, 'Could not send invitation.')),
      },
    )

  return (
    <form
      onSubmit={handleSubmit(onSubmit)}
      noValidate
      aria-label="Invite user"
      className="flex flex-col gap-4 rounded-sm border border-slate-200 bg-white p-4"
    >
      <h2 className="text-base font-semibold text-slate-900">Invite a user</h2>
      <div className="grid gap-4 sm:grid-cols-2">
        <FormField label="Email" type="email" error={errors.email?.message} {...register('email')} />
        <div className="flex flex-col gap-1">
          <label htmlFor="roleId" className="text-xs font-medium text-slate-900">
            Role
          </label>
          <select
            id="roleId"
            aria-invalid={errors.roleId ? true : undefined}
            aria-describedby={errors.roleId ? 'roleId-error' : undefined}
            className="h-9 w-full rounded-sm border border-slate-300 bg-white px-3 text-sm text-slate-900 focus:border-slate-900 focus:outline-none focus:ring-1 focus:ring-slate-900 aria-[invalid=true]:border-red-600"
            {...register('roleId')}
          >
            <option value="">{roles.isPending ? 'Loading roles...' : 'Select a role'}</option>
            {roleOptions.map((r) => (
              <option key={r.id} value={r.id}>
                {r.name}
              </option>
            ))}
          </select>
          {roles.isError && (
            <p role="alert" className="text-xs text-red-700">
              Could not load roles.
            </p>
          )}
          {errors.roleId && (
            <p id="roleId-error" role="alert" className="text-xs text-red-700">
              {errors.roleId.message}
            </p>
          )}
        </div>
        <FormField label="First name (optional)" error={errors.firstName?.message} {...register('firstName')} />
        <FormField label="Last name (optional)" error={errors.lastName?.message} {...register('lastName')} />
        <FormField label="Department ID (optional)" error={errors.departmentId?.message} {...register('departmentId')} />
        <FormField label="Position ID (optional)" error={errors.positionId?.message} {...register('positionId')} />
      </div>
      <div className="flex items-center gap-2">
        <SubmitButton loading={mutation.isPending}>Send invitation</SubmitButton>
        {onCancel && (
          <button
            type="button"
            onClick={onCancel}
            className="h-9 rounded-sm border border-slate-300 bg-white px-4 text-sm font-medium text-slate-800 hover:bg-slate-100 active:bg-slate-200 focus:outline-none focus-visible:ring-2 focus-visible:ring-slate-900 focus-visible:ring-offset-2"
          >
            Cancel
          </button>
        )}
      </div>
    </form>
  )
}
