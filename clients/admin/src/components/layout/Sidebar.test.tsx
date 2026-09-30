import { beforeEach, describe, expect, it } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { AdminShell } from './AdminShell'
import { useUiStore } from '../../stores/uiStore'
import { useAuthStore } from '../../stores/authStore'

function renderShell(path = '/invitations') {
  return render(
    <QueryClientProvider client={new QueryClient()}>
      <MemoryRouter initialEntries={[path]}>
        <Routes>
          <Route element={<AdminShell />}>
            <Route path="/invitations" element={<p>invitations page</p>} />
            <Route path="/" element={<p>home</p>} />
          </Route>
        </Routes>
      </MemoryRouter>
    </QueryClientProvider>,
  )
}

beforeEach(() => {
  useUiStore.setState({ sidebarCollapsed: false, sidebarMobileOpen: false })
  useAuthStore.setState({ user: { id: 'u1', email: 'admin@x.com', roles: ['admin'] } })
})

describe('Sidebar', () => {
  it('renders nav links, highlights the active one, and shows the user', () => {
    renderShell()
    const link = screen.getByRole('link', { name: 'Invitations' })
    expect(link).toHaveAttribute('href', '/invitations')
    expect(link).toHaveAttribute('aria-current', 'page')
    expect(screen.getByText('admin@x.com')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Sign out' })).toBeInTheDocument()
  })

  it('does not mark the link active on other routes', () => {
    renderShell('/')
    expect(screen.getByRole('link', { name: 'Invitations' })).not.toHaveAttribute('aria-current')
  })

  it('toggles collapsed state in the ui store', async () => {
    renderShell()
    await userEvent.click(screen.getByRole('button', { name: 'Collapse sidebar' }))
    expect(useUiStore.getState().sidebarCollapsed).toBe(true)
    expect(screen.getByRole('complementary', { name: 'Sidebar' })).toHaveAttribute('data-collapsed', 'true')
    await userEvent.click(screen.getByRole('button', { name: 'Expand sidebar' }))
    expect(useUiStore.getState().sidebarCollapsed).toBe(false)
  })

  it('opens and closes the mobile drawer', async () => {
    renderShell()
    await userEvent.click(screen.getByRole('button', { name: 'Open menu' }))
    expect(useUiStore.getState().sidebarMobileOpen).toBe(true)
    await userEvent.click(screen.getByTestId('sidebar-backdrop'))
    expect(useUiStore.getState().sidebarMobileOpen).toBe(false)
    await userEvent.click(screen.getByRole('button', { name: 'Open menu' }))
    await userEvent.click(screen.getByRole('link', { name: 'Invitations' }))
    expect(useUiStore.getState().sidebarMobileOpen).toBe(false)
  })
})
