import { useEffect, type ReactNode } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { Link, useSearchParams } from 'react-router-dom'
import { acceptInvitationSchema, type AcceptInvitationFormValues } from '../schemas/authSchemas'
import { useAcceptInvitationMutation, useValidateInvitationQuery } from '../queries/authMutations'
import {
  INVITATION_ERROR_CODES,
  clearSession,
  getErrorCode,
  getErrorMessage,
  resolveInvitationRedirect,
} from '@employee360/api-client'
import { useAuthStore } from '../../../stores/authStore'
import { AuthLayout } from '../components/AuthLayout'
import { FormField } from '../components/FormField'
import { SubmitButton } from '../components/SubmitButton'
import { InlineAlert } from '../components/InlineAlert'

const SIGN_IN_LABEL = 'Sign in to Admin'

type InvitationProblem = {
  title: string
  message: string
  /** Show a primary "Sign in" button (invitation already accepted). */
  signIn?: boolean
}

const INVITATION_PROBLEMS = {
  INVALID_TOKEN: {
    title: 'Invalid invitation link',
    message: 'This invitation link is invalid. Please check the link in your invitation email and try again.',
  },
  EXPIRED: {
    title: 'Invitation expired',
    message: 'This invitation has expired. Ask your administrator to resend the invitation.',
  },
  REVOKED: {
    title: 'Invitation revoked',
    message: 'This invitation has been revoked. Please contact your administrator.',
  },
  ACCEPTED: {
    title: 'Invitation already accepted',
    message:
      'This invitation has already been accepted. You can sign in with your password. If you were invited to the Employee Portal, sign in there instead.',
    signIn: true,
  },
} satisfies Record<string, InvitationProblem>

/** Map a backend error code to a distinct invitation problem state (never by message text). */
function problemForError(error: unknown): InvitationProblem | null {
  switch (getErrorCode(error)) {
    case INVITATION_ERROR_CODES.INVALID_TOKEN:
      return INVITATION_PROBLEMS.INVALID_TOKEN
    case INVITATION_ERROR_CODES.EXPIRED:
      return INVITATION_PROBLEMS.EXPIRED
    case INVITATION_ERROR_CODES.REVOKED:
      return INVITATION_PROBLEMS.REVOKED
    case INVITATION_ERROR_CODES.ACCEPTED:
      return INVITATION_PROBLEMS.ACCEPTED
    default:
      return null
  }
}

function ProblemView({ problem }: { problem: InvitationProblem }) {
  return (
    <AuthLayout title={problem.title}>
      <div className="flex flex-col gap-4">
        <InlineAlert tone="error">{problem.message}</InlineAlert>
        {problem.signIn ? (
          <Link
            to="/login"
            className="inline-flex h-9 items-center justify-center gap-2 rounded-sm bg-slate-900 px-4 text-sm font-semibold text-white hover:bg-slate-800 active:bg-slate-950"
          >
            Sign in
          </Link>
        ) : (
          <Link to="/login" className="text-sm text-slate-900 underline hover:text-slate-700">
            Return to sign in
          </Link>
        )}
      </div>
    </AuthLayout>
  )
}

function Notice({ title, children }: { title: string; children: ReactNode }) {
  return (
    <AuthLayout title={title}>
      <InlineAlert tone="neutral">{children}</InlineAlert>
    </AuthLayout>
  )
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

  const invitation = validateQuery.data
  const redirect =
    token && invitation
      ? resolveInvitationRedirect({
          role: invitation.role,
          currentApp: 'admin',
          token,
          employeeAppUrl: import.meta.env.VITE_EMPLOYEE_APP_URL,
        })
      : null
  const redirectUrl = redirect?.kind === 'redirect' ? redirect.url : null

  useEffect(() => {
    if (redirectUrl) window.location.replace(redirectUrl)
  }, [redirectUrl])

  if (!token) return <ProblemView problem={INVITATION_PROBLEMS.INVALID_TOKEN} />

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
    const problem = problemForError(validateQuery.error)
    if (problem) return <ProblemView problem={problem} />
    return (
      <AuthLayout title="Accept invitation">
        <div className="flex flex-col gap-4">
          <InlineAlert tone="error">
            {getErrorMessage(validateQuery.error, 'Unable to verify this invitation. Please try again.')}
          </InlineAlert>
          <Link to="/login" className="text-sm text-slate-900 underline hover:text-slate-700">
            Return to sign in
          </Link>
        </div>
      </AuthLayout>
    )
  }

  if (redirect?.kind === 'redirect') {
    return (
      <Notice title="Redirecting">
        This invitation is for the Employee Portal. Redirecting you now...{' '}
        <a href={redirect.url} className="underline">
          Continue
        </a>
      </Notice>
    )
  }

  if (redirect?.kind === 'unconfigured') {
    return (
      <Notice title="Open the Employee Portal">
        This invitation is for the Employee Portal; open the link from your invitation email in that app
        (the Employee Portal address is not configured here).
      </Notice>
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
              {SIGN_IN_LABEL}
            </Link>
          </div>
        </div>
      </AuthLayout>
    )
  }

  // A race (e.g. accepted/revoked/expired between validate and submit) gets the same distinct state.
  const acceptProblem = mutation.isError ? problemForError(mutation.error) : null
  if (acceptProblem) return <ProblemView problem={acceptProblem} />

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
            {getErrorMessage(mutation.error, 'Unable to accept invitation. Please try again.')}
          </InlineAlert>
        )}
        <p className="text-xs text-slate-600">
          {invitation?.email ? (
            <>
              Setting a password for <span className="font-semibold text-slate-900">{invitation.email}</span>
            </>
          ) : (
            'Set a password to complete your account setup and accept the invitation.'
          )}
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
