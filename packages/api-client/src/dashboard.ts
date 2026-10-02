import { unwrapSingleEntity } from './unwrap.ts';
import {
  getApiV1DashboardAdmin,
  getApiV1DashboardSuperAdmin,
} from './generated/hooks/dashboard/dashboard.ts';
import type { GithubComSkryfonEmployee360BackendInternalTypesDashboardAdminDashboardResponse as AdminDashboard } from './generated/models/githubComSkryfonEmployee360BackendInternalTypesDashboardAdminDashboardResponse.ts';
import type { GithubComSkryfonEmployee360BackendInternalTypesDashboardSuperAdminDashboardResponse as SuperAdminDashboard } from './generated/models/githubComSkryfonEmployee360BackendInternalTypesDashboardSuperAdminDashboardResponse.ts';
import type { GithubComSkryfonEmployee360BackendInternalTypesDashboardRoleUserCountResponse as RoleUserCount } from './generated/models/githubComSkryfonEmployee360BackendInternalTypesDashboardRoleUserCountResponse.ts';
import type { GithubComSkryfonEmployee360BackendInternalTypesDashboardTenantInfoResponse as TenantInfo } from './generated/models/githubComSkryfonEmployee360BackendInternalTypesDashboardTenantInfoResponse.ts';

export type { AdminDashboard, SuperAdminDashboard, RoleUserCount, TenantInfo };

/** `GET /api/v1/dashboard/admin` — role `admin` only. */
export async function fetchAdminDashboard(signal?: AbortSignal): Promise<AdminDashboard> {
  return unwrapSingleEntity(await getApiV1DashboardAdmin(signal));
}

/** `GET /api/v1/dashboard/super-admin` — role `super_admin` only. */
export async function fetchSuperAdminDashboard(signal?: AbortSignal): Promise<SuperAdminDashboard> {
  return unwrapSingleEntity(await getApiV1DashboardSuperAdmin(signal));
}
