# ttanic HLD

## Description

`ttanic` (pronounced tee-tanic) is a short for tar-tanic. An app for managing your archive folders. It allows you to manage and automatically compress your directories.

## Need

Recently, I started working on organising and backing up my data. As a result, I realised that some of the data can be compressed to save some space. However, considering how long it takes to decompress. I'd like to manage it the compression of my files and directories more easily.

The problem is that I also don't remember all the commands to compress and decompress -- especially the more complicated ones. As a result, it's not easy to manage it manually. That's why I need something to help me with that.

## Features

I'd like the program to have to options to work with:

- CLI, where I directly give commands to do something
- TUI, which will be a bit easier and interactive way to use the app

I'd like to have a nice TUI, where the user can navigate tree structure. The user can navigate, select files, and directory, group compress, group decompress, etc.

The program also has a flow, where the user can select to compress all the files in the selected directory separately. The program also keeps the manifest of all the files in the compressed directories. As a result, the user can easily browse the contents of compressed archives.

The user can use the program in initialised directory as a tar-tanic project to keep the manifest and configs locally or it can be used as a simple compression tool standalone.

The program should have a `?` command to display the help menu.

The program should have `/` command for looking for particular files in the directory.

The user can use `<SUPER>/` to do fuzzy search from the files in the tar-tanic project or from the file from the directory the program was open down (if the program was opened outside of tar-tanic project).

I'd like to also have a config view for the user to modify the config. If the user is in the tar-tanic project, the program should edit directory-level settings. If the user outside of tar-tanic project, the user edits global settings.

All the TUI features are also available from CLI.

### User flows

#### Start

1. I open TUI in an uninitialised directory
2. I press `S` or type `:init` to start initialisation process
3. The programs launches a wizard to ask me questions about the project
4. After I answer to questions, the program creates manifest and config to initialise tar-tanic project.

#### Simple compression

1. I open the TUI
2. I navigate to the folder I want to compress with `hjkl` (vim-style navigation) or arrows
3. I press `c` for compression
4. The program compresses the directory/file

#### Simple decompression

1. I open the TUI
2. I navigate to the folder I want to compress with `hjkl` (vim-style navigation) or arrows
3. I press `i` for decompression (inflate)
4. The program decompresses the directory/file

#### Deleting the directory

1. I open the TUI
2. I navigate to the folder I want to compress with `hjkl` (vim-style navigation) or arrows
3. I press `d` for deletion
	1. If there is a compressed directory/file the user wants to delete, the file/directory gets deleted
	2. If the program doesn't see the compressed directory/file in the same directory, it prompts for confirmation or compress and delete.
		1. If the I choose delete, the program deletes the directory/file
		2. If the I choose abort, the program aborts
		3. If the I choose compress and delete, the program compresses, then deletes the file/directory

#### Compres and delete

1. I open the TUI
2. I navigate to the folder I want to compress with `hjkl` (vim-style navigation) or arrows
3. I press `:` to start a command typing, and type `cd` for compression and deletion
4. The program compresses the directory/file and then deletes the original directory/file

#### Compress recursively

1. I open the TUI
2. I navigate to the folder I want to compress with `hjkl` (vim-style navigation) or arrows
3. I press `C` for recursive compression
4. The program compresses the all the files inside the selected directory into separate archives. If the selected thing is a file, it behaves like `c` (single compression)

#### Decompress recursively

1. I open the TUI
2. I navigate to the folder I want to compress with `hjkl` (vim-style navigation) or arrows
3. I press `I` for recursive decompression (inflate)
4. The program decompresses all the archives inside the selected directory. If the selected thing is a file, it behaves like `i` (single decompression)

#### Multi selection

1. I open the TUI
2. I navigate to the folder I want to compress with `hjkl` (vim-style navigation) or arrows
3. I press `v` for selecting the directory/file
4. I navigate to a different directory/file
5. I press `v` again to select the other directory/file
6. I press `V` for multi selection and can go up/down to select all the file I go thogh
7. I press `V` to stop multi-selection
8. I can type whatever command and it applies to all the files/directories I selected
9. I can press `ESC` to deselect all the files

## Tech-stack

I'd like to use `go` programming language with nice TUI made with tools from charm.land: `Bubble Tea`, `Huh`, and `Lip Gloss`.

For archiving and compression: `zstd` compresses a single stream, so directories are first archived with tar, producing `.tar.zst`. Both steps happen in-process with Go libraries (`archive/tar` from the stdlib + `klauspost/compress/zstd`) -- no external binaries required on the host.

Of course, I'd like the program to be easily distributable/installable.

## Design decisions

### Archive format

- ttanic produces exactly one format: `.tar.zst`. Single files and directories both go through tar, so every archive looks the same to the manifest and the decompressor.
- Reading/creating foreign formats (`.zip`, `.tar.gz`, `.tar.xz`, ...) is out of scope for v1.

### Manifest

The manifest is the differentiating core of ttanic. It guarantees:

- **Browsing**: the file listing (paths, sizes, mtimes) of every archive ttanic creates, so the TUI can expand an archive like a read-only folder without decompressing it.
- **Integrity**: checksums of archives and their contents, used by verify-then-delete and by `verify`/`scan` to detect corruption.
- **Search**: fuzzy search (`<SUPER>/`) finds files *inside* archives, not just loose files.
- **Inventory**: a record of everything archived in the project, enabling "where did I put X" queries.

