import axios from 'axios';

/**
 * Extract a user-facing message from an error: the backend's
 * `{ error: { message } }` body for an AxiosError, else `err.message`, else
 * `fallback`.
 */
export function getErrorMessage(err: unknown, fallback = 'Something went wrong'): string {
  if (axios.isAxiosError<{ error?: { message?: string } }>(err)) {
    return err.response?.data?.error?.message || err.message || fallback;
  }
  if (err instanceof Error && err.message) return err.message;
  return fallback;
}
