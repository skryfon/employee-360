import { afterEach, beforeEach, describe, expect, it } from 'vitest'
import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import MockAdapter from 'axios-mock-adapter'
import { apiClient, clearSession, setSession } from '@employee360/api-client'
import App from '../../../App'
import { useAuthStore } from '../../../stores/authStore'

let mock: MockAdapter
const ROLE_ID = '11111111-1111-4111-8111-111111111111'

const inv = (id: string, email: string, status: string) => ({ id, email, status, created_at: '2026-01-01T00:00:00Z' })
let rows = [inv('i1', 'p@x.com', 'pending'), inv('i2', 'a@x.com', 'accepted'), inv('i3', 'r@x.com', 'revoked')]

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

function signIn() {
  setSession({ accessToken: 'a', refreshToken: 'r', tenantId: 't1' })
  useAuthStore.setState({
    accessToken: 'a',
    user: { id: 'u1', email: 'admin@x.com', roles: ['admin'] },
  })
}

beforeEach(() => {
  rows = [inv('i1', 'p@x.com', 'pending'), inv('i2', 'a@x.com', 'accepted'), inv('i3', 'r@x.com', 'revoked')]
  mock = new MockAdapter(apiClient)
  mock.onGet('/api/v1/users/invitations').reply(() => [200, { success: true, data: rows, meta: { total_pages: 1 } }])
  mock.onGet('/api/v1/roles').reply(200, { success: true, data: [{ id: ROLE_ID, name: 'admin' }] })
  clearSession()
  useAuthStore.getState().clear()
  localStorage.clear()
})
afterEach(() => mock.restore())

