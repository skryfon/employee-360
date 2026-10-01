import { afterEach, beforeEach, describe, expect, it } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import MockAdapter from 'axios-mock-adapter'
import { apiClient, clearSession, getSession, setSession } from '@employee360/api-client'
import App from '../../../App'
import { FORGOT_SUCCESS_MESSAGE } from './ForgotPasswordPage'
import { useAuthStore } from '../../../stores/authStore'

let mock: MockAdapter

function renderAt(path: string) {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false }, mutations: { retry: false } } })
  return render(
    <QueryClientProvider client={qc}>
      <MemoryRouter initialEntries={[path]}>
        <App />
      </MemoryRouter>
    </QueryClientProvider>,
  )
}

const loginOk = (role = 'employee') => ({
  success: true,
  data: {
    access_token: 'emp_access_tok',
    refresh_token: 'emp_refresh_tok',
    user: { id: 'u1', email: 'emp@x.com', first_name: 'Jane', last_name: 'Doe', tenant_id: 't1', roles: [{ name: role }] },
  },
})

beforeEach(() => {
  mock = new MockAdapter(apiClient)
  mock.onGet('/api/v1/health').reply(200, { success: true, data: { status: 'ok', app: 'employee360', database: 'ok' } })
  clearSession()
  useAuthStore.getState().clear()
  localStorage.clear()
})
afterEach(() => mock.restore())

describe('employee login', () => {
  it('redirects unauthenticated users to login', () => {
    renderAt('/')
    expect(screen.getByRole('heading', { name: /sign in to employee portal/i })).toBeInTheDocument()
  })

  it('signs in and lands in the employee shell', async () => {
    mock.onPost('/api/v1/auth/login').reply(200, loginOk('employee'))
    renderAt('/login')
    await userEvent.type(screen.getByLabelText('Email'), 'emp@x.com')
    await userEvent.type(screen.getByLabelText('Password'), 'secret123')
    await userEvent.click(screen.getByRole('button', { name: 'Sign in' }))
    expect(await screen.findByRole('button', { name: 'Sign out' })).toBeInTheDocument()
    expect(screen.getByText(/welcome, Jane Doe/i)).toBeInTheDocument()
    expect(getSession()?.accessToken).toBe('emp_access_tok')
    expect(useAuthStore.getState().user?.roles).toEqual(['employee'])
  })

  it('shows inline error on 401 and retains entered email allowing retry', async () => {
    mock.onPost('/api/v1/auth/login').reply(401, { error: { message: 'invalid credentials' } })
    renderAt('/login')
    const emailInput = screen.getByLabelText('Email')
    const passwordInput = screen.getByLabelText('Password')
    await userEvent.type(emailInput, 'emp@x.com')
    await userEvent.type(passwordInput, 'wrongpw')
    await userEvent.click(screen.getByRole('button', { name: 'Sign in' }))

    expect(await screen.findByRole('alert')).toHaveTextContent('Invalid email or password.')
    // Email is preserved in the input for easy retry
    expect(emailInput).toHaveValue('emp@x.com')

    // Retry with valid credentials
    mock.onPost('/api/v1/auth/login').reply(200, loginOk('employee'))
    await userEvent.clear(passwordInput)
    await userEvent.type(passwordInput, 'correctpw')
    await userEvent.click(screen.getByRole('button', { name: 'Sign in' }))

    expect(await screen.findByRole('button', { name: 'Sign out' })).toBeInTheDocument()
  })

  it('shows inline error on 403 without losing email', async () => {
    mock.onPost('/api/v1/auth/login').reply(403, {})
    renderAt('/login')
    const emailInput = screen.getByLabelText('Email')
    await userEvent.type(emailInput, 'emp@x.com')
    await userEvent.type(screen.getByLabelText('Password'), 'pw')
    await userEvent.click(screen.getByRole('button', { name: 'Sign in' }))

    expect(await screen.findByRole('alert')).toHaveTextContent('Invalid email or password.')
    expect(emailInput).toHaveValue('emp@x.com')
  })

  it('rejects accounts with empty/unknown roles with generic error', async () => {
    mock.onPost('/api/v1/auth/login').reply(200, {
      success: true,
      data: {
        access_token: 'a',
        refresh_token: 'r',
        user: { id: 'u1', email: 'emp@x.com', roles: [{ name: 'guest' }] },
      },
    })
    mock.onPost('/api/v1/auth/logout').reply(200, {})
    renderAt('/login')
    await userEvent.type(screen.getByLabelText('Email'), 'emp@x.com')
    await userEvent.type(screen.getByLabelText('Password'), 'pw')
    await userEvent.click(screen.getByRole('button', { name: 'Sign in' }))
    expect(await screen.findByRole('alert')).toHaveTextContent('Invalid email or password.')
    expect(useAuthStore.getState().user).toBeNull()
    expect(getSession()).toBeNull()
  })

  it.each([
    ['500', (m: MockAdapter) => m.onPost('/api/v1/auth/logout').reply(500, {})],
    ['network error', (m: MockAdapter) => m.onPost('/api/v1/auth/logout').networkError()],
  ])('clears the session for non-employee roles even when logout fails (%s)', async (_label, failLogout) => {
    mock.onPost('/api/v1/auth/login').reply(200, {
      success: true,
      data: {
        access_token: 'a',
        refresh_token: 'r',
        user: { id: 'u1', email: 'emp@x.com', roles: [{ name: 'guest' }] },
      },
    })
    failLogout(mock)
    renderAt('/login')
    await userEvent.type(screen.getByLabelText('Email'), 'emp@x.com')
    await userEvent.type(screen.getByLabelText('Password'), 'pw')
    await userEvent.click(screen.getByRole('button', { name: 'Sign in' }))
    expect(await screen.findByRole('alert')).toHaveTextContent('Invalid email or password.')
    expect(getSession()).toBeNull()
    expect(useAuthStore.getState().user).toBeNull()
    expect(useAuthStore.getState().accessToken).toBeNull()
  })
})

