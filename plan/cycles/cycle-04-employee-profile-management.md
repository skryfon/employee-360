# Cycle 4 — Employee Profile & Employee Management

| | |
|---|---|
| **Status** | Planned |
| **Module** | Employee Profile + Employee Management (admin) — full stack |
| **Depends on** | `plan/cycles/cycle-02-auth-onboarding.md` (tenants, users, roles, DB-verified auth middleware, audit columns from migration 000015, invitation UI/shell in `clients/admin`) |
| **Source** | Request for employee personal/employment/address/emergency-contact data and an admin page to manage employees; patterns from migrations 000005, 000011, 000015 |

This is the scope/status doc for Cycle 4. Read this before starting or resuming work.
Cross-cutting principles live in `CLAUDE.md`.

**Scope note:** migrations first, then backend API (`backend-agent`), then admin frontend
(`frontend-agent`). The employee self-service portal (`clients/employee`) is **not** in
this cycle. Migration number is the next free one (`000018` at time of writing — the repo
has no `000016`); independent of Cycle 3 (Holiday Calendar).

---

## Objective

Give tenant admins and super admins one place to see and manage the people in their
tenant: list users/employees, view and update their profile details, and activate /
deactivate them. Requires a tenant-scoped schema for employee details, which `users`
(identity/auth fields only) doesn't have today.

---

## Sub-Features

### 1. Migration `000018_create_employee_profiles` (up + down, via `create-migration` skill)

**`employee_profiles`** — 1:1 with `users`
- [ ] `tenant_id` FK (`ON DELETE CASCADE`) + index; `user_id` FK (`ON DELETE CASCADE`) + index
- [ ] Personal: `employee_code`, `gender`, `date_of_birth`, `blood_group`, `marital_status`,
      `nationality`, `phone`, `personal_email` (CHECK constraints on enumerated values)
- [ ] Employment: `joining_date`, `employment_type`, `employment_status` (default `active`),
      `probation_end_date`, `confirmation_date`, `exit_date`, `work_location`,
      `manager_id` → `users` (`ON DELETE SET NULL`, indexed)
- [ ] Address: `address_line1`, `address_line2`, `city`, `state`, `country`, `pincode` (TEXT)
- [ ] Audit columns inline: `created_at`, `updated_at`, `created_by`, `updated_by`,
      `deleted_by`, `deleted_at` (actor columns plain UUID, no FK — per 000015)
- [ ] Partial unique indexes (`WHERE deleted_at IS NULL`): `(tenant_id, user_id)`;
      `(tenant_id, employee_code)` where code is not null

**`emergency_contacts`** — many per employee
- [ ] `tenant_id` FK + index; `user_id` FK + index
- [ ] `name`, `relationship`, `phone` (NOT NULL); `alternate_phone`, `email`
- [ ] `is_primary` + partial unique index: one primary per `(tenant_id, user_id)` among
      non-deleted rows
- [ ] Audit columns inline

**Down:** drop `emergency_contacts`, then `employee_profiles` (indexes first).

### 2. Backend API (`backend-agent`, `new-backend-feature` skill) — `/api/v1`, roles `admin` + `super_admin`

All endpoints derive `tenant_id` from auth context (never from payload); usecases take
explicit `tenantID`/actor params; handlers never touch GORM.

- [ ] `GET /employees` — paginated list of tenant users with profile summary; search
      (name/email/employee code), filters (status active/inactive, department, position,
      employment status); excludes soft-deleted
- [ ] `GET /employees/:id` — user + profile + emergency contacts
- [ ] `PUT/PATCH /employees/:id` — update user fields (name, department, position) and
      profile fields; upsert profile on first update; validate department/position/manager
      belong to the tenant
- [ ] `PUT /employees/:id/emergency-contacts` (or sub-resource CRUD) — manage contacts,
      enforce single primary
- [ ] `POST /employees/:id/deactivate` and `/activate` — flips `users.is_active`; revoke
      refresh tokens on deactivate; cannot deactivate yourself or the last active admin;
      `admin` cannot act on `super_admin`
- [ ] Role assignment: change an employee's role between `employee` / `admin`
      (super_admin only for granting `admin`) — confirm exact rules before building
- [ ] Entities, repository ports + GORM adapters, handler/usecase/repo tests (incl.
      cross-tenant isolation and soft-delete read tests), swagger regeneration

### 3. Admin frontend (`frontend-agent`, `new-frontend-feature` + `design-system` skills) — `clients/admin/src/features/employees/`

- [ ] Sidebar nav item "Employees" (admin + super_admin only) and routes
- [ ] Employees list page: table, search, filters, pagination, status badge, row actions
- [ ] Employee detail/edit page: personal, employment, address sections; emergency
      contacts editor; RHF + Zod schemas
- [ ] Activate / deactivate with confirm dialog; role change control
- [ ] TanStack Query hooks in `queries/`; Axios client additions in `packages/api-client`;
      component + page tests

### 4. Employee creation + self-completed profile (Plane EMPLOYEE36-31, sub-ticket of EMPLOYEE36-30)

Two-step setup: admin creates the user with **employment details only**; the employee
fills in the rest themselves.

- [ ] Admin: create employee with employment details (employee_code, joining_date,
      employment_type/status, department, position, manager, work_location, probation /
      confirmation dates) → user + `employee_profiles` row with personal fields NULL
- [ ] Employee: `GET/PATCH /me/profile` + emergency contacts CRUD — basic details
      (phone, personal email, DOB, marital status, nationality), gender, blood group,
      address, emergency contact(s); user ID from auth context only
- [ ] Employment fields read-only for the employee, editable by admin only
- [ ] `clients/employee` profile page (RHF + Zod) with "profile incomplete" prompt
- [ ] Open: which self-filled fields are mandatory vs optional

### Out of scope (deliberately)
- Government IDs (PAN/Aadhaar/SSN), bank details, salary — sensitive PII/compensation;
  belongs to the Salary/Taxation module with its own access rules.
- Employee self-service beyond the onboarding profile completion in sub-feature 4
  (no leave, documents, or other self-service).
- Bulk import/export, org chart, department/position CRUD.
- Hard delete / purge of users.

**Done when:** migrations apply and revert cleanly; an admin can list, search, filter,
view, update, activate and deactivate employees in their tenant through the admin UI;
a deactivated user can no longer authenticate (DB-verified middleware → 401); no endpoint
returns or mutates another tenant's data (covered by tests);
`make test` and `pnpm --filter admin test` + `pnpm typecheck` pass.
