import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { Link, useSearchParams } from 'react-router-dom'
import { acceptInvitationSchema, type AcceptInvitationFormValues } from '../schemas/authSchemas'
import { useAcceptInvitationMutation, useValidateInvitationQuery } from '../queries/authMutations'
import { clearSession, getErrorMessage } from '@employee360/api-client'
import { useAuthStore } from '../../../stores/authStore'
import { AuthLayout } from '../components/AuthLayout'
import { FormField } from '../components/FormField'
import { SubmitButton } from '../components/SubmitButton'
import { InlineAlert } from '../components/InlineAlert'

function formatAcceptInvitationError(error: unknown): string {
  const message = getErrorMessage(error)
  const lower = message.toLowerCase()
  if (
    lower.includes('token') ||
    lower.includes('invitation') ||
    lower.includes('expire') ||
    lower.includes('revoked') ||
    lower.includes('pending') ||
    lower.includes('not found') ||
    lower.includes('conflict')
  ) {
    return 'This invitation link is invalid, has expired, or has already been accepted. Please contact your organization administrator for a new invitation.'
  }
  return message || 'Unable to accept invitation. Please try again.'
}

export default function AcceptInvitationPage() {
  const [params] = useSearchParams()
  const token = params.get('token')
  const validateQuery = useValidateInvitationQuery(token)
  const mutation = useAcceptInvitationMutation()
  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<AcceptInvitationFormValues>({ resolver: zodResolver(acceptInvitationSchema) })

  if (!token) {
    return (
      <AuthLayout title="Accept invitation">
        <div className="flex flex-col gap-4">
          <InlineAlert tone="error">
            This invitation link is invalid or missing a token. Please check your invitation email.
          </InlineAlert>
          <Link to="/login" className="text-sm text-slate-900 underline hover:text-slate-700">
            Return to sign in
          </Link>
        </div>
      </AuthLayout>
    )
  }

  if (validateQuery.isLoading) {
    return (
      <AuthLayout title="Verifying invitation">
        <div className="flex flex-col items-center justify-center gap-3 py-6" role="status" aria-busy="true">
          <svg className="h-6 w-6 animate-spin text-slate-900" viewBox="0 0 24 24" fill="none" aria-hidden="true">
            <circle cx="12" cy="12" r="9" stroke="currentColor" strokeWidth="3" opacity="0.25" />
            <path d="M21 12a9 9 0 0 0-9-9" stroke="currentColor" strokeWidth="3" />
          </svg>
          <p className="text-sm text-slate-600">Verifying your invitation link...</p>
        </div>
      </AuthLayout>
    )
  }

  if (validateQuery.isError) {
    return (
      <AuthLayout title="Accept invitation">
        <div className="flex flex-col gap-4">
          <InlineAlert tone="error">
            {formatAcceptInvitationError(validateQuery.error)}
          </InlineAlert>
          <Link to="/login" className="text-sm text-slate-900 underline hover:text-slate-700">
            Return to sign in
          </Link>
        </div>
      </AuthLayout>
    )
  }

  if (mutation.isSuccess) {
    return (
      <AuthLayout title="Invitation accepted">
        <div className="flex flex-col gap-4">
          <InlineAlert tone="neutral">
            Your invitation has been accepted and your password has been set. You can now sign in to your account.
          </InlineAlert>
          <div className="flex flex-col gap-2 pt-2">
            <Link
              to="/login"
              state={{ invitationAccepted: true }}
              className="inline-flex h-9 items-center justify-center gap-2 rounded-sm bg-slate-900 px-4 text-sm font-semibold text-white hover:bg-slate-800 active:bg-slate-950"
            >
              Sign in to Admin
            </Link>
          </div>
        </div>
      </AuthLayout>
    )
  }

  const onSubmit = (v: AcceptInvitationFormValues) => {
    mutation.mutate(
      { token, password: v.password },
      {
        onSuccess: () => {
          clearSession()
          useAuthStore.getState().clear()
        },
      },
    )
  }

  return (
    <AuthLayout title="Set your password">
      <form onSubmit={handleSubmit(onSubmit)} noValidate className="flex flex-col gap-4">
        {mutation.isError && (
          <InlineAlert tone="error">
            {formatAcceptInvitationError(mutation.error)}
          </InlineAlert>
        )}
        <p className="text-xs text-slate-600">
          Set a password to complete your account setup and accept the invitation.
        </p>
        <FormField
          label="Password"
          type="password"
          autoComplete="new-password"
          error={errors.password?.message}
          {...register('password')}
        />
        <FormField
          label="Confirm password"
          type="password"
          autoComplete="new-password"
          error={errors.confirmPassword?.message}
          {...register('confirmPassword')}
        />
        <SubmitButton loading={mutation.isPending}>Set password & accept</SubmitButton>
        <Link to="/login" className="text-sm text-slate-900 underline hover:text-slate-700">
          Already accepted? Sign in
        </Link>
      </form>
    </AuthLayout>
  )
}
