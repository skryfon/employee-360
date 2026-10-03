-- Revert adding description, is_active and unique index on (tenant_id, LOWER(name)) to positions.
DROP INDEX IF EXISTS uq_positions_tenant_id_name;

ALTER TABLE positions
    DROP COLUMN IF EXISTS is_active,
    DROP COLUMN IF EXISTS description;
