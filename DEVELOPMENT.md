# Development

Working on ttanic requires:

- **Go 1.26+** — the only requirement for building and testing (`CGO_ENABLED=0` is set by the justfile)
- **[just](https://github.com/casey/just)** — task runner; all recipes live in `justfile`
- **[golangci-lint](https://golangci-lint.run/) v2** — used by `just lint` and `just ci`
- **[goreleaser](https://goreleaser.com/)** — optional; only `just snapshot` needs it

On macOS: `brew install just golangci-lint goreleaser`.

## Common tasks

```
just            # list all recipes
just ci         # what CI runs: fmt-check + lint + test — must pass before every PR
just run        # build and run ttanic
```

Design docs are in `docs/` (`ttanic-hld.md`, `ttanic-lld.md`); contributor workflow and project conventions are in `AGENTS.md`.