describe('invitations', () => {
  it('redirects unauthenticated users to login', () => {
    renderAt('/invitations')
    expect(screen.getByRole('heading', { name: /sign in/i })).toBeInTheDocument()
  })

  it('shows resend/revoke only for pending invitations', async () => {
    signIn()
    renderAt('/invitations')
    expect(await screen.findByText('p@x.com')).toBeInTheDocument()
    expect(screen.getAllByRole('button', { name: /^Resend/ })).toHaveLength(1)
    expect(screen.getAllByRole('button', { name: /^Revoke/ })).toHaveLength(1)
    expect(screen.getByText('Accepted')).toBeInTheDocument()
  })

  it('revokes after confirmation and refreshes the list', async () => {
    signIn()
    mock.onDelete('/api/v1/users/invitations/i1').reply(() => {
      rows = rows.map((r) => (r.id === 'i1' ? { ...r, status: 'revoked' } : r))
      return [200, { success: true }]
    })
    renderAt('/invitations')
    await userEvent.click(await screen.findByRole('button', { name: /Revoke invitation for p@x.com/ }))
    const dialog = screen.getByRole('dialog')
    expect(mock.history.delete).toHaveLength(0)
    await userEvent.click(within(dialog).getByRole('button', { name: 'Revoke' }))
    await waitForRevokeDone()
    expect(mock.history.delete).toHaveLength(1)
  })

  it('resends and invalidates the list', async () => {
    signIn()
    mock.onPost('/api/v1/users/invitations/i1/resend').reply(200, { success: true, data: rows[0] })
    renderAt('/invitations')
    await userEvent.click(await screen.findByRole('button', { name: /Resend invitation to p@x.com/ }))
    await waitForCondition(() => mock.history.get.filter((r) => r.url === '/api/v1/users/invitations').length >= 2)
    expect(mock.history.post).toHaveLength(1)
  })

  it('list page has a Create invitation button that navigates and no inline form', async () => {
    signIn()
    renderAt('/invitations')
    expect(screen.queryByRole('form', { name: 'Invite user' })).toBeNull()
    await userEvent.click(await screen.findByRole('link', { name: 'Create invitation' }))
    expect(await screen.findByRole('heading', { name: 'Create invitation' })).toBeInTheDocument()
    expect(screen.getByRole('form', { name: 'Invite user' })).toBeInTheDocument()
    expect(screen.getByRole('link', { name: 'Invitations' })).toHaveAttribute('aria-current', 'page')
  })

  it('redirects unauthenticated users from /invitations/new to login', () => {
    renderAt('/invitations/new')
    expect(screen.getByRole('heading', { name: /sign in/i })).toBeInTheDocument()
  })

  it('cancel and back links return to the list', async () => {
    signIn()
    renderAt('/invitations/new')
    await userEvent.click(await screen.findByRole('button', { name: 'Cancel' }))
    expect(await screen.findByRole('link', { name: 'Create invitation' })).toBeInTheDocument()
    await userEvent.click(screen.getByRole('link', { name: 'Create invitation' }))
    await userEvent.click(await screen.findByRole('link', { name: 'Back to invitations' }))
    expect(await screen.findByRole('link', { name: 'Create invitation' })).toBeInTheDocument()
  })

  it('validates, submits the invite form, and redirects to the list', async () => {
    signIn()
    mock.onPost('/api/v1/users/invitations').reply(201, { success: true, data: rows[0] })
    renderAt('/invitations/new')
    await userEvent.click(await screen.findByRole('button', { name: 'Send invitation' }))
    expect(await screen.findByText('Email is required')).toBeInTheDocument()
    expect(screen.getAllByRole('alert').map((a) => a.textContent)).toContain('Select a role')
    await userEvent.type(screen.getByLabelText('Email'), 'new@x.com')
    await screen.findByRole('option', { name: 'admin' })
    await userEvent.selectOptions(screen.getByLabelText('Role'), ROLE_ID)
    await userEvent.click(screen.getByRole('button', { name: 'Send invitation' }))
    expect(await screen.findByRole('link', { name: 'Create invitation' })).toBeInTheDocument()
    expect(await screen.findByText('p@x.com')).toBeInTheDocument()
    expect(JSON.parse(mock.history.post[0].data)).toEqual({ email: 'new@x.com', role_id: ROLE_ID })
  })

  it('shows an error when roles fail to load', async () => {
    signIn()
    mock.onGet('/api/v1/roles').reply(500, { success: false, error: { message: 'boom' } })
    renderAt('/invitations/new')
    expect(await screen.findByText('Could not load roles.')).toBeInTheDocument()
  })

  it('shows an error when resend fails', async () => {
    signIn()
    mock.onPost('/api/v1/users/invitations/i1/resend').reply(500, { success: false, error: { message: 'Resend exploded' } })
    renderAt('/invitations')
    await userEvent.click(await screen.findByRole('button', { name: /Resend invitation to p@x.com/ }))
    expect(await screen.findByText('Resend exploded')).toBeInTheDocument()
  })

  it('shows revoke failure in the dialog and keeps it open', async () => {
    signIn()
    mock.onDelete('/api/v1/users/invitations/i1').reply(500, { success: false, error: { message: 'Revoke exploded' } })
    renderAt('/invitations')
    await userEvent.click(await screen.findByRole('button', { name: /Revoke invitation for p@x.com/ }))
    const dialog = screen.getByRole('dialog')
    await userEvent.click(within(dialog).getByRole('button', { name: 'Revoke' }))
    expect(await within(dialog).findByText('Revoke exploded')).toBeInTheDocument()
    expect(screen.getByRole('dialog')).toBeInTheDocument()
  })

  it('modal moves focus in, closes on Escape, and restores focus to the trigger', async () => {
    signIn()
    renderAt('/invitations')
    const trigger = await screen.findByRole('button', { name: /Revoke invitation for p@x.com/ })
    await userEvent.click(trigger)
    const dialog = screen.getByRole('dialog')
    expect(within(dialog).getByRole('button', { name: 'Cancel' })).toHaveFocus()
    await userEvent.tab()
    await userEvent.tab()
    expect(dialog.contains(document.activeElement)).toBe(true)
    await userEvent.keyboard('{Escape}')
    expect(screen.queryByRole('dialog')).toBeNull()
    expect(trigger).toHaveFocus()
  })
})

import { waitFor } from '@testing-library/react'
async function waitForCondition(cond: () => boolean) {
  await waitFor(() => expect(cond()).toBe(true))
}
async function waitForRevokeDone() {
  await waitFor(() => expect(screen.queryByRole('dialog')).toBeNull())
  await waitFor(() => expect(screen.queryAllByRole('button', { name: /^Revoke invitation/ })).toHaveLength(0))
}
