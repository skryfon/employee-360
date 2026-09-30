export { apiClient, apiRequest, configureApiClient, configureAuth } from './client.ts';
export type { ApiClientConfig, AuthConfig } from './client.ts';

export {
  getSession,
  setSession,
  clearSession,
  subscribeSession,
  setSessionStorage,
} from './session.ts';
export type { Session, SessionStorageAdapter } from './session.ts';
export {
  createLocalStorageSessionAdapter,
  DEFAULT_SESSION_STORAGE_KEY,
} from './sessionStorage.ts';

export { unwrapSingleEntity, unwrapListResponse } from './unwrap.ts';
export type { ApiEnvelope } from './unwrap.ts';
export { getErrorMessage } from './errors.ts';

export {
  login,
  refresh,
  logout,
  forgotPassword,
  resetPassword,
  inviteUser,
  listInvitations,
  listRoles,
  resendInvitation,
  revokeInvitation,
  acceptInvitation,
} from './auth.ts';
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
  RoleOption,
} from './auth.ts';

export { fetchHealth, useHealth } from './health.ts';
export type { HealthResponse } from './health.ts';

export * from './generated/models/index.ts';
export * from './generated/hooks/index.ts';
