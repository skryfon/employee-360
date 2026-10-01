import { describe, expect, it } from 'vitest'
import { render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { AcceptInvitationPage } from './AcceptInvitationPage'

describe('AcceptInvitationPage (shared)', () => {
  it('shows the invalid-link state when no token is present', () => {
    render(
      <QueryClientProvider client={new QueryClient()}>
        <MemoryRouter initialEntries={['/accept-invitation']}>
          <AcceptInvitationPage
            currentApp="admin"
            signInLabel="Sign in to Admin"
            otherPortalName="Employee Portal"
            otherPortalNameInAcceptedHint="Employee Portal"
          />
        </MemoryRouter>
      </QueryClientProvider>,
    )
    expect(screen.getByText('Invalid invitation link')).toBeInTheDocument()
    expect(screen.getByRole('link', { name: 'Return to sign in' })).toBeInTheDocument()
  })
})
