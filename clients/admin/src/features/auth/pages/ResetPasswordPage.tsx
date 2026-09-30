import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { Link, useSearchParams } from 'react-router-dom'
import { resetPasswordSchema, type ResetPasswordFormValues } from '../schemas/authSchemas'
import { useResetPasswordMutation } from '../queries/authMutations'
import { AuthLayout } from '../components/AuthLayout'
import { FormField } from '../components/FormField'
import { SubmitButton } from '../components/SubmitButton'
import { InlineAlert } from '../components/InlineAlert'

export default function ResetPasswordPage() {
  const [params] = useSearchParams()
  const token = params.get('token')
  const mutation = useResetPasswordMutation()
  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<ResetPasswordFormValues>({ resolver: zodResolver(resetPasswordSchema) })

  if (!token) {
    return (
      <AuthLayout title="Reset password">
        <div className="flex flex-col gap-4">
          <InlineAlert tone="error">This reset link is invalid. Request a new one.</InlineAlert>
          <Link to="/forgot-password" className="text-sm text-slate-900 underline hover:text-slate-700">
            Request a new link
          </Link>
        </div>
      </AuthLayout>
    )
  }

  if (mutation.isSuccess) {
    return (
      <AuthLayout title="Reset password">
        <div className="flex flex-col gap-4">
          <InlineAlert tone="neutral">Your password has been reset. You can now sign in.</InlineAlert>
          <Link to="/login" className="text-sm text-slate-900 underline hover:text-slate-700">
            Go to sign in
          </Link>
        </div>
      </AuthLayout>
    )
  }

  const onSubmit = (v: ResetPasswordFormValues) =>
    mutation.mutate({ token, new_password: v.newPassword })

  return (
    <AuthLayout title="Reset password">
      <form onSubmit={handleSubmit(onSubmit)} noValidate className="flex flex-col gap-4">
        {mutation.isError && (
          <InlineAlert tone="error">
            This reset link is invalid or has expired. Request a new one.
          </InlineAlert>
        )}
        <FormField
          label="New password"
          type="password"
          autoComplete="new-password"
          error={errors.newPassword?.message}
          {...register('newPassword')}
        />
        <FormField
          label="Confirm new password"
          type="password"
          autoComplete="new-password"
          error={errors.confirmPassword?.message}
          {...register('confirmPassword')}
        />
        <SubmitButton loading={mutation.isPending}>Reset password</SubmitButton>
        <Link to="/forgot-password" className="text-sm text-slate-900 underline hover:text-slate-700">
          Request a new link
        </Link>
      </form>
    </AuthLayout>
  )
}
