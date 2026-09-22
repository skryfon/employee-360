import { afterEach, beforeEach, describe, expect, it } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import MockAdapter from 'axios-mock-adapter';
import { apiClient } from '@employee360/api-client';
import App from './App.tsx';

// Mocks at the HTTP layer (the shared apiClient instance), never the
// useHealth hook or the @employee360/api-client module itself, so these
// tests prove the real wiring: App -> useHealth() -> apiClient -> a real
// HTTP request to /api/v1/health (AC-6, AC-7, AC-8).
const HEALTH_URL = '/api/v1/health';

function renderApp() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return render(
    <QueryClientProvider client={queryClient}>
      <App />
    </QueryClientProvider>,
  );
}

describe('App', () => {
  let mock: MockAdapter;

  beforeEach(() => {
    mock = new MockAdapter(apiClient);
  });

  afterEach(() => {
    mock.restore();
  });

  it('shows loading copy while the health check is pending', () => {
    mock.onGet(HEALTH_URL).reply(200, {
      success: true,
      data: { status: 'ok', app: 'employee360', database: 'ok' },
    });

    renderApp();

    expect(screen.getByText('Checking API health…')).toBeInTheDocument();
  });

  it('shows error copy when the health check fails', async () => {
    mock.onGet(HEALTH_URL).reply(500);

    renderApp();

    await waitFor(() =>
      expect(screen.getByText(/API health check failed/i)).toBeInTheDocument(),
    );
  });

  it('renders the health data and actually calls GET /api/v1/health through apiClient', async () => {
    mock.onGet(HEALTH_URL).reply(200, {
      success: true,
      data: { status: 'ok', app: 'employee360', database: 'ok' },
    });

    renderApp();

    await waitFor(() => expect(screen.getAllByText('ok')).toHaveLength(2));
    expect(screen.getByText('employee360')).toBeInTheDocument();

    // Proves the app really called through packages/api-client, not just
    // rendered static markup.
    expect(mock.history.get).toHaveLength(1);
    expect(mock.history.get[0].url).toBe(HEALTH_URL);
    expect(mock.history.get[0].method).toBe('get');
  });
});
