---
description: Implement a GitHub issue end-to-end — research, plan, branch, code, test, PR
argument-hint: <issue-number>
---

# /pick-up-issue

Take a GitHub issue number, understand it deeply, plan the implementation, and ship a production-ready PR. The command stops and asks for your approval at three gates: **after planning** (before any code changes), **before pushing** the branch, and **before opening** the pull request. Never push or open a PR without explicit approval.

If `$ARGUMENTS` is empty or not a positive integer, stop and ask for the issue number.

## Workflow

### 1. Fetch the issue and its context

Run all of these in parallel — they're independent reads:

- `gh issue view $ARGUMENTS --json number,title,body,state,labels,assignees,milestone,comments,closedByPullRequestsReferences`
- `gh issue view $ARGUMENTS --comments` (full comment thread, formatted)
- `git fetch origin && git log --oneline -20 origin/main` (ground truth for the base branch)
- `git status` and `git branch --show-current` (confirm clean working tree on `main`; if not, stop and ask before proceeding)

Reject early if:
- The issue is **closed** — confirm with the user before doing anything.
- The issue is **assigned to someone else** — confirm before claiming.
- A linked PR already exists in `closedByPullRequestsReferences` — surface it and ask whether to continue.

### 2. Understand the issue

Read every linked artefact before writing any plan:

- The issue body and full comment thread.
- Any URLs in the body — fetch them with `WebFetch` if they're public docs/specs. Skip authenticated systems (Linear, Jira, internal Confluence) unless an MCP tool is available.
- Cross-references to other issues/PRs — `gh issue view <n>` / `gh pr view <n>` for each.
- Repo conventions that constrain the implementation:
  - [CLAUDE.md](../../CLAUDE.md) — architecture, commands, conventions.
  - [.claude/rules/](../../.claude/rules/) — read every file relevant to the area being changed (e.g. `translators.md` for a translator issue).
  - [.golangci.yml](../../.golangci.yml) — lint rules that will block the PR.
  - `go.mod` — minimum Go version, dependency constraints.
  - `.github/workflows/ci.yml` — what CI will actually run.

Then survey the code area the issue touches: read the relevant packages end-to-end before planning, not just the files mentioned in the issue. If the issue requests a new translator, read at least two existing translator packages plus `internal/translate/prepare.go` and `internal/translate/resolver.go`. If it's a bug fix, locate the failing path and read the surrounding code so the fix isn't a band-aid.

### 3. Write the plan and get approval (Gate 1)

Use the `Plan` subagent for non-trivial work (new feature, new package, cross-cutting change). For a small bug fix, draft the plan inline.

The plan must include:

- **Acceptance criteria** restated in concrete terms — what input produces what output, what existing tests must still pass, what new tests will exist.
- **Files to add or modify**, each with one-line intent.
- **Public API changes** (exported types, function signatures, CLI flags) — call these out explicitly so the user can object before implementation.
- **Risk and reversibility** — does this change generated output for existing users? Is there a deprecation path? Does it touch the spec parser (`internal/opendpi/`) or session loader where regressions are silent?
- **Test plan** — list each test case with a one-line description. For translators, follow the matrix in [.claude/rules/translators.md](../../.claude/rules/translators.md): primitives, formats, refs, required-vs-optional, inline objects, plus any constraint-driven behavior.
- **Out of scope** — explicit list of things the issue might suggest but this PR will not do, with rationale.

Present the plan to the user. **Stop. Wait for explicit approval.** Iterate on the plan until the user agrees. Do not write code yet.

### 4. Branch

After plan approval, create the branch from `origin/main` (not local `main` — avoid stale base):

```
git fetch origin
git switch -c <type>/<issue-number>-<short-kebab-slug> origin/main
```

`<type>` follows Conventional Commits: `feat`, `fix`, `docs`, `chore`, `test`, `refactor`, `ci`, `build`, `perf`, `style`. `<short-kebab-slug>` is 3-6 words from the issue title. Example: `feat/100-iceberg-translator`.

If `make setup` hasn't been run in this clone, run it now so the `commit-msg` hook is active — Conventional Commits are enforced and a missing hook will let bad messages through.