describe('employee login error states', () => {
  it.each([500, 429])('shows distinct error message for status %i', async (status) => {
    mock.onPost('/api/v1/auth/login').reply(status, {})
    renderAt('/login')
    await userEvent.type(screen.getByLabelText('Email'), 'emp@x.com')
    await userEvent.type(screen.getByLabelText('Password'), 'pw')
    await userEvent.click(screen.getByRole('button', { name: 'Sign in' }))
    expect(await screen.findByRole('alert')).toHaveTextContent('Unable to sign in. Please try again.')
    expect(screen.getByLabelText('Email')).toHaveValue('emp@x.com')
  })

  it('shows distinct error message on network failure', async () => {
    mock.onPost('/api/v1/auth/login').networkError()
    renderAt('/login')
    await userEvent.type(screen.getByLabelText('Email'), 'emp@x.com')
    await userEvent.type(screen.getByLabelText('Password'), 'pw')
    await userEvent.click(screen.getByRole('button', { name: 'Sign in' }))
    expect(await screen.findByRole('alert')).toHaveTextContent('Unable to sign in. Please try again.')
    expect(screen.getByLabelText('Email')).toHaveValue('emp@x.com')
  })
})

describe('reload persistence', () => {
  const session = { accessToken: 'a', refreshToken: 'r', tenantId: 't1' }
  const employeeUser = { id: 'u1', email: 'emp@x.com', firstName: 'Jane', lastName: 'Doe', roles: ['employee'] }

  it('renders the employee shell from a persisted session and user', async () => {
    setSession(session)
    localStorage.setItem(
      'employee360.employee.auth',
      JSON.stringify({ state: { user: employeeUser }, version: 0 }),
    )
    await useAuthStore.persist.rehydrate()
    renderAt('/')
    expect(screen.getByRole('button', { name: 'Sign out' })).toBeInTheDocument()
    expect(screen.getByText(/welcome, Jane Doe/i)).toBeInTheDocument()
  })

  it('persists a real UI login across a simulated reload', async () => {
    mock.onPost('/api/v1/auth/login').reply(200, loginOk('employee'))
    const first = renderAt('/login')
    await userEvent.type(screen.getByLabelText('Email'), 'emp@x.com')
    await userEvent.type(screen.getByLabelText('Password'), 'secret123')
    await userEvent.click(screen.getByRole('button', { name: 'Sign in' }))
    expect(await screen.findByRole('button', { name: 'Sign out' })).toBeInTheDocument()
    first.unmount()

    // Simulate reload: wipe in-memory store state, keep localStorage, rehydrate.
    // (setState re-persists, so snapshot the stored value and restore it before rehydrating)
    const persisted = localStorage.getItem('employee360.employee.auth')
    expect(persisted).not.toBeNull()
    useAuthStore.setState({ user: null, accessToken: null })
    localStorage.setItem('employee360.employee.auth', persisted as string)
    await useAuthStore.persist.rehydrate()
    useAuthStore.setState({ accessToken: getSession()?.accessToken ?? null })

    renderAt('/')
    expect(screen.getByRole('button', { name: 'Sign out' })).toBeInTheDocument()
    expect(screen.getByText(/welcome, Jane Doe/i)).toBeInTheDocument()
  })

  it('redirects to login with a token but no user', () => {
    setSession(session)
    renderAt('/')
    expect(screen.getByRole('heading', { name: /sign in to employee portal/i })).toBeInTheDocument()
  })

  it('redirects to login with a user but no token', () => {
    useAuthStore.setState({ user: employeeUser })
    renderAt('/')
    expect(screen.getByRole('heading', { name: /sign in to employee portal/i })).toBeInTheDocument()
  })

  it('redirects to login when persisted user has no valid role', () => {
    setSession(session)
    useAuthStore.setState({ user: { ...employeeUser, roles: ['unknown_role'] } })
    renderAt('/')
    expect(screen.getByRole('heading', { name: /sign in to employee portal/i })).toBeInTheDocument()
  })
})