Storage sits behind a small interface. MVP backend: SQLite via a pure-Go driver (`modernc.org/sqlite`, keeps the build CGo-free). A JSONL backend (slower, but human-readable and diff-able) becomes a config choice later.

**Drift handling**: archives can be moved, deleted, or modified outside ttanic. On TUI startup a quick scan compares existence/size/mtime against the manifest and visibly marks stale entries, offering a fix. `:scan` (TUI) / `ttanic scan` (CLI) does a deep re-verify with checksums. Nothing in the manifest is changed silently.

### Project model

- A project is marked by a `.ttanic/` directory (like `.git/`) holding `config.toml`, the manifest database, and any future state. The project root is the nearest ancestor containing `.ttanic/`.
- **Standalone mode** (outside a project): all manifest features are disabled. If a manifest is on the critical path of a command (e.g. search inside archives, browse archive contents), the user gets an error that the project isn't initialised. If the manifest part is incidental (e.g. plain compression, which would normally also record to the manifest), the operation runs and the manifest step is skipped with a warning.

### Safety

- **Verify-then-delete**: an original is only deleted after the freshly written archive is re-read and its contents verified against checksums. This applies to compress-and-delete (`:cd`) and to any flow that removes originals.
- **All deletes confirm**: `d` shows a y/n prompt (default: no) naming the target. The prompt also accepts `d` as confirmation, so `dd` deletes -- matching vim muscle memory while still preventing fat-finger accidents.

### Operation semantics

- **Placement**: compressing `photos/` creates `photos.tar.zst` as a sibling. On collision the user is prompted: overwrite / rename with suffix / abort. The original is kept unless the operation was compress-and-delete.
- **Recursive compress (`C`)**: covers immediate children only -- each direct child (file or subdirectory) of the selected directory becomes its own archive; subdirectories are archived whole. Already-compressed children are skipped, and ignore patterns are respected.
- **Jobs**: MVP runs one operation at a time, modal with a progress bar and cancellation. Operations are modeled as messages from the start, so a background job queue can be added later without a rewrite.

### File operations (TUI)

The TUI is a file manager, so it includes basic file operations beyond compress/decompress:

- `d` -- delete (already covered above)
- `y` / `x` / `p` -- copy / cut / paste (vim-style yank register; works with multi-selection)
- `r` -- rename
- `e` -- open an uncompressed file in the editor (config `editor`, falling back to `$VISUAL`/`$EDITOR`; the TUI suspends while the editor runs)

When one of these operations touches a manifest-tracked archive, the manifest is updated as part of the operation (rename/move updates the path, paste of a copied archive adds an entry, delete removes it). This is the main reason to do file operations *inside* ttanic instead of in a shell: no drift.

These are TUI-only. The CLI doesn't mirror them (`cp`/`mv`/`rm` already exist in the shell); whether to add e.g. `ttanic mv` later purely for manifest-synced moves stays an open question -- until then, scan-on-open catches drift from shell-side moves.

### Config

- Global config at `~/.config/ttanic/config.toml`, overridden per-project by `.ttanic/config.toml`.
- Format is TOML. Rationale: config is hand-edited, and TOML supports comments (documenting settings, keeping commented-out defaults), is typed and unambiguous, and is the de-facto Go/CLI-tool convention. JSON has no comments and is noisy to hand-edit; YAML is error-prone (indentation, implicit typing). JSON remains the right tool where machines exchange data (future `--json` CLI output).
- Contents: compression settings (zstd level, worker count), gitignore-style ignore patterns (what recursive compress and scan skip: `.git`, `node_modules`, ...), UI preferences (theme name, hidden files, sorting, prompts), and `editor` (empty = fall back to `$VISUAL`, then `$EDITOR`).
- Remappable keybindings are deferred to v2.

### Themes

- All TUI styling flows through a single `Theme` struct (colors + Lip Gloss styles) resolved once at startup -- no hardcoded styles scattered through views. This makes custom themes cheap to add.
- MVP ships a built-in default theme. Custom themes come from files: `~/.config/ttanic/themes/<name>.toml`, selected by `theme = "<name>"` in config (post-MVP).

### CLI

- Verb subcommands with full TUI feature parity: `ttanic init`, `compress`, `decompress`, `ls <archive>`, `search <query>`, `scan`, `verify`, `delete`. Plain `ttanic` opens the TUI.

### In-archive browsing

- MVP: expanding an archive in the TUI shows a read-only listing (names, sizes, mtimes) served from the manifest. Getting a file out means decompressing the whole archive.
- Single-file extraction from an archive is planned for MVP+1.

### Distribution

- MVP: tagged releases build cross-platform static binaries with goreleaser, published to GitHub Releases; `go install` works out of the box (public repo, CGo-free build).
- Post-MVP: Homebrew tap (maintained by goreleaser); Nix/AUR/other package managers later.

### Future considerations (explicitly not designed yet)

- **Plugins**: extensibility is attractive, but the right API surface isn't clear yet (exec-style hooks like git? embedded scripting?). No commitment now; the decisions already made keep the door open -- the core engine is a library separate from the UIs, operations are messages, and the CLI can grow `--json` output for external tooling.

## Open questions for low-level design

- Checksum algorithm and granularity: fast xxhash vs sha256; per-file, per-archive, or both.
- Symlink handling (likely tar's default: store the link itself, don't follow).
- Behavior for nested `.ttanic/` projects (likely: nearest ancestor wins, warn on nesting).
- Manifest schema versioning and migrations.
- Partial-failure semantics for recursive compress (continue and report vs abort).
- Windows support level (paths, config location, terminal behavior).

