-- Add description column and case-insensitive unique name per tenant for active (non-deleted) departments.
-- Precondition: Any pre-existing duplicate department names (case-insensitive) per tenant must be deduped before running this migration.
ALTER TABLE departments
    ADD COLUMN description TEXT;

CREATE UNIQUE INDEX uq_departments_tenant_id_name ON departments (tenant_id, LOWER(name)) WHERE deleted_at IS NULL;