describe('sign out', () => {
  it('signs out and navigates back to login', async () => {
    mock.onPost('/api/v1/auth/logout').reply(200, {})
    setSession({ accessToken: 'a', refreshToken: 'r', tenantId: 't1' })
    useAuthStore.setState({ user: { id: 'u1', email: 'emp@x.com', roles: ['employee'] } })

    renderAt('/')
    const signOutBtn = screen.getByRole('button', { name: 'Sign out' })
    await userEvent.click(signOutBtn)

    expect(await screen.findByRole('heading', { name: /sign in to employee portal/i })).toBeInTheDocument()
    expect(getSession()).toBeNull()
    expect(useAuthStore.getState().user).toBeNull()
  })
})

describe('forgot password', () => {
  it.each([200, 404, 500])('shows generic success message for API status %i', async (status) => {
    mock.onPost('/api/v1/auth/forgot-password').reply(status, {})
    renderAt('/forgot-password')
    await userEvent.type(screen.getByLabelText('Email'), 'emp@x.com')
    await userEvent.click(screen.getByRole('button', { name: 'Send reset link' }))
    expect(await screen.findByRole('status')).toHaveTextContent(FORGOT_SUCCESS_MESSAGE)
  })

  it('shows the same message on network failure', async () => {
    mock.onPost('/api/v1/auth/forgot-password').networkError()
    renderAt('/forgot-password')
    await userEvent.type(screen.getByLabelText('Email'), 'emp@x.com')
    await userEvent.click(screen.getByRole('button', { name: 'Send reset link' }))
    expect(await screen.findByRole('status')).toHaveTextContent(FORGOT_SUCCESS_MESSAGE)
  })
})

describe('reset password', () => {
  it('validates match and posts token + new password', async () => {
    mock.onPost('/api/v1/auth/reset-password').reply(200, {})
    renderAt('/reset-password?token=tok')
    await userEvent.type(screen.getByLabelText('New password'), 'password123')
    await userEvent.type(screen.getByLabelText('Confirm new password'), 'different1')
    await userEvent.click(screen.getByRole('button', { name: 'Reset password' }))
    expect(await screen.findByText('Passwords do not match')).toBeInTheDocument()

    await userEvent.clear(screen.getByLabelText('Confirm new password'))
    await userEvent.type(screen.getByLabelText('Confirm new password'), 'password123')
    await userEvent.click(screen.getByRole('button', { name: 'Reset password' }))
    expect(await screen.findByText(/has been reset/i)).toBeInTheDocument()
    expect(screen.getByRole('heading', { name: /sign in to employee portal/i })).toBeInTheDocument()
    expect(JSON.parse(mock.history.post[0].data)).toEqual({ token: 'tok', new_password: 'password123' })
  })

  it('shows invalid-link state without a token', () => {
    renderAt('/reset-password')
    expect(screen.getByRole('alert')).toHaveTextContent(/invalid/i)
  })

  it('clears any active session on success', async () => {
    setSession({ accessToken: 'a', refreshToken: 'r', tenantId: 't1' })
    useAuthStore.setState({ user: { id: 'u1', email: 'emp@x.com', roles: ['employee'] } })
    mock.onPost('/api/v1/auth/reset-password').reply(200, {})
    renderAt('/reset-password?token=tok')
    await userEvent.type(screen.getByLabelText('New password'), 'password123')
    await userEvent.type(screen.getByLabelText('Confirm new password'), 'password123')
    await userEvent.click(screen.getByRole('button', { name: 'Reset password' }))
    expect(await screen.findByText(/has been reset/i)).toBeInTheDocument()
    expect(getSession()).toBeNull()
    expect(useAuthStore.getState().user).toBeNull()
  })
})

