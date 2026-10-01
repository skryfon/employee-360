# Review: EMPLOYEE36-18 — EPIC-D: Cycle 2 — Auth & Onboarding Frontend (Overview)

> Branch: feature/EPIC-D | Last reviewed: 2026-10-01 | Iteration: 2 | Verdict: 🟢 Merge

## Ticket
**Identifier:** EMPLOYEE36-18 (epic-anchor mode; children EMPLOYEE36-19..23)
**State:** started
**Link:** Plane project 84ac89ce-4980-4bcb-946f-cd252071f83a, work item 1ff6aeaa-f5a8-4345-a687-a1852e710191
**Cycle:** plan/cycles/cycle-02-auth-onboarding.md (Frontend sub-feature)

### Description
Frontend consumers for EPIC-B (auth) and EPIC-C (invitations): admin login, employee login, forgot/reset password, admin invitation-management UI, invitation-accept page. No SSO/OAuth UI, no other module UI.

Review range: `15e2bb5..HEAD` (EPIC-D only; the diff against `origin/main` also contains the earlier EPIC-B/C work). Per-story reviews exist and are all 🟢: 19 (99), 20 (100), 21 (99), 22 (95), 23 (98).

### Acceptance Criteria
Epic Definition of Done:
- [x] DoD-1: Admin can log in (D2 — `clients/admin/src/features/auth/pages/LoginPage.tsx`; `auth.test.tsx`)
- [x] DoD-2: Admin can request/complete password reset (D2 — Forgot/ResetPasswordPage; `auth.test.tsx`)
- [x] DoD-3: Admin can invite a user and see it in the list (D3 — `features/invitations`; `invitations.test.tsx`)
- [x] DoD-4: Invitee (admin or employee) can set a password on the accept page (D5 — `AcceptInvitationPage.tsx` in both apps; both `auth.test.tsx`)
- [x] DoD-5: Employee can log in with email + password (D4 — `clients/employee/.../LoginPage.tsx`; employee `auth.test.tsx`)

[EMPLOYEE36-19] D1 api-client
- [x] AC-1: Methods typed against real handler shapes (`auth.ts`, generated models from swagger; `auth.test.ts`)
- [x] AC-2: Expired token refreshes once transparently (`client.ts`; `client.test.ts`, `refreshLock.test.ts`)
- [x] AC-3: No component/page builds its own Axios call (grep of `clients/*/src` finds none outside tests)

[EMPLOYEE36-20] D2
- [x] AC-1: Login redirects to shell and persists across reload (`auth.test.tsx` reload tests)
- [x] AC-2: Invalid login shows inline error without leaking email vs. password
- [x] AC-3: Forgot-password success state identical regardless of email

[EMPLOYEE36-21] D3
- [x] AC-1: Reachable only by authenticated admin/super_admin (`RequireAuth` + `ADMIN_ROLES` in `authStore.ts`; backend `RequireRole`)
- [x] AC-2: List reflects resend/revoke via query invalidation (`invitationQueries.ts`; `invitations.test.tsx`)
- [x] AC-3: Resend/revoke disabled/hidden for accepted/revoked (`InvitationsTable.tsx`; test)

[EMPLOYEE36-22] D4
- [x] AC-1: Login redirects into employee shell and persists across reload
- [x] AC-2: Invalid credentials show inline error, retry keeps email

[EMPLOYEE36-23] D5
- [x] AC-1: Invited admin sets password and can log in via D2
- [x] AC-2: Invited employee sets password and can log in via D4
- [x] AC-3: Expired/accepted/revoked token shows a distinct, actionable error

## Latest commit reviewed
`39a91fb` — refactor(ui): share auth components, schemas and accept-invitation page via packages/ui (EMPLOYEE36-18)

## Findings

### 🔴 Critical
- [ ] (none)

### 🟡 Major
- [ ] (none)

### 🟢 Minor
- [x] (resolved in 39a91fb: shared components, schemas and AcceptInvitationPage moved to `packages/ui`) `clients/admin/src/features/auth/pages/AcceptInvitationPage.tsx` vs `clients/employee/.../AcceptInvitationPage.tsx` — near-identical 251-line pages; `FormField`, `InlineAlert`, `SubmitButton`, `AuthLayout`, `authSchemas` and `authMutations` are also copied between the two apps. Only labels and `currentApp` differ. Drift risk (a fix to one app is easy to miss in the other). Extract the shared pieces into `packages/ui` or `api-client` once a third consumer appears (carried over from EMPLOYEE36-23).
- [ ] `clients/admin/src/features/invitations/components/InviteUserForm.tsx` — department/position are raw UUID text inputs. Not usable for a real admin. Add selects once the org-reference list endpoints exist (follow-up ticket; carried over from EMPLOYEE36-21). Author confirmed this will be fixed in an upcoming ticket, so it is accepted as deferred and non-blocking.
- [ ] History hygiene — EMPLOYEE36-21 has several width/Back-link restyle commits, and EMPLOYEE36-23 bundles backend validate endpoint and frontend in one commit. Consider squashing before the PR to main.

## Verdict
- **Score:** 98/100
- **Flag:** 🟢 Merge
- **Notes:** All five epic DoD items and all 16 child ACs are met, relying on the five per-story reviews (all 🟢) plus fresh epic-level verification. `pnpm -r typecheck` is clean, `pnpm -r test` passes (api-client 83, admin 55, employee 42), and `go build ./...` and `go test ./internal/...` pass. Invariants hold. The tenant comes from the login response and is never client-derived, and the backend does not trust `X-Tenant-ID`. Bearer auth keeps the API headless. `/roles` and the invitation management routes carry `RequireRole(admin, super_admin)` on top of tenant middleware. The frontend uses feature folders, TanStack Query for server state, Zustand only for the session mirror and UI state, RHF + Zod forms, and no raw axios/fetch outside `packages/api-client`. The backend `/roles` and `validate` additions inside a frontend epic are recorded in the cycle-02 scope doc, so they are not scope creep. Only minor duplication and hygiene nits remain. I did not re-run browser or e2e checks of the cross-app redirect on accept; it is covered by unit tests only.

## Re-review Log
### Iteration 2 — 2026-10-01 — sha `39a91fb`
- **Resolved:** minor #1 (admin/employee duplication) — shared components, schemas and `AcceptInvitationPage` now in `packages/ui`; the apps keep thin wrappers, so stores and login mutations stay per app. `packages/ui` never reads `import.meta.env`; props supply URLs and copy. Both apps' `index.css` add `@source` for the package (Tailwind v4).
- **Still open:** UUID inputs for department/position (deferred to an upcoming ticket); history hygiene nit.
- **New issues:** none. Typecheck, lint and tests pass: api-client 83, ui 1, admin 55, employee 42. The apps' `auth.test.tsx` are unchanged. The new shared page has only one smoke test, but the full accept flow is still covered by the apps' tests.
- **Score:** 97 → 98 (+1)
- **Verdict:** 🟢 Merge
