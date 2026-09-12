---
name: team-mate-review
description: Pre-merge code review of a teammate branch against `main`, anchored to a Plane ticket (e.g. PROJ-N) or epic (EPIC-N). Fetches ticket/epic via Plane MCP, diffs branch vs main, runs structured review (acceptance criteria, code quality, logic, security, cycle-scope adherence, project invariants, tests). Epic-anchor mode lists the epic's child stories and reviews against the union of their ACs. Local mode persists findings to plan/reviews/review-{ticket}.md; CI mode (--ci --pr N) reads/writes state via PR comments instead. Re-reviews load prior context from the file (local) or the most recent review comment (CI). Use whenever the user wants to review a teammate's PR or branch tied to a Plane ticket, validate fixes on a re-pushed branch, or get a merge verdict with a score. Trigger phrases: "review PROJ-N", "review teammate branch/PR", "re-review PROJ-N", "check this branch against the ticket", "code review for ticket".
allowed-tools: Bash, Read, Write, Edit, Grep, Glob, AskUserQuestion, mcp__plane__retrieve_work_item_by_identifier, mcp__plane__list_work_items, mcp__plane__list_work_item_comments, mcp__plane__list_work_item_properties
---

# Team-mate Review Skill

Structured code review for teammate branches tied to Plane tickets or epics.

- **Local mode** (default): persists findings to `plan/reviews/review-{TICKET}.md` so re-reviews across sessions are low-effort.
- **CI mode** (`--ci --pr PR_NUMBER`): reads/writes state via PR comments — no file I/O. The PR comment thread is the re-review log.

## Invocation

```
/team-mate-review [TICKET] [--re-review] [--ci --pr PR_NUMBER]
```

- `TICKET` — Plane ticket (e.g. `PROJ-42`) or epic (`EPIC-3`) identifier, using whatever prefix this repo's Plane project actually uses. Optional if already on the feature branch.
- `--re-review` — validate fixes from a prior review; skip if doing first review.
- `--ci --pr PR_NUMBER` — CI mode. Must be paired. Replaces file I/O with PR comment I/O.
  In CI, `--re-review` loads prior context from the most recent review comment on the PR.
  If no prior comment exists, falls back to a fresh review automatically.

---

## Initial Review

### Step 0 — Resolve ticket + target branch

Run `git rev-parse --abbrev-ref HEAD` to get the current branch.

Try to match current branch against these patterns (in order):

1. **Story branch**: `^([^/]+)/(feat|fix|task)/(?:EPIC-[A-Z0-9]+/)?([A-Z]+-\d+(?:-\d+)*)$` — capture group 3 is the ticket id (e.g. `PROJ-42`).
2. **Epic branch**: `^([^/]+)/(feat|fix|task)/(EPIC-[A-Z0-9]+)$` — capture group 3 is the epic id (e.g. `EPIC-3`).

If this repo's actual branch-naming convention differs (no convention is fixed yet in `CLAUDE.md`), ask once and remember the answer for the rest of the session rather than guessing silently.

Resolution order:
1. Ticket arg provided + current branch matches → stay, use ticket arg.
2. No ticket arg + current branch matches → derive id from branch.
3. Ticket arg provided but current branch doesn't match → run `git fetch --all`, then `git branch -r | grep -E "/(feat|fix|task)/(EPIC-[A-Z0-9]+/)?${ID}$"`. If a remote branch is found, inform the user which branch will be checked out and confirm before running `git checkout -b <local-name> <remote>`.
4. Still unresolved → `AskUserQuestion`: "Which branch should I review? (e.g. `<you>/feat/EPIC-3/PROJ-42` or `<you>/feat/EPIC-3`)"

**In `--ci` mode:** trust the passed ticket/epic id verbatim, review the current HEAD, skip both the regex match and any branch checkout. If the id is absent, fail with a non-zero exit — there is no interactive user to ask.

### Step 1 — Fetch Plane ticket (or epic + children)

**Story mode** (id does NOT start with `EPIC-`):

Call `mcp__plane__retrieve_work_item_by_identifier` with the id.