describe('accept invitation', () => {
  it('shows invalid-link state without a token', () => {
    renderAt('/accept-invitation')
    expect(screen.getByRole('alert')).toHaveTextContent(/invalid or missing a token/i)
    expect(screen.getByRole('link', { name: /return to sign in/i })).toBeInTheDocument()
  })

  it('validates token on load before showing password fields and completes acceptance', async () => {
    mock.onGet('/api/v1/invitations/validate', { params: { token: 'inv-emp-token-123' } }).reply(200, {
      success: true,
      data: { valid: true, email: 'emp@acme.com' },
    })
    mock.onPost('/api/v1/invitations/accept').reply(200, { success: true, data: { message: 'invitation accepted' } })
    renderAt('/accept-invitation?token=inv-emp-token-123')

    // Password fields appear once token is validated
    expect(await screen.findByRole('heading', { name: /set your password/i })).toBeInTheDocument()

    // Password length validation
    await userEvent.type(screen.getByLabelText(/^password$/i), 'short')
    await userEvent.type(screen.getByLabelText(/^confirm password$/i), 'short')
    await userEvent.click(screen.getByRole('button', { name: /set password & accept/i }))
    expect(await screen.findByText('Password must be at least 8 characters')).toBeInTheDocument()

    // Password mismatch validation
    await userEvent.clear(screen.getByLabelText(/^password$/i))
    await userEvent.type(screen.getByLabelText(/^password$/i), 'employeeSecret123')
    await userEvent.clear(screen.getByLabelText(/^confirm password$/i))
    await userEvent.type(screen.getByLabelText(/^confirm password$/i), 'differentPassword')
    await userEvent.click(screen.getByRole('button', { name: /set password & accept/i }))
    expect(await screen.findByText('Passwords do not match')).toBeInTheDocument()

    // Valid submission
    await userEvent.clear(screen.getByLabelText(/^confirm password$/i))
    await userEvent.type(screen.getByLabelText(/^confirm password$/i), 'employeeSecret123')
    await userEvent.click(screen.getByRole('button', { name: /set password & accept/i }))

    expect(await screen.findByRole('heading', { name: /invitation accepted/i })).toBeInTheDocument()
    expect(screen.getByText(/your invitation has been accepted/i)).toBeInTheDocument()
    expect(JSON.parse(mock.history.post[0].data)).toEqual({ token: 'inv-emp-token-123', password: 'employeeSecret123' })

    // Clicking sign in lands on login page with acceptance alert
    await userEvent.click(screen.getByRole('link', { name: /sign in to employee portal/i }))
    expect(screen.getByRole('heading', { name: /sign in to employee portal/i })).toBeInTheDocument()
    expect(screen.getByText(/your invitation was accepted\. you can now sign in\./i)).toBeInTheDocument()
  })

  it('shows distinct actionable error immediately on load if token is expired or invalid without showing password fields', async () => {
    mock.onGet('/api/v1/invitations/validate', { params: { token: 'expired-token' } }).reply(400, {
      error: { code: 'BAD_REQUEST', message: 'invalid or expired invitation token' },
    })
    renderAt('/accept-invitation?token=expired-token')

    const alert = await screen.findByRole('alert')
    expect(alert).toHaveTextContent(/this invitation link is invalid, has expired, or has already been accepted/i)
    expect(screen.queryByLabelText(/^password$/i)).toBeNull()
    expect(screen.queryByRole('button', { name: /set password & accept/i })).toBeNull()
    expect(screen.getByRole('link', { name: /return to sign in/i })).toBeInTheDocument()
  })

  it('shows a distinct actionable error if token is revoked / not pending on load', async () => {
    mock.onGet('/api/v1/invitations/validate', { params: { token: 'revoked-token' } }).reply(409, {
      error: { code: 'CONFLICT', message: 'invitation is no longer pending' },
    })
    renderAt('/accept-invitation?token=revoked-token')

    const alert = await screen.findByRole('alert')
    expect(alert).toHaveTextContent(/this invitation link is invalid, has expired, or has already been accepted/i)
    expect(screen.queryByLabelText(/^password$/i)).toBeNull()
  })

  it('clears active session state on successful acceptance', async () => {
    mock.onGet('/api/v1/invitations/validate', { params: { token: 'valid-tok' } }).reply(200, {
      success: true,
      data: { valid: true, email: 'emp@acme.com' },
    })
    mock.onPost('/api/v1/invitations/accept').reply(200, { success: true, data: { message: 'invitation accepted' } })
    setSession({ accessToken: 'old-access', refreshToken: 'old-refresh', tenantId: 'old-tenant' })
    useAuthStore.setState({ user: { id: 'old-u', email: 'old@x.com', roles: ['employee'] } })

    renderAt('/accept-invitation?token=valid-tok')
    expect(await screen.findByRole('heading', { name: /set your password/i })).toBeInTheDocument()

    await userEvent.type(screen.getByLabelText(/^password$/i), 'employeeSecret123')
    await userEvent.type(screen.getByLabelText(/^confirm password$/i), 'employeeSecret123')
    await userEvent.click(screen.getByRole('button', { name: /set password & accept/i }))

    expect(await screen.findByRole('heading', { name: /invitation accepted/i })).toBeInTheDocument()
    expect(getSession()).toBeNull()
    expect(useAuthStore.getState().user).toBeNull()
  })
})


