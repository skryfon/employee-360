import { afterEach, beforeEach, describe, expect, it } from 'vitest';
import MockAdapter from 'axios-mock-adapter';
import { apiClient, configureApiClient } from './client.ts';

/**
 * Minimal shape of axios's internal (untyped-in-the-public-API) interceptor
 * manager, used only to assert that handlers were actually registered.
 * Avoids `any` per project coding rules while still reaching into the one
 * axios internal this suite needs to introspect.
 */
interface InterceptorManager {
  handlers: Array<unknown | null>;
}

describe('configureApiClient (AC-2)', () => {
  const originalBaseURL = apiClient.defaults.baseURL;

  afterEach(() => {
    apiClient.defaults.baseURL = originalBaseURL;
  });

  it('updates apiClient.defaults.baseURL to the configured value', () => {
    configureApiClient({ baseURL: 'http://example.test' });

    expect(apiClient.defaults.baseURL).toBe('http://example.test');
  });

  it('leaves the existing baseURL untouched when given an empty string', () => {
    configureApiClient({ baseURL: 'http://example.test' });
    configureApiClient({ baseURL: '' });

    expect(apiClient.defaults.baseURL).toBe('http://example.test');
  });
});

describe('request interceptor stubs (AC-3, AC-4)', () => {
  let mock: MockAdapter;

  beforeEach(() => {
    mock = new MockAdapter(apiClient);
  });

  afterEach(() => {
    mock.restore();
  });

  it('registers at least two request interceptor handlers (JWT stub + X-Tenant-ID stub)', () => {
    const manager = apiClient.interceptors.request as unknown as InterceptorManager;
    const activeHandlers = manager.handlers.filter((handler) => handler !== null);

    expect(activeHandlers.length).toBeGreaterThanOrEqual(2);
  });

  it('does not attach an Authorization header today (JWT interceptor is a no-op stub)', async () => {
    mock.onGet('/probe').reply(200, {});

    await apiClient.get('/probe');

    const [request] = mock.history.get;
    expect(request.headers?.Authorization).toBeUndefined();
  });

  it('does not attach an X-Tenant-ID header today (tenant interceptor is a no-op stub)', async () => {
    mock.onGet('/probe').reply(200, {});

    await apiClient.get('/probe');

    const [request] = mock.history.get;
    expect(request.headers?.['X-Tenant-ID']).toBeUndefined();
  });

  it('passes the rest of the outgoing request config through unchanged', async () => {
    mock.onGet('/probe').reply(200, {});

    await apiClient.get('/probe', { params: { a: 1 } });

    const [request] = mock.history.get;
    expect(request.url).toBe('/probe');
    expect(request.params).toEqual({ a: 1 });
  });
});
