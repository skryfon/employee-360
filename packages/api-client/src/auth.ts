import { isAxiosError } from 'axios';
import { refreshSession } from './client.ts';
import { unwrapListResponse, unwrapSingleEntity } from './unwrap.ts';
import { clearSession, getSession, setSession } from './session.ts';
import {
  postApiV1AuthForgotPassword,
  postApiV1AuthLogin,
  postApiV1AuthLogout,
  postApiV1AuthResetPassword,
} from './generated/hooks/auth/auth.ts';
import {
  deleteApiV1UsersInvitationsId,
  getApiV1UsersInvitations,
  postApiV1InvitationsAccept,
  postApiV1UsersInvitations,
  postApiV1UsersInvitationsIdResend,
} from './generated/hooks/invitations/invitations.ts';
import { getApiV1Roles } from './generated/hooks/roles/roles.ts';
import type { GithubComSkryfonEmployee360BackendInternalTypesAuthLoginRequest as LoginRequest } from './generated/models/githubComSkryfonEmployee360BackendInternalTypesAuthLoginRequest.ts';
import type { GithubComSkryfonEmployee360BackendInternalTypesAuthLoginResponse as LoginResponse } from './generated/models/githubComSkryfonEmployee360BackendInternalTypesAuthLoginResponse.ts';
import type { GithubComSkryfonEmployee360BackendInternalTypesAuthForgotPasswordRequest as ForgotPasswordRequest } from './generated/models/githubComSkryfonEmployee360BackendInternalTypesAuthForgotPasswordRequest.ts';
import type { GithubComSkryfonEmployee360BackendInternalTypesAuthResetPasswordRequest as ResetPasswordRequest } from './generated/models/githubComSkryfonEmployee360BackendInternalTypesAuthResetPasswordRequest.ts';
import type { GithubComSkryfonEmployee360BackendInternalTypesInvitationInviteUserRequest as InviteUserRequest } from './generated/models/githubComSkryfonEmployee360BackendInternalTypesInvitationInviteUserRequest.ts';
import type { GithubComSkryfonEmployee360BackendInternalTypesInvitationAcceptInvitationRequest as AcceptInvitationRequest } from './generated/models/githubComSkryfonEmployee360BackendInternalTypesInvitationAcceptInvitationRequest.ts';
import type { GithubComSkryfonEmployee360BackendInternalTypesInvitationInvitationResponse as Invitation } from './generated/models/githubComSkryfonEmployee360BackendInternalTypesInvitationInvitationResponse.ts';
import type { GithubComSkryfonEmployee360BackendInternalDeliveryHttpResponseMeta as PageMeta } from './generated/models/githubComSkryfonEmployee360BackendInternalDeliveryHttpResponseMeta.ts';
import type { GetApiV1UsersInvitationsParams as ListInvitationsParams } from './generated/models/getApiV1UsersInvitationsParams.ts';

/** Role option for the invite form (`GET /api/v1/roles`; `id` is the invite `role_id`). */
export interface RoleOption {
  id: string;
  name: string;
}

export type {
  LoginRequest,
  LoginResponse,
  ForgotPasswordRequest,
  ResetPasswordRequest,
  InviteUserRequest,
  AcceptInvitationRequest,
  Invitation,
  PageMeta,
  ListInvitationsParams,
};

/**
 * `POST /api/v1/auth/login` — serves every role. Stores the token pair and
 * tenant (from the returned user) in the shared session so the interceptors
 * pick them up.
 */
export async function login(credentials: LoginRequest): Promise<LoginResponse> {
  const data = unwrapSingleEntity(await postApiV1AuthLogin(credentials));
  if (!data.access_token || !data.refresh_token) {
    throw new Error('Login response missing tokens');
  }
  setSession({
    accessToken: data.access_token,
    refreshToken: data.refresh_token,
    tenantId: data.user?.tenant_id ?? null,
    expiresAt: data.expires_at,
  });
  return data;
}

/**
 * `POST /api/v1/auth/refresh` — rotates the token pair and updates the
 * session. Shares the single in-flight refresh used by the 401 interceptor.
 * Resolves with the new access token.
 */
export function refresh(): Promise<string> {
  return refreshSession();
}

/**
 * `POST /api/v1/auth/logout` — revokes the refresh token server-side. The
 * endpoint requires a valid access token, but `/auth/logout` is deliberately
 * excluded from the 401 interceptor's silent refresh (the body carries the
 * refresh token, which a replay would have rotated). So on a 401 we do one
 * best-effort refresh here and retry with the rotated token, so the server-side
 * revoke is not skipped for an expired access token. The local session is
 * always cleared, even if the request fails.
 */
export async function logout(): Promise<void> {
  try {
    const refreshToken = getSession()?.refreshToken;
    if (!refreshToken) return;
    try {
      await postApiV1AuthLogout({ refresh_token: refreshToken });
    } catch (error) {
      if (!isAxiosError(error) || error.response?.status !== 401) throw error;
      try {
        await refreshSession();
      } catch {
        return; // refresh token already invalid/unreachable: nothing more to revoke
      }
      const rotated = getSession()?.refreshToken;
      if (rotated) await postApiV1AuthLogout({ refresh_token: rotated });
    }
  } finally {
    clearSession();
  }
}

export async function forgotPassword(body: ForgotPasswordRequest): Promise<void> {
  await postApiV1AuthForgotPassword(body);
}

export async function resetPassword(body: ResetPasswordRequest): Promise<void> {
  await postApiV1AuthResetPassword(body);
}

export async function inviteUser(body: InviteUserRequest): Promise<Invitation> {
  return unwrapSingleEntity(await postApiV1UsersInvitations(body));
}

export async function listInvitations(
  params?: ListInvitationsParams,
  signal?: AbortSignal,
): Promise<{ data: Invitation[]; meta?: PageMeta }> {
  return unwrapListResponse(await getApiV1UsersInvitations(params, signal));
}

export async function resendInvitation(id: string): Promise<Invitation> {
  return unwrapSingleEntity(await postApiV1UsersInvitationsIdResend(id));
}

export async function revokeInvitation(id: string): Promise<void> {
  await deleteApiV1UsersInvitationsId(id);
}

/** Unauthenticated; the invitee sets a password to activate their account. */
export async function acceptInvitation(body: AcceptInvitationRequest): Promise<void> {
  await postApiV1InvitationsAccept(body);
}

/**
 * `GET /api/v1/roles` — admin/super_admin only; super_admin is already excluded
 * server-side. Narrows the generated (all-optional) role shape to `RoleOption`.
 */
export async function listRoles(signal?: AbortSignal): Promise<RoleOption[]> {
  const { data } = unwrapListResponse(await getApiV1Roles(signal));
  return data.flatMap((r) => (r.id && r.name ? [{ id: r.id, name: r.name }] : []));
}
