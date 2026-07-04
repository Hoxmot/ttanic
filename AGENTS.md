# ttanic

`ttanic` (tar-tanic) is a CLI/TUI tool for managing archive folders: compressing, decompressing, and browsing directories as `.tar.zst` archives, with a per-project manifest that enables browsing archive contents without decompressing, integrity verification, and fuzzy search inside archives.

## Project status

M1 in progress — repository scaffolded (M1.1), no functionality yet (update this line as milestones land). The sources of truth for design decisions are `docs/ttanic-hld.md` (what and why) and `docs/ttanic-lld.md` (how: packages, schemas, interfaces). Read them before proposing or implementing anything, and keep them updated when decisions change.

## Workflow

Implementation is tracked in GitHub issues #1–#52 (`gh issue list`), grouped into milestones M1–M4 mirroring the LLD. Issue number = task ID: task M1.13 is issue #13. Each issue body has Context / Needs / Scope / Done when.

The `work-issue` skill (`.claude/skills/work-issue/`) encodes the full flow — issue → branch → plan → operator ack → implement → PR → review loop. Use it when working on an issue; the rules below are its foundation.

- **Pick up an issue**: read its body plus the LLD sections it references. **Needs** lists blocking task IDs — do not start an issue whose dependencies aren't merged.
- **Branch + PR per issue**: branch from `main` named like `m1-1-scaffold`; the PR body must contain `Closes #<n>` and note any judgment calls. Never commit or push to `main` directly; the user reviews and squash-merges (one commit on `main` per issue). When opening or updating a PR, suggest the squash-merge commit message.
- **Gate**: `just ci` must pass locally before the PR and in CI.
- **Scope**: implement what the issue's Done-when requires — no drive-by refactors; spotted problems become new issues (`gh issue create`).
- **Deviations**: if implementation forces a change to the HLD/LLD, update the doc in the same PR and call it out in the PR description. Docs and code must never silently disagree.

## Tech stack

- Go, CGo-free (required for `go install` and cross-platform goreleaser builds)
- TUI: charm.land stack — Bubble Tea, Huh, Lip Gloss
- Archiving: `archive/tar` (stdlib) + `klauspost/compress/zstd` — in-process, never shell out to `tar`/`zstd` binaries
- Manifest storage: SQLite via `modernc.org/sqlite` (pure Go), behind a storage interface so other backends (e.g. JSONL) can be added later

## Key design constraints

These are settled decisions (see the HLD for rationale) — do not silently deviate:

- Single archive format: `.tar.zst`. Both files and directories go through tar.
- A project is marked by a `.ttanic/` directory (config + manifest). Outside a project, manifest-dependent features error or degrade with a warning; they never write state elsewhere.
- Safety first: originals are deleted only after the written archive is verified (verify-then-delete). Every delete requires confirmation.
- Operations are modeled as messages so the MVP's modal execution can become a background job queue later without a rewrite.
- CLI (verb subcommands) and TUI share the same core engine; every TUI feature must be reachable from the CLI.

## Conventions

- Commit messages: conventional-commit style prefixes as seen in history (`docs:`, `feat:`, `fix:`, ...)
- Documentation lives in `docs/`
- Task runner: `just` (not make) — build/test/lint recipes live in `justfile` once code lands; `just ci` mirrors CI
