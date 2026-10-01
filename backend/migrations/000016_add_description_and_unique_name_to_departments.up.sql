-- Add description column and unique name per tenant for departments.
ALTER TABLE departments
    ADD COLUMN description TEXT;

CREATE UNIQUE INDEX uq_departments_tenant_id_name ON departments (tenant_id, name);
