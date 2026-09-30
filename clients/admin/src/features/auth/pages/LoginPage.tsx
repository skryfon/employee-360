import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { Link, Navigate, useLocation, useNavigate } from 'react-router-dom'
import { loginSchema, type LoginFormValues } from '../schemas/authSchemas'
import { useLoginMutation } from '../queries/authMutations'
import { isAuthenticated, useAuthStore } from '../../../stores/authStore'
import { AuthLayout } from '../components/AuthLayout'
import { FormField } from '../components/FormField'
import { SubmitButton } from '../components/SubmitButton'
import { InlineAlert } from '../components/InlineAlert'

const GENERIC_ERROR = 'Invalid email or password.'

export default function LoginPage() {
  const navigate = useNavigate()
  const location = useLocation()
  const authed = useAuthStore(isAuthenticated)
  const mutation = useLoginMutation()
  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<LoginFormValues>({ resolver: zodResolver(loginSchema) })

  const from = (location.state as { from?: string } | null)?.from ?? '/'
  if (authed && !mutation.isPending) return <Navigate to={from} replace />

  const onSubmit = (values: LoginFormValues) =>
    mutation.mutate(values, { onSuccess: () => navigate(from, { replace: true }) })

  return (
    <AuthLayout title="Sign in to Admin">
      <form onSubmit={handleSubmit(onSubmit)} noValidate className="flex flex-col gap-4">
        {mutation.isError && <InlineAlert tone="error">{GENERIC_ERROR}</InlineAlert>}
        <FormField
          label="Email"
          type="email"
          autoComplete="username"
          error={errors.email?.message}
          {...register('email')}
        />
        <FormField
          label="Password"
          type="password"
          autoComplete="current-password"
          error={errors.password?.message}
          {...register('password')}
        />
        <SubmitButton loading={mutation.isPending}>Sign in</SubmitButton>
        <Link to="/forgot-password" className="text-sm text-slate-900 underline hover:text-slate-700">
          Forgot password?
        </Link>
      </form>
    </AuthLayout>
  )
}
