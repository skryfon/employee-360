import { afterEach, beforeEach, describe, expect, it } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import MockAdapter from 'axios-mock-adapter'
import { apiClient, clearSession, getSession } from '@employee360/api-client'
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

const loginOk = (role: string) => ({
  success: true,
  data: {
    access_token: 'a',
    refresh_token: 'r',
    user: { id: 'u1', email: 'a@x.com', tenant_id: 't1', roles: [{ name: role }] },
  },
})

beforeEach(() => {
  mock = new MockAdapter(apiClient)
  mock.onGet('/api/v1/health').reply(200, { success: true, data: { status: 'ok', app: 'e', database: 'ok' } })
  clearSession()
  useAuthStore.getState().clear()
})
afterEach(() => mock.restore())

describe('login', () => {
  it('redirects unauthenticated users to login', () => {
    renderAt('/')
    expect(screen.getByRole('heading', { name: /sign in/i })).toBeInTheDocument()
  })

  it('signs in and lands in the admin shell', async () => {
    mock.onPost('/api/v1/auth/login').reply(200, loginOk('admin'))
    renderAt('/login')
    await userEvent.type(screen.getByLabelText('Email'), 'a@x.com')
    await userEvent.type(screen.getByLabelText('Password'), 'pw')
    await userEvent.click(screen.getByRole('button', { name: 'Sign in' }))
    expect(await screen.findByRole('button', { name: 'Sign out' })).toBeInTheDocument()
    expect(getSession()?.accessToken).toBe('a')
    expect(useAuthStore.getState().user?.roles).toEqual(['admin'])
  })

  it('shows one generic error for bad credentials', async () => {
    mock.onPost('/api/v1/auth/login').reply(401, { error: { message: 'user not found' } })
    renderAt('/login')
    await userEvent.type(screen.getByLabelText('Email'), 'a@x.com')
    await userEvent.type(screen.getByLabelText('Password'), 'pw')
    await userEvent.click(screen.getByRole('button', { name: 'Sign in' }))
    expect(await screen.findByRole('alert')).toHaveTextContent('Invalid email or password.')
    expect(screen.queryByText(/not found/i)).toBeNull()
  })

  it('rejects non-admin accounts with the same generic error', async () => {
    mock.onPost('/api/v1/auth/login').reply(200, loginOk('employee'))
    mock.onPost('/api/v1/auth/logout').reply(200, {})
    renderAt('/login')
    await userEvent.type(screen.getByLabelText('Email'), 'a@x.com')
    await userEvent.type(screen.getByLabelText('Password'), 'pw')
    await userEvent.click(screen.getByRole('button', { name: 'Sign in' }))
    expect(await screen.findByRole('alert')).toHaveTextContent('Invalid email or password.')
    expect(useAuthStore.getState().user).toBeNull()
  })
})

describe('forgot password', () => {
  it.each([200, 404, 500])('shows the same generic message for API status %i', async (status) => {
    mock.onPost('/api/v1/auth/forgot-password').reply(status, {})
    renderAt('/forgot-password')
    await userEvent.type(screen.getByLabelText('Email'), 'a@x.com')
    await userEvent.click(screen.getByRole('button', { name: 'Send reset link' }))
    expect(await screen.findByRole('status')).toHaveTextContent(FORGOT_SUCCESS_MESSAGE)
  })
  it('shows the same message on network failure', async () => {
    mock.onPost('/api/v1/auth/forgot-password').networkError()
    renderAt('/forgot-password')
    await userEvent.type(screen.getByLabelText('Email'), 'a@x.com')
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
    await waitFor(() => expect(screen.getByText(/has been reset/i)).toBeInTheDocument())
    expect(JSON.parse(mock.history.post[0].data)).toEqual({ token: 'tok', new_password: 'password123' })
  })
  it('shows invalid-link state without a token', () => {
    renderAt('/reset-password')
    expect(screen.getByRole('alert')).toHaveTextContent(/invalid/i)
  })
})
