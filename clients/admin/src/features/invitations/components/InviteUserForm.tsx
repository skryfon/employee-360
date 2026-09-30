import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { getErrorMessage } from '@employee360/api-client'
import { inviteUserSchema, type InviteUserFormValues } from '../schemas/invitationSchemas'
import { useInviteUserMutation } from '../queries/invitationQueries'
import { useAuthStore } from '../../../stores/authStore'
import { FormField } from '../../auth/components/FormField'
import { SubmitButton } from '../../auth/components/SubmitButton'
import { InlineAlert } from '../../auth/components/InlineAlert'

const EMPTY: InviteUserFormValues = {
  email: '',
  roleId: '',
  firstName: '',
  lastName: '',
  departmentId: '',
  positionId: '',
}

export function InviteUserForm() {
  const mutation = useInviteUserMutation()
  const roleOptions = useAuthStore((s) => s.user?.roleOptions) ?? []
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
      { onSuccess: () => reset(EMPTY) },
    )

  return (
    <form
      onSubmit={handleSubmit(onSubmit)}
      noValidate
      aria-label="Invite user"
      className="flex flex-col gap-4 rounded-sm border border-slate-200 bg-white p-4"
    >
      <h2 className="text-base font-semibold text-slate-900">Invite a user</h2>
      {mutation.isSuccess && <InlineAlert tone="neutral">Invitation sent.</InlineAlert>}
      {mutation.isError && <InlineAlert tone="error">{getErrorMessage(mutation.error, 'Could not send invitation.')}</InlineAlert>}
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
            <option value="">Select a role</option>
            {roleOptions.map((r) => (
              <option key={r.id} value={r.id}>
                {r.name}
              </option>
            ))}
          </select>
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
      <div>
        <SubmitButton loading={mutation.isPending}>Send invitation</SubmitButton>
      </div>
    </form>
  )
}