Also fetch:
- `mcp__plane__list_work_item_comments` — reviewer comments already on the ticket.
- `mcp__plane__list_work_item_properties` — acceptance criteria often live here as custom properties.

Extract: ticket title, description, acceptance criteria (numbered list), ticket state.

---

**Epic-anchor mode** (id starts with `EPIC-` or the retrieved work item's type is Epic):

1. Call `mcp__plane__retrieve_work_item_by_identifier` with the epic id to get its title, description, and `project_id`.
2. Fetch child stories: `mcp__plane__list_work_items(project_id, pql='childOf("{TICKET}")')` (using the epic's human-readable identifier).
3. For each child story, call `mcp__plane__list_work_item_properties` to gather its ACs.
4. Build a unified AC set, **grouped per child story** (e.g. `[PROJ-42] AC-1: ...`).

The AC hard-gate in Step 5 requires **every AC across every child story** to be met.

---

If the Plane MCP tool errors or returns an auth failure (e.g. the `plane` MCP server hasn't been signed into yet — run `/mcp` to authenticate), call `AskUserQuestion`:
> "Plane MCP isn't responding. Please either authenticate via `/mcp`, or paste the ticket/epic details manually (title, description, acceptance criteria)."

### Step 2 — Diff vs main

```bash
git fetch origin main
git log origin/main..HEAD --oneline
git diff origin/main...HEAD --stat
git diff origin/main...HEAD
```

Record the current HEAD sha — this becomes `Latest commit reviewed` in the handoff.

If the diff is very large (> 2000 lines), read changed files individually rather than dumping the whole diff to context.

### Step 3 — Review

Apply the checklist below to every changed file. For each finding record `file:line — what — why — suggested fix`.

#### Checklist

**1. Acceptance criteria coverage**
Map each AC from Step 1 to **two pieces of evidence** in the diff:
- (a) Implementation evidence — file:line where the behaviour is implemented.
- (b) Test evidence — file:line of a test that specifically exercises the AC scenario.

An AC is **only met** when BOTH (a) and (b) exist. Mark the AC checkbox `[x]` only when both are present.

Rules:
- If implementation exists but no test covers the specific AC scenario → AC is **unmet** → 🔴 Critical.
- If an AC describes a specific edge case, a generic happy-path test does NOT satisfy it — the test must exercise that exact scenario.
- If an AC requires pre-existing behaviour with no diff change → mark met only if you can confirm an existing test covers it; otherwise treat as unmet → 🔴 Critical.

**2. Cycle-scope adherence** (project-specific — see `CLAUDE.md`'s "Execution Model: Cycles")
- Identify which `plan/cycles/cycle-NN-<name>.md` this ticket belongs to (ask if unclear).
- Check the diff against that cycle's declared sub-features. Work that belongs to a *later*, not-yet-started cycle (e.g. handlers/usecases landing during a migrations-only cycle) is scope creep → 🟡 Major, note which later cycle it actually belongs to.

**3. Project invariants** (non-negotiable per `CLAUDE.md`)
- **Multi-tenancy**: every tenant-owned table row (all except `tenants`) carries `tenant_id`; every persistence query/mutation scopes by `tenant_id` pulled from context — never a client-supplied value (body/query param/header). Any repo method that accepts a tenant ID as a parameter instead of reading it from context is 🔴.
- **Headless/client-agnostic API**: no web-only auth mechanism as the *sole* transport (e.g. httpOnly-cookie-only auth) — bearer-token auth must keep working. New endpoints live under `/api/v1/...`.
- **Self-hostable**: no new hard dependency on a cloud-provider-specific managed service (e.g. a proprietary queue/email API with no self-hosted equivalent) without flagging it.
- **Admin-only routes**: must have an explicit role check *in addition to* tenant-scoping middleware — tenant scoping alone is not authorization. Missing role check on an admin route is 🔴.

**4. Backend layering** (see `.claude/agents/backend-agent.md`)
- Flow: `infrastructure → delivery → usecase → domain`. No cross-layer leaks (handler calling a repo directly; a usecase importing Gin/GORM types; domain importing anything external).
- Handlers are thin: bind/validate → call usecase → respond via `internal/delivery/http/response/response.go` helpers. Raw `c.JSON(...)` in a handler is 🟡.
- DI container: constructor injection via `internal/infrastructure/container/`; no global singletons mid-request.
- Admin mutations should write an audit log entry via the `audit` usecase — a missing audit call on a new admin mutation is 🟡.

**5. Frontend conventions** (see `.claude/agents/frontend-agent.md`)
- Feature-folder structure respected (`components/`, `pages/`, `queries/`, `schemas/`, `routes.tsx`).
- TanStack Query for server state; no raw `fetch`/`axios` in a feature — must go through `packages/api-client`.
- Zustand holds only ephemeral client/UI/auth state — server data duplicated into a Zustand store is 🔴.
- React Hook Form + Zod for forms.
- Tenant is resolved once at login and never re-derived/switched mid-session — UI that lets a user switch tenants without re-authenticating is 🔴.

**6. Code quality**
- Clear naming; no leftover debug code or TODOs without ticket reference.
- Error handling at system boundaries (user input, external APIs); not invented deep in pure logic.
- No unused imports or dead code paths.

**7. Logical correctness**
- Edge cases: empty collections, nil pointers, zero values, concurrent access on shared mutable state.
- Off-by-one risks in pagination, indexing, year-boundary logic (holiday year-wise calendars).
- Correct handling of Go errors (not swallowed silently).

**8. Security**
- All SQL via GORM parameterized queries — no string concatenation.
- Auth/tenant/role middleware applied to every new route.
- No secrets or tokens in source; use env/config (Viper).
- **Tenant isolation**: for any new repo method, verify it's impossible to read/write another tenant's row even with a crafted request — this is the single most important thing to check.
- Admin views: XSS risk for rendered user content.

**9. Tests**
- Each acceptance criterion has at least one automated test that covers the specific scenario described. "At least one test exists for this usecase" is not sufficient — the test must target the AC's exact scenario (see checklist §1).
- Happy path + at least one error/edge case per critical function.
- New routes have integration tests; new usecases have unit tests with mock repos.
- New repository implementations should have a tenant-isolation test (writing under tenant A's context must never be readable under tenant B's context).
- Note: a missing test for an AC is classified 🔴 Critical (unmet AC), not 🟡 Major.

**10. Migrations** (if any)
- New migration file follows the project's sequential 6-digit naming (`create-migration` skill) and has both `up`/`down`.
- Every tenant-owned table has `tenant_id` + an index on it; every FK column has an index.
- No `DROP COLUMN` / destructive change without an explanation of why reversal is impossible in the down migration.
- The branch must not modify an already-committed migration file (the `check-migration.sh` hook should have blocked this at edit time — if it slipped through anyway, flag as 🔴).

### Step 4 — Classify and summarize findings

Group findings into three buckets:

**🔴 Critical — block merge** Any of: security vulnerability, tenant-isolation leak, acceptance criterion unmet (including missing test for a specific AC scenario), project invariant violated, data corruption risk.

**🟡 Major — fix strongly recommended** Any of: missing tests for AC, edge case that plausibly causes bugs in production, layering breach, cycle-scope creep, migration risk, missing audit log call.

**🟢 Minor — nit** Naming, comments, style, non-blocking suggestions.

Write a short rationale for every finding. Be specific: state the file and line, describe the actual risk, and propose the fix.

### Step 5 — Score and verdict

**Hard gate (check first, before scoring):**
Count the AC checkboxes from Step 3 checklist §1. If ANY AC is still `[ ]` (unchecked), the verdict is **immediately 🔴 Block** — do not compute the score-based verdict. State which ACs are unmet and stop at the verdict.

In epic-anchor mode, all ACs across all child stories must be `[x]`.

Compute score deterministically (only when all ACs are `[x]`):
- Start at **100**.
- Each 🔴 finding: **−20**.
- Each 🟡 finding: **−5**.
- Each 🟢 finding: **−1**.
- Clamp to [0, 100].

Verdict (score-based, only reached when all ACs met):
| Condition | Flag |
|---|---|
| Score ≥ 85 AND 0 critical | 🟢 **Merge** |
| Score 60–84 OR any major open | 🟡 **Reviewer call** |
| Score < 60 OR ≥1 critical | 🔴 **Block** |

Note: "all AC met" is now a precondition for reaching the score table, not a column in it. An unmet AC bypasses scoring entirely → 🔴 Block.

### Step 6 — Persist findings

**Local mode** (default):

Path: `plan/reviews/review-{TICKET}.md`

If the `plan/reviews/` directory doesn't exist, create it first (`mkdir -p plan/reviews`).

On initial review, **overwrite** the file with the full template below. Populate every section from what you gathered in Steps 0–5.

```markdown
# Review: {TICKET} — {title}

> Branch: {branch} | Last reviewed: {YYYY-MM-DD HH:MM} | Iteration: 1 | Verdict: {🟢/🟡/🔴}

## Ticket
**Identifier:** {TICKET}
**State:** {ticket state}
**Link:** {URL if available from Plane response}
**Cycle:** {plan/cycles/cycle-NN-<name>.md this ticket belongs to}

### Description
{description}

### Acceptance Criteria
- [ ] AC-1: ...
- [ ] AC-2: ...

## Latest commit reviewed
`{sha}` — {commit subject line}

## Findings

### 🔴 Critical
<!-- Each entry: file:line — what — why — suggested fix -->
- [ ] (none)

### 🟡 Major
- [ ] (none)

### 🟢 Minor
- [ ] (none)

## Verdict
- **Score:** {N}/100
- **Flag:** {🟢 Merge | 🟡 Reviewer call | 🔴 Block}
- **Notes:** {one paragraph summary of the overall review}

## Re-review Log
_(empty — first review)_
```

---

**CI mode** (`--ci --pr PR_NUMBER`):

Post a PR comment instead of writing a file. The comment body uses the same template **minus** the Re-review Log section (the PR comment thread is the log). Prefix the body with an HTML sentinel so later runs can find it:

```bash
gh pr comment {PR_NUMBER} --body "<!-- team-mate-review: {TICKET} -->
# Review: {TICKET} — {title}

> Branch: {branch} | Reviewed: {YYYY-MM-DD HH:MM} | Iteration: 1 | Verdict: {flag}

## Ticket
...

### Acceptance Criteria
...

## Latest commit reviewed
\`{sha}\` — {commit subject line}

## Findings

### 🔴 Critical
...

### 🟡 Major
...

### 🟢 Minor
...

## Verdict
- **Score:** {N}/100
- **Flag:** {flag}
- **Notes:** {summary}
"
```

### Step 7 — Stop

Output a brief summary:
- Branch reviewed, commit sha, score, flag.
- Count of 🔴/🟡/🟢 findings.
- **Local:** path to handoff file. **CI:** "Posted review comment on PR #{PR_NUMBER}."

Then stop. Do not proceed to re-review steps.

---

## Re-review (`--re-review`)

### Step R0 — Load prior context

**Local mode:**

Read `plan/reviews/review-{TICKET}.md`.

If the file doesn't exist, stop with:
> "No prior review found for {TICKET}. Run without `--re-review` first."

---

**CI mode (`--ci --re-review --pr PR_NUMBER`):**

Fetch the most recent review comment for this ticket:

```bash
gh pr view {PR_NUMBER} --json comments \
  --jq '[.comments[] | select(.body | startswith("<!-- team-mate-review: {TICKET} -->"))] | last | .body'
```

If the output is empty or null → no prior review exists → **fall back to a fresh review** (run Steps 1–6 as if `--re-review` was not passed). Do not error.

---

From whichever source, extract:
- `Latest commit reviewed` sha (`prev_sha`).
- All open (unchecked `- [ ]`) findings by category.
- Acceptance criteria checklist state.
- Previous score and verdict.
- Current iteration number.

### Step R1 — Diff since last review

```bash
git fetch origin
git log {prev_sha}..HEAD --oneline
git diff {prev_sha}..HEAD --stat
git diff {prev_sha}..HEAD
```

If `git log {prev_sha}..HEAD` is empty, the branch hasn't changed since last review. Tell the user and stop.

### Step R2 — Validate prior findings

For each open finding from the prior context, check whether the new diff addresses it:
- Find the file:line in the new diff.
- Assess if the issue is resolved, partially fixed, or unchanged.
- Mark accordingly in your working notes (you'll update the output in Step R5).

Also re-validate each AC against the new commits.

### Step R3 — Review new commits only

Apply the Step 3 checklist to the new diff (`{prev_sha}..HEAD`). Don't re-scan unchanged files — focus on what changed. Identify any new 🔴/🟡/🟢 findings introduced since last review.

### Step R4 — Recompute score and verdict

Build the full current open-findings set:
- Prior findings that are **still open** (carry forward).
- **New findings** from Step R3.
- **Resolved** findings are removed.

**Hard gate (same as Step 5):** If any AC checkbox is still `[ ]`, verdict is immediately 🔴 Block. Do not proceed to score computation.

Apply the same scoring rubric as Step 5. Compute new score and verdict.

### Step R5 — Update handoff

**Local mode:**

**Do not overwrite** — surgically update:
1. Update header line (`Last reviewed`, `Iteration` +1, `Verdict`).
2. Update `Latest commit reviewed` to new HEAD sha.
3. For each resolved finding: change `- [ ]` → `- [x]` and append `(resolved in {sha})`.
4. Append any new findings to the appropriate category section.
5. Update the `Verdict` block with new score, flag, and notes.
6. Append a new `### Iteration {N} — {date} — sha \`{new sha}\`` entry to the Re-review Log:

```markdown
### Iteration {N} — {YYYY-MM-DD} — sha `{new sha}`
- **Resolved:** {list of fixed items, or "none"}
- **Still open:** {list, or "none"}
- **New issues:** {list, or "none"}
- **Score:** {prev} → {new} ({delta:+d})
- **Verdict:** {flag}
```

---

**CI mode:**

Post a **new** PR comment with the full current review state (same template as Step 6 CI mode, iteration N+1). Do not attempt to edit the old comment — the PR comment thread is the iteration history.

```bash
gh pr comment {PR_NUMBER} --body "<!-- team-mate-review: {TICKET} -->
# Review: {TICKET} — {title}

> Branch: {branch} | Reviewed: {YYYY-MM-DD HH:MM} | Iteration: {N+1} | Verdict: {flag}

## Iteration summary
- **Resolved:** {list or "none"}
- **Still open:** {list or "none"}
- **New issues:** {list or "none"}
- **Score:** {prev} → {new} ({delta:+d})

## Latest commit reviewed
\`{sha}\` — {commit subject line}

## Findings (open only)

### 🔴 Critical
...

### 🟡 Major
...

### 🟢 Minor
...

## Verdict
- **Score:** {N}/100
- **Flag:** {flag}
- **Notes:** {summary}
"
```

### Step R6 — Output and verdict

Output:
- Commit range reviewed (`{prev_sha}..HEAD`).
- What was resolved vs still open.
- Any new issues found.
- Final score + verdict.
- **Local:** path to updated handoff file. **CI:** "Posted iteration {N} comment on PR #{PR_NUMBER}."

If verdict is 🟢 → note the branch is clear for merge. If 🟡 or 🔴 → the handoff is updated for the next round.

---

## Key files to read during review

These project rule files inform what's "correct" for this codebase. Read the relevant one(s) based on what the diff touches:

| Area | File |
|---|---|
| Project invariants (multi-tenancy, headless, self-hosting) | `CLAUDE.md` |
| Cycle scope (what's actually in scope right now) | `plan/cycles/cycle-NN-<name>.md` for the ticket's cycle |
| Go backend architecture, layering, DI, migrations | `.claude/agents/backend-agent.md` |
| Frontend conventions, TanStack Query, feature-folder | `.claude/agents/frontend-agent.md` |
| Full backend architecture reference | `plan/architecture/backend.md` |
| Full frontend architecture reference | `plan/architecture/frontend.md` |

Read lazily — only if the diff touches that layer. A pure backend change doesn't need `frontend-agent.md`.
