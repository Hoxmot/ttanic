---
name: work-issue
description: Implement a ttanic GitHub issue end-to-end with the operator in the loop — read the issue, branch, plan, wait for ack, implement, open a PR, iterate on review. Use when asked to work on, pick up, or implement an issue (e.g. "work on issue 13", "/work-issue 13").
---

# Working a ttanic issue

Contract with the operator: **plan before code, explicit ack before implementation, operator merges — never you.** Never commit to `main`. `AGENTS.md` rules apply throughout.

**AI attribution**: anything you post to GitHub appears under the operator's account, so every comment (PR comments, issue comments, inline review replies) must state it is AI-generated. Open each one with the following, substituting `<MODEL>` with the model you are running as (e.g. Claude Fable 5):

```
> [!NOTE]
> 🤖 Beep bop boop — it's <MODEL>. This comment is AI-generated via [Claude Code](https://claude.com/claude-code); the operator stays at the helm (and the merge button).
```

(PR bodies carry the `🤖 Generated with [Claude Code](https://claude.com/claude-code)` footer instead, and commits the `Co-Authored-By` trailer, per AGENTS.md.)

Argument: an issue number (e.g. `13`). If none given, ask which issue to work on.

## 1. Understand

- `gh issue view <n> --comments` — the body has Context / Needs / Scope / Done when.
- Read the `docs/ttanic-lld.md` (and HLD if referenced) sections the issue points at.
- **Dependency gate**: for every task ID in **Needs**, verify that issue is closed with its PR merged (`gh issue view <dep>`). If any dependency is open, STOP and report which — do not start.
- Check nobody else is on it: `gh pr list --search "<n> in:body"` for an existing PR referencing the issue.

## 2. Branch

```
git fetch origin
git switch -c m<X>-<Y>-<slug> origin/main   # e.g. m1-6-archive-create for issue #6 (task M1.6)
```

## 3. Plan — and stop

Write an implementation plan covering:

- **Approach**: how you'll satisfy the issue's Scope, referencing the LLD design (types, functions, package boundaries you'll add or touch).
- **Files**: what gets created/modified.
- **Test plan**: mapped item-by-item to the issue's Done-when criteria.
- **Judgment calls / open questions**: anything the issue or LLD leaves ambiguous, with your proposed resolution.
- **Doc impact**: any HLD/LLD deviation you foresee (updated in this same branch if so).

Present the plan in the conversation. Do **not** commit it and do **not** write implementation code yet. STOP and wait for the operator's review. Address their comments by revising the plan and re-presenting. Only proceed on an explicit ack ("ack", "approved", "go ahead").

## 4. Implement

- Follow the acked plan; if reality forces a change to it, say so and get a nod before continuing down a different path.
- Scope discipline: build what Done-when requires, nothing more. Problems you spot along the way become new issues (`gh issue create`), not drive-by fixes.
- Commit style per AGENTS.md (conventional prefixes). Small, coherent commits.
- Gate before pushing: `just ci` passes locally (before issue #1 lands the justfile: `go build ./... && go vet ./... && go test ./...`).

## 5. Pull request

Push the branch and open the PR:

```
git push -u origin <branch>
gh pr create --title "M<X>.<Y>: <short summary>" --body-file <body>
```

PR body structure:

```
Closes #<n>

## Plan
<the acked plan, updated to match what was actually built>

## Judgment calls
<decisions made during implementation, incl. any doc updates in this PR>

## Testing
<what just ci covers for this change; any manual verification performed>
```

Report the PR URL to the operator. Do not merge.

## 6. Review loop

When asked to address review feedback:

- Fetch it all: `gh pr view <pr> --comments` plus inline comments via `gh api repos/Hoxmot/ttanic/pulls/<pr>/comments`.
- Address every comment: either make the change or push back with reasoning — never silently skip one.
- Commit, run the gate, push to the same branch.
- Reply on the PR (`gh pr comment`) summarizing what changed for each point, with the AI-attribution note on top; leave thread resolution to the operator.
- Repeat until the operator merges.
