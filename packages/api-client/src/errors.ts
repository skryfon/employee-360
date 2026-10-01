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

/**
 * Stable error codes returned by the invitation validate/accept endpoints.
 * Features must branch on these, never on message text.
 */
export const INVITATION_ERROR_CODES = {
  INVALID_TOKEN: 'INVALID_TOKEN',
  EXPIRED: 'INVITATION_EXPIRED',
  REVOKED: 'INVITATION_REVOKED',
  ACCEPTED: 'INVITATION_ACCEPTED',
} as const;

/** Read the backend's `{ error: { code } }` from an AxiosError, else `undefined`. */
export function getErrorCode(err: unknown): string | undefined {
  if (axios.isAxiosError<{ error?: { code?: string } }>(err)) {
    return err.response?.data?.error?.code || undefined;
  }
  return undefined;
}
