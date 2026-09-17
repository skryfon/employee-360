import axios, { type AxiosRequestConfig, type AxiosResponse } from 'axios';

/**
 * Configuration accepted by {@link configureApiClient}.
 *
 * This package is framework-independent: it never reaches into
 * `import.meta.env` or `process.env` itself. Each consuming Vite app is
 * responsible for reading its own environment (e.g.
 * `import.meta.env.VITE_API_BASE_URL`) and passing it in explicitly.
 */
export interface ApiClientConfig {
  baseURL: string;
}

/**
 * Sensible fallback so the client is usable (e.g. in tests) even if
 * `configureApiClient` is never called.
 */
const DEFAULT_BASE_URL = 'http://localhost:8080';

/**
 * The shared Axios instance used by every generated hook and hand-written
 * wrapper in this package. Feature code should never construct its own
 * Axios instance — import this one (or go through the generated hooks,
 * which are wired to it via the mutator below).
 */
export const apiClient = axios.create({
  baseURL: DEFAULT_BASE_URL,
});

/**
 * Configure the shared Axios instance. Call this once from each consuming
 * app's entrypoint (e.g. `main.tsx`), before rendering, passing that app's
 * own environment configuration:
 *
 *   configureApiClient({ baseURL: import.meta.env.VITE_API_BASE_URL })
 */
export function configureApiClient(config: ApiClientConfig): void {
  if (config.baseURL) {
    apiClient.defaults.baseURL = config.baseURL;
  }
}

// --- JWT auth interceptor (stub) ---
// Follow-up: once a Zustand auth store exposes the in-memory access token
// (per the multi-tenant invariant, tokens are never persisted/re-derived
// mid-session), attaching it here is a one-line change:
//
//   const token = getAccessToken();
//   if (token) {
//     config.headers.set('Authorization', `Bearer ${token}`);
//   }
apiClient.interceptors.request.use((config) => {
  // Intentionally a no-op for now — see comment above.
  return config;
});

// --- X-Tenant-ID interceptor (stub) ---
// Follow-up: tenant id is resolved once at login (email domain or JWT
// claims) and never re-derived or switched mid-session. Once the auth
// store exposes it, attaching it here is a one-line change:
//
//   const tenantId = getTenantId();
//   if (tenantId) {
//     config.headers.set('X-Tenant-ID', tenantId);
//   }
apiClient.interceptors.request.use((config) => {
  // Intentionally a no-op for now — see comment above.
  return config;
});

/**
 * The Orval custom mutator (see `orval.config.ts` -> `override.mutator`).
 * Every Orval-generated hook/function calls this instead of raw axios, so
 * generated code always goes through the shared instance above (and its
 * interceptors) rather than creating its own.
 */
export const apiRequest = async <T>(config: AxiosRequestConfig): Promise<T> => {
  const response: AxiosResponse<T> = await apiClient.request<T>(config);
  return response.data;
};

export default apiClient;