### 5. Implement

Code to production standards. Non-negotiable:

- **License header** on every new Go file:
  ```
  // SPDX-License-Identifier: Apache-2.0
  // Copyright 2026 Daco Labs
  ```
- **Package comment** on the first file of each new package: `// Package x ...`.
- **No new public API surface outside `internal/`** unless the issue explicitly requires it.
- **Errors wrapped with `%w`** and lowercase strings (lint enforces `errorf` and `error-strings`).
- **No new dependencies** without checking `go.mod` and surfacing the addition to the user before adding.
- **Templates parsed once at init** with `template.Must(template.ParseFS(...))` — never per-call (per `.claude/rules/translators.md`).
- **Tests alongside code**, not after. Each behavior the plan calls out gets a test case. Table tests are the default shape — see `internal/translate/pyspark/translator_test.go`.
- **Deterministic output** — preserve property order (`Prepare` already does this; don't sort in templates), don't iterate Go maps in output paths.

After each meaningful chunk of work, run the full local check before moving on:

```
make format
make lint
make test
make build
```

If any of these fail, **fix the underlying issue** — never skip a hook with `--no-verify`, never `// nolint` to silence a real warning, never delete a failing test. If the failure suggests the plan was wrong, stop and revise the plan with the user.

### 6. Commit

Use Conventional Commits. Default to a single focused commit per logical step; squash later if needed. Message format:

```
<type>(<scope>): <imperative summary>

<body explaining why, not what — the diff already shows what>

Refs #<issue-number>
```

For the final commit on the branch, use `Closes #<issue-number>` so the PR auto-closes the issue on merge. **Stage specific files**, not `git add -A`. Commit messages must be passed via heredoc to preserve formatting (see CLAUDE.md). Never amend a published commit; create a new one if you need to fix something.

### 7. Push (Gate 2)

Before pushing, run the full check matrix one final time and report results:

- `make format` — diff must be empty after this.
- `make lint` — zero warnings.
- `make test` — all green, including `-race`.
- `make build` — binary produced.
- `git diff origin/main...HEAD --stat` — show the user what's about to be pushed.

**Stop. Confirm with the user before `git push -u origin <branch>`.** Pushing makes the branch visible on the remote and may trigger CI; do not push unattended.

### 8. Open the PR (Gate 3)

After the push succeeds, draft the PR title and body. **Show them to the user and wait for approval before running `gh pr create`.**

PR title: same Conventional-Commits prefix as the branch, under 70 chars. Example: `feat: add Apache Iceberg schema translator (#100)`.

PR body (use heredoc):

```markdown
## Summary
<2-4 bullets — what changed and why, in plain language>

## Closes
Closes #<issue-number>

## Implementation notes
<anything reviewers need to know that isn't obvious from the diff: design choices, trade-offs taken, things deliberately left out of scope>

## Test plan
- [x] `make lint` passes
- [x] `make test` passes (including `-race`)
- [x] `make build` produces a working binary
- [x] <feature-specific manual check, e.g. "ran `daco ports translate --format iceberg` against the example port and verified output against Iceberg spec V2">

## Out of scope
<bulleted list from the plan, copied verbatim>
```

After PR creation, return the PR URL to the user. Do not request reviewers, do not merge.

## Stop conditions

Halt and surface the situation to the user — do not improvise — when:

- Tests fail and the fix isn't obvious within one or two attempts.
- The plan turns out to be wrong mid-implementation.
- The issue requires a dependency upgrade, a breaking API change, or a spec version bump.
- An external system (CI, release pipeline, published artefacts) would be affected.
- The user's working tree had uncommitted changes when the command started.

## What this command never does

- Push without approval.
- Open a PR without approval.
- Skip hooks (`--no-verify`), suppress lint warnings to make the build green, or comment out failing tests.
- Force-push, amend published commits, or rewrite history on a shared branch.
- Add or upgrade dependencies without surfacing the change first.
- Touch unrelated files for "drive-by" cleanup.
- Mark the issue closed manually — let the merged PR do it via `Closes #N`.
