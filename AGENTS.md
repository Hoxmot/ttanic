# ttanic

`ttanic` (tar-tanic) is a CLI/TUI tool for managing archive folders: compressing, decompressing, and browsing directories as `.tar.zst` archives, with a per-project manifest that enables browsing archive contents without decompressing, integrity verification, and fuzzy search inside archives.

## Project status

Design phase — no code yet. The source of truth for all design decisions is `docs/ttanic-hld.md`. Read it before proposing or implementing anything, and keep it updated when decisions change.

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
