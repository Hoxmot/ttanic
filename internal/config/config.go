package config

import (
	"errors"
	"fmt"
)

// Level selects the zstd compression level.
type Level string

// Compression levels, mapped to zstd encoder levels by internal/archive.
const (
	LevelFastest Level = "fastest"
	LevelDefault Level = "default"
	LevelBetter  Level = "better"
	LevelBest    Level = "best"
)

// SymlinkPolicy selects how archiving treats symlinks.
type SymlinkPolicy string

// Symlink policies ("follow" is future work, per the LLD).
const (
	SymlinkError SymlinkPolicy = "error"
	SymlinkSkip  SymlinkPolicy = "skip"
)

// SortKey selects the tree-view sort order.
type SortKey string

// Sort orders for the tree view.
const (
	SortName  SortKey = "name"
	SortSize  SortKey = "size"
	SortMtime SortKey = "mtime"
)

// IconSet selects the glyphs used by the TUI.
type IconSet string

// Icon sets for the TUI.
const (
	IconsUnicode IconSet = "unicode"
	IconsNerd    IconSet = "nerd"
	IconsASCII   IconSet = "ascii"
)

// Config is the resolved ttanic configuration. It mirrors the TOML schema in
// docs/ttanic-lld.md ("internal/config"). The generated project config.toml
// deliberately carries no schema -- key discovery is "ttanic config"'s job
// (M3.13) -- so a new field or enum value here needs no other change unless
// it also becomes wizard-settable (internal/project's InitAnswers).
type Config struct {
	Compression Compression `toml:"compression"`
	Archive     Archive     `toml:"archive"`
	UI          UI          `toml:"ui"`
	Keys        Keys        `toml:"keys"`
}

// Compression holds the [compression] section.
type Compression struct {
	Level   Level `toml:"level"`
	Workers int   `toml:"workers"` // 0 = GOMAXPROCS
}

// Archive holds the [archive] section.
type Archive struct {
	OnSymlink SymlinkPolicy `toml:"on_symlink"`
}

// UI holds the [ui] section.
type UI struct {
	Theme      string  `toml:"theme"`
	ShowHidden bool    `toml:"show_hidden"`
	Sort       SortKey `toml:"sort"`
	Editor     string  `toml:"editor"` // "" -> $VISUAL -> $EDITOR
	Icons      IconSet `toml:"icons"`
}

// Keys holds the [keys] section. Custom key mappings will join this section
// when they land.
type Keys struct {
	Leader string `toml:"leader"`
}

// Validation errors. Load wraps them with the offending value; check with
// errors.Is.
var (
	ErrUnknownLevel         = errors.New("unknown compression level")
	ErrInvalidWorkers       = errors.New("workers must not be negative")
	ErrUnknownSymlinkPolicy = errors.New("unknown on_symlink policy")
	ErrUnknownSort          = errors.New("unknown sort key")
	ErrUnknownIcons         = errors.New("unknown icon set")
)

// Validate checks every constrained key and returns the first violation.
func (c Config) Validate() error {
	switch c.Compression.Level {
	case LevelFastest, LevelDefault, LevelBetter, LevelBest:
	default:
		return fmt.Errorf("%w: %q (want fastest, default, better, or best)", ErrUnknownLevel, c.Compression.Level)
	}
	if c.Compression.Workers < 0 {
		return fmt.Errorf("%w: %d", ErrInvalidWorkers, c.Compression.Workers)
	}
	switch c.Archive.OnSymlink {
	case SymlinkError, SymlinkSkip:
	default:
		return fmt.Errorf("%w: %q (want error or skip)", ErrUnknownSymlinkPolicy, c.Archive.OnSymlink)
	}
	switch c.UI.Sort {
	case SortName, SortSize, SortMtime:
	default:
		return fmt.Errorf("%w: %q (want name, size, or mtime)", ErrUnknownSort, c.UI.Sort)
	}
	switch c.UI.Icons {
	case IconsUnicode, IconsNerd, IconsASCII:
	default:
		return fmt.Errorf("%w: %q (want unicode, nerd, or ascii)", ErrUnknownIcons, c.UI.Icons)
	}
	// TODO(#25): validate UI.Theme once the Theme type exists.
	// TODO(#27): validate/parse Keys.Leader once key handling lands.
	return nil
}
