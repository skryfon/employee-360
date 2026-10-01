-- Add audit/soft-delete columns (created_by, updated_by, deleted_by, deleted_at)
-- to tables that lack them. Actor columns are plain nullable UUIDs with NO
-- foreign key (bootstrap/system/pre-auth writes have no user row, and an FK to
-- users would block hard-purging users). Skipped on purpose: audit_logs
-- (append-only), refresh_tokens / password_reset_tokens (ephemeral credential
-- rows, system-generated and purged), river_* (third-party schema).
-- user_roles is a join table: only created_by/updated_by (rows are hard-deleted;
-- revocation is recorded in audit_logs).

ALTER TABLE tenants
    ADD COLUMN created_by UUID,
    ADD COLUMN updated_by UUID,
    ADD COLUMN deleted_by UUID,
    ADD COLUMN deleted_at TIMESTAMPTZ;

ALTER TABLE tenant_domains
    ADD COLUMN created_by UUID,
    ADD COLUMN updated_by UUID,
    ADD COLUMN deleted_by UUID,
    ADD COLUMN deleted_at TIMESTAMPTZ;

ALTER TABLE departments
    ADD COLUMN created_by UUID,
    ADD COLUMN updated_by UUID,
    ADD COLUMN deleted_by UUID,
    ADD COLUMN deleted_at TIMESTAMPTZ;

ALTER TABLE positions
    ADD COLUMN created_by UUID,
    ADD COLUMN updated_by UUID,
    ADD COLUMN deleted_by UUID,
    ADD COLUMN deleted_at TIMESTAMPTZ;

ALTER TABLE users
    ADD COLUMN created_by UUID,
    ADD COLUMN updated_by UUID,
    ADD COLUMN deleted_by UUID,
    ADD COLUMN deleted_at TIMESTAMPTZ;

ALTER TABLE roles
    ADD COLUMN created_by UUID,
    ADD COLUMN updated_by UUID,
    ADD COLUMN deleted_by UUID,
    ADD COLUMN deleted_at TIMESTAMPTZ;

ALTER TABLE user_invitations
    ADD COLUMN created_by UUID,
    ADD COLUMN updated_by UUID,
    ADD COLUMN deleted_by UUID,
    ADD COLUMN deleted_at TIMESTAMPTZ;

ALTER TABLE user_roles
    ADD COLUMN created_by UUID,
    ADD COLUMN updated_by UUID;

-- Users are soft-deleted (revoking an invitation removes the pending user), so
-- email uniqueness must only apply to live rows or the email could never be re-invited.
ALTER TABLE users DROP CONSTRAINT uq_users_tenant_id_email;
CREATE UNIQUE INDEX uq_users_tenant_id_email ON users (tenant_id, email) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_deleted_at ON users (deleted_at);
