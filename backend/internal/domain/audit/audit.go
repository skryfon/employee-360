// Package audit centralizes the audit-log action and entity-type identifiers.
// Values are persisted in audit_logs and matched by the admin viewer, so they
// must never change. Actions follow the "entity.verb" convention.
package audit

// Entity types.
const (
	EntityDepartment   = "department"
	EntityPosition     = "position"
	EntityInvitation   = "user_invitation"
	EntityTenant       = "tenant"
	EntityTenantDomain = "tenant_domain"
)

// Actions.
const (
	ActionDepartmentCreate     = "department.create"
	ActionDepartmentUpdate     = "department.update"
	ActionDepartmentDelete     = "department.delete"
	ActionDepartmentActivate   = "department.activate"
	ActionDepartmentDeactivate = "department.deactivate"

	ActionPositionCreate     = "position.create"
	ActionPositionUpdate     = "position.update"
	ActionPositionDelete     = "position.delete"
	ActionPositionActivate   = "position.activate"
	ActionPositionDeactivate = "position.deactivate"

	ActionInvitationInvite = "invitation.invite"
	ActionInvitationResend = "invitation.resend"
	ActionInvitationRevoke = "invitation.revoke"

	ActionTenantRename       = "tenant.rename"
	ActionTenantDomainAdd    = "tenant.domain.add"
	ActionTenantDomainUpdate = "tenant.domain.update"
	ActionTenantDomainRemove = "tenant.domain.remove"
)
