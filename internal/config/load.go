package config

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

// Overrides is the final merge layer: per-run values from CLI flags. A nil
// field means "flag not passed". Only keys with a persistent root flag are
// represented; add a field (and its applyOverrides line) when a new flag
// lands.
type Overrides struct {
	Level      *Level
	Workers    *int
	OnSymlink  *SymlinkPolicy
	Sort       *SortKey
	ShowHidden *bool
}

// Well-known file and directory names.
const (
	projectDirName = ".ttanic"     // the project marker directory
	configFileName = "config.toml" // per-scope config, global and project
	ignoreFileName = "ignore"      // gitignore-syntax ignore patterns
)

// GlobalDir returns the global ttanic config directory:
// $XDG_CONFIG_HOME/ttanic when the variable is set, otherwise
// ~/.config/ttanic — on macOS too (deliberately not os.UserConfigDir).
func GlobalDir() (string, error) {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "ttanic"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolving config dir: %w", err)
	}
	return filepath.Join(home, ".config", "ttanic"), nil
}

// Load resolves the effective configuration: built-in defaults, then the
// global config.toml, then <projectDir>/.ttanic/config.toml, then flag
// overrides. Each file layer overrides only the keys it actually defines.
// projectDir == "" means standalone mode (no project layer). Missing files
// are fine; malformed or invalid ones are errors naming the file.
func Load(projectDir string, ov Overrides) (Config, error) {
	cfg := Default()

	globalDir, err := GlobalDir()
	if err != nil {
		return Config{}, err
	}
	if err := applyFile(&cfg, filepath.Join(globalDir, configFileName)); err != nil {
		return Config{}, err
	}
	if projectDir != "" {
		if err := applyFile(&cfg, filepath.Join(projectDir, projectDirName, configFileName)); err != nil {
			return Config{}, err
		}
	}
	applyOverrides(&cfg, ov)

	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

// LoadIgnore resolves the layered ignore rules: the global ignore file
// (GlobalDir()/ignore), then <projectDir>/.ttanic/ignore. Project rules are
// appended after global ones, so a project "!" pattern can re-include what
// the global file ignores. projectDir == "" means standalone mode (no project
// layer). Missing files are fine; with no ignore file at all the returned
// matcher matches nothing.
func LoadIgnore(projectDir string) (*Matcher, error) {
	globalDir, err := GlobalDir()
	if err != nil {
		return nil, err
	}
	paths := []string{filepath.Join(globalDir, ignoreFileName)}
	if projectDir != "" {
		paths = append(paths, filepath.Join(projectDir, projectDirName, ignoreFileName))
	}
	var layers [][]Rule
	for _, p := range paths {
		data, err := os.ReadFile(p)
		if errors.Is(err, fs.ErrNotExist) {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("%s: %w", p, err)
		}
		layers = append(layers, ParseIgnore(string(data)))
	}
	return NewMatcher(layers...), nil
}

// applyFile merges the keys defined in the TOML file at path into cfg. A
// missing file is a no-op.
func applyFile(cfg *Config, path string) error {
	var layer Config
	md, err := toml.DecodeFile(path, &layer)
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}
	if undecoded := md.Undecoded(); len(undecoded) > 0 {
		keys := make([]string, len(undecoded))
		for i, k := range undecoded {
			keys[i] = k.String()
		}
		return fmt.Errorf("%s: unknown key(s): %s", path, strings.Join(keys, ", "))
	}
	merge(cfg, layer, md)
	return nil
}

// merge copies from src into dst every key the decoded file defined,
// per toml.MetaData.IsDefined — so a file setting only compression.level
// leaves every other key untouched.
func merge(dst *Config, src Config, md toml.MetaData) {
	if md.IsDefined("compression", "level") {
		dst.Compression.Level = src.Compression.Level
	}
	if md.IsDefined("compression", "workers") {
		dst.Compression.Workers = src.Compression.Workers
	}
	if md.IsDefined("archive", "on_symlink") {
		dst.Archive.OnSymlink = src.Archive.OnSymlink
	}
	if md.IsDefined("ui", "theme") {
		dst.UI.Theme = src.UI.Theme
	}
	if md.IsDefined("ui", "show_hidden") {
		dst.UI.ShowHidden = src.UI.ShowHidden
	}
	if md.IsDefined("ui", "sort") {
		dst.UI.Sort = src.UI.Sort
	}
	if md.IsDefined("ui", "editor") {
		dst.UI.Editor = src.UI.Editor
	}
	if md.IsDefined("ui", "icons") {
		dst.UI.Icons = src.UI.Icons
	}
	if md.IsDefined("keys", "leader") {
		dst.Keys.Leader = src.Keys.Leader
	}
}

func applyOverrides(cfg *Config, ov Overrides) {
	if ov.Level != nil {
		cfg.Compression.Level = *ov.Level
	}
	if ov.Workers != nil {
		cfg.Compression.Workers = *ov.Workers
	}
	if ov.OnSymlink != nil {
		cfg.Archive.OnSymlink = *ov.OnSymlink
	}
	if ov.Sort != nil {
		cfg.UI.Sort = *ov.Sort
	}
	if ov.ShowHidden != nil {
		cfg.UI.ShowHidden = *ov.ShowHidden
	}
}
