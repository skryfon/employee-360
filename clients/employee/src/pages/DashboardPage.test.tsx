import { afterEach, beforeEach, describe, expect, it } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import MockAdapter from 'axios-mock-adapter'
import { apiClient } from '@employee360/api-client'
import DashboardPage from './DashboardPage'

const HEALTH_URL = '/api/v1/health'

function renderPage() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return render(
    <QueryClientProvider client={queryClient}>
      <DashboardPage />
    </QueryClientProvider>,
  )
}

describe('DashboardPage', () => {
  let mock: MockAdapter

  beforeEach(() => {
    mock = new MockAdapter(apiClient)
  })

  afterEach(() => {
    mock.restore()
  })

  it('shows a skeleton while the health check is pending', () => {
    mock.onGet(HEALTH_URL).reply(200, {
      success: true,
      data: { status: 'ok', app: 'employee360', database: 'ok' },
    })

    renderPage()

    expect(screen.getByRole('status')).toHaveTextContent('Checking API health…')
    expect(screen.getByRole('status')).toHaveAttribute('aria-busy', 'true')
  })

  it('shows error copy when the health check fails', async () => {
    mock.onGet(HEALTH_URL).reply(500)

    renderPage()

    await waitFor(() =>
      expect(screen.getByText(/API health check failed/i)).toBeInTheDocument(),
    )
  })

  it('renders the health data and actually calls GET /api/v1/health through apiClient', async () => {
    mock.onGet(HEALTH_URL).reply(200, {
      success: true,
      data: { status: 'ok', app: 'employee360', database: 'ok' },
    })

    renderPage()

    await waitFor(() => expect(screen.getAllByText('ok')).toHaveLength(2))
    expect(screen.getByText('employee360')).toBeInTheDocument()

    expect(mock.history.get).toHaveLength(1)
    expect(mock.history.get[0].url).toBe(HEALTH_URL)
    expect(mock.history.get[0].method).toBe('get')
  })
})
