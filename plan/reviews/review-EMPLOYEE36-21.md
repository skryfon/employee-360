# Review: EMPLOYEE36-21 — D3 — Admin app: invitation-management UI

> Branch: ebin/feat/EPIC-D/EMPLOYEE36-21 | Last reviewed: 2026-09-30 22:30 | Iteration: 1 (fixes applied, awaiting re-review) | Verdict: 🟡

## Ticket
**Identifier:** EMPLOYEE36-21
**State:** unstarted (Plane state group at fetch time)
**Link:** n/a (Plane response carried no URL)
**Cycle:** plan/cycles/cycle-02-auth-onboarding.md

### Description
`clients/admin/src/features/invitations/`. Depends on D1, D2 (needs an authenticated admin session), and EPIC-C's C2 (invitation routes).

Scope:
- Invite-user form — email + role select (+ optional department/position), calls `inviteUser`
- Invitations list — pending/accepted/revoked, calls `listInvitations` via TanStack Query
- Resend action (pending invitations only) and revoke action, with confirmation on revoke

### Acceptance Criteria
- [x] AC-1: Only reachable by an authenticated admin/super_admin session — impl `clients/admin/src/stores/authStore.ts:47` (`isAuthenticated` requires an admin role) used by `RequireAuth.tsx`, routes mounted under it in `App.tsx`; tests `invitations.test.tsx` "redirects unauthenticated users…" (x2) and `features/auth/pages/auth.test.tsx:135` (employee-role session is not authenticated)
- [x] AC-2: List reflects resend/revoke immediately (query invalidation) — impl `queries/invitationQueries.ts` `useInvalidatingMutation` (`onSuccess` → invalidate `['invitations']`); tests `invitations.test.tsx` "revokes after confirmation and refreshes the list", "resends and invalidates the list" (asserts a 2nd GET)
- [x] AC-3: Resend/revoke hidden for accepted/revoked — impl `components/InvitationsTable.tsx` (`pending && …`); test `invitations.test.tsx` "shows resend/revoke only for pending invitations"

## Latest commit reviewed
`ef78b6d` — style(admin): compact borderless Back link with icon (EMPLOYEE36-21)

Reviewed range: `d07fecf..ef78b6d` (this ticket's 10 commits). Local `origin/main` is stale (still lacks earlier merged work), so a `origin/main...HEAD` diff would pull in unrelated tickets; no `git fetch` was run. Tests were not re-run by the reviewer beyond typecheck/lint on an earlier commit; implementer reports api-client 74, admin 35, employee 3 passing.

## Findings

### 🔴 Critical
- [ ] (none)

### 🟡 Major
- [x] (resolved: documented in cycle-02 scope doc; commit not split to avoid history rewrite) Scope creep — `backend/internal/delivery/http/routes.go:81-89` (+ `handlers/role_handler.go`, `usecase/*/role/*`, `types/role/role.go`, container wiring) — a new backend endpoint `GET /api/v1/roles` lands inside a frontend ticket (D3). Cycle 2 does not list a roles-list route, and the backend work belongs to EPIC-C, not D3. The implementation itself is sound (Auth → Tenant → `RequireRole(admin, super_admin)`, tenant from ctx, `super_admin` filtered, tests at handler/usecase/route level). — Fix: either split the backend commit `1184329` into its own ticket/PR under EPIC-C (preferred) or add the endpoint to cycle-02's scope doc so the record matches.

### 🟢 Minor
- [x] (resolved in 8428881 + 508de5f) `backend/docs/swagger.json` / `packages/api-client/src/auth.ts:136-145` — handler has swagger annotations but the spec was not regenerated, so `listRoles` is hand-written instead of Orval-generated. — Fix: `make swagger && pnpm generate:api`, then replace the hand-written call.
- [x] (resolved in 508de5f) `clients/admin/src/features/invitations/components/ConfirmRevokeModal.tsx` — modal has no Escape-to-close, no focus trap, and no initial focus on the dialog. — Fix: add keydown handler + focus management.
- [x] (resolved in 508de5f) `clients/admin/src/features/invitations/components/InviteUserForm.tsx:55` — `mutation.isSuccess && "Invitation sent."` is effectively dead now that the create page navigates away on success. — Fix: drop it, or show a success notice on the list page.
- [x] (resolved in 508de5f) `invitations.test.tsx` — no test for the roles-load error state ("Could not load roles."), nor for resend/revoke failure messages. — Fix: add one test each.
- [ ] (deferred: needs org-reference list endpoints, follow-up ticket) `InviteUserForm.tsx` — department/position are raw UUID text inputs (no lookup endpoints exist yet), which is unusable for real admins. Acceptable as a stopgap; track a follow-up ticket for org-reference lists.

## Verdict
- **Score:** 90/100
- **Flag:** 🟡 Reviewer call
- **Notes:** All three ACs are met with implementation and test evidence, and no invariant violations were found: the frontend uses TanStack Query for server state, RHF+Zod for forms, the api-client for all HTTP, and Zustand only for UI/auth state. The new backend route is correctly protected and tenant-scoped. The score is 90, but the open Major (the backend roles endpoint living in a frontend ticket) puts the verdict at Reviewer call. Resolving it, by splitting or documenting it, plus the stale swagger would clear it. The branch also carries unrequested-looking style churn (several width/Back-link restyle commits); squash before merge if the history matters.

## Re-review Log
_(empty — first review)_
