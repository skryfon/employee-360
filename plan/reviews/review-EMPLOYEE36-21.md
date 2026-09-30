# Review: EMPLOYEE36-21 — D3 — Admin app: invitation-management UI

> Branch: ebin/feat/EPIC-D/EMPLOYEE36-21 | Last reviewed: 2026-09-30 23:10 | Iteration: 2 | Verdict: 🟢

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
`bfb6bd5` — docs: record roles route and admin shell in cycle-02 scope; update EMPLOYEE36-21 review (EMPLOYEE36-21)

Reviewed range: `d07fecf..ef78b6d` (this ticket's 10 commits). Local `origin/main` is stale (still lacks earlier merged work), so a `origin/main...HEAD` diff would pull in unrelated tickets; no `git fetch` was run. Tests were not re-run by the reviewer beyond typecheck/lint on an earlier commit; implementer reports api-client 74, admin 35, employee 3 passing.

## Findings

### 🔴 Critical
- [ ] (none)

### 🟡 Major
- [x] (resolved in bfb6bd5: documented in cycle-02 scope doc; commit not split to avoid history rewrite — accepted) Scope creep — `backend/internal/delivery/http/routes.go:81-89` (+ `handlers/role_handler.go`, `usecase/*/role/*`, `types/role/role.go`, container wiring) — a new backend endpoint `GET /api/v1/roles` lands inside a frontend ticket (D3). Cycle 2 does not list a roles-list route, and the backend work belongs to EPIC-C, not D3. The implementation itself is sound (Auth → Tenant → `RequireRole(admin, super_admin)`, tenant from ctx, `super_admin` filtered, tests at handler/usecase/route level). — Fix: either split the backend commit `1184329` into its own ticket/PR under EPIC-C (preferred) or add the endpoint to cycle-02's scope doc so the record matches.

### 🟢 Minor
- [x] (resolved in 8428881 + 508de5f) `backend/docs/swagger.json` / `packages/api-client/src/auth.ts:136-145` — handler has swagger annotations but the spec was not regenerated, so `listRoles` is hand-written instead of Orval-generated. — Fix: `make swagger && pnpm generate:api`, then replace the hand-written call.
- [x] (resolved in 508de5f) `clients/admin/src/features/invitations/components/ConfirmRevokeModal.tsx` — modal has no Escape-to-close, no focus trap, and no initial focus on the dialog. — Fix: add keydown handler + focus management.
- [x] (resolved in 508de5f) `clients/admin/src/features/invitations/components/InviteUserForm.tsx:55` — `mutation.isSuccess && "Invitation sent."` is effectively dead now that the create page navigates away on success. — Fix: drop it, or show a success notice on the list page.
- [x] (resolved in 508de5f) `invitations.test.tsx` — no test for the roles-load error state ("Could not load roles."), nor for resend/revoke failure messages. — Fix: add one test each.
- [ ] (still open, deferred: needs org-reference list endpoints, follow-up ticket) `InviteUserForm.tsx` — department/position are raw UUID text inputs (no lookup endpoints exist yet), which is unusable for real admins. Acceptable as a stopgap; track a follow-up ticket for org-reference lists.
- [x] (resolved in follow-up commit: exact messages asserted) `clients/admin/src/features/invitations/pages/invitations.test.tsx:139,150` — the resend/revoke failure tests assert on a loose regex alternation (`/Resend exploded|Could not resend invitation/`), so they pass whichever message is shown. — Fix: assert the exact expected message.

## Verdict
- **Score:** 98/100
- **Flag:** 🟢 Merge
- **Notes:** All three ACs remain met with implementation and test evidence. The iteration-1 Major (backend roles route inside a frontend ticket) is resolved by recording the route, the `/invitations` split and the admin shell in the cycle-02 scope doc, rather than splitting the commit; accepted. The swagger, modal accessibility, dead success alert and missing-test findings are all fixed and verified in the diff (generated `listRoles` wrapper, focus trap/Escape/focus-restore in `ConfirmRevokeModal`, four new tests). Two minors remain: raw UUID inputs for department/position (deferred to a follow-up ticket) and loose regex assertions in two failure tests. Clear for merge; consider squashing the width/Back-link restyle commits.

## Re-review Log
### Iteration 2 — 2026-09-30 — sha `bfb6bd5`
- **Resolved:** roles route scope creep (documented in cycle-02), stale swagger / hand-written `listRoles`, modal accessibility, dead "Invitation sent." alert, missing error-state tests
- **Still open:** raw UUID department/position inputs (deferred, follow-up ticket)
- **New issues:** loose regex assertions in the resend/revoke failure tests (🟢)
- **Score:** 90 → 98 (+8)
- **Verdict:** 🟢 Merge
