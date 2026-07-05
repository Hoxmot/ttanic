package project

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Hoxmot/ttanic/internal/config"
)

// ErrAlreadyInitialized is returned by Init when dir already has a .ttanic
// directory.
var ErrAlreadyInitialized = errors.New("project already initialized")

// InitAnswers carries the choices made by the init wizard or equivalent CLI
// flags. A nil pointer field means "not set": Init leaves the corresponding
// config.toml key commented, showing config.Default()'s value, exactly as an
// unset key already behaves under config.Load's merge rules.
type InitAnswers struct {
	Level     *config.Level
	Workers   *int
	OnSymlink *config.SymlinkPolicy
	Ignore    []string // initial .ttanic/ignore lines; nil/empty -> empty file
}

// Validate reports whether the set fields describe values config.Config
// will accept, so a bad answer (e.g. a mistyped level) fails at Init time
// rather than silently producing a config.toml that the next config.Load
// rejects.
func (a InitAnswers) Validate() error {
	cfg := config.Default()
	if a.Level != nil {
		cfg.Compression.Level = *a.Level
	}
	if a.Workers != nil {
		cfg.Compression.Workers = *a.Workers
	}
	if a.OnSymlink != nil {
		cfg.Archive.OnSymlink = *a.OnSymlink
	}
	return cfg.Validate()
}

// Init creates dir's .ttanic directory, a config.toml seeded from answers
// (unset keys commented out, showing the built-in default), and a .ttanic/
// ignore file from answers.Ignore. Manifest creation is delegated to the
// store (M1.11). It returns ErrAlreadyInitialized if dir is already a
// project root; if dir/.ttanic exists but isn't a directory, or an answer
// fails Validate, it returns a distinct, descriptive error instead. A
// failure after the directory is created is rolled back (the directory is
// removed) so a retry isn't permanently blocked by a half-written project.
func Init(dir string, answers InitAnswers) error {
	if err := answers.Validate(); err != nil {
		return fmt.Errorf("invalid init answers: %w", err)
	}

	projectDir := filepath.Join(dir, config.ProjectDirName)
	if err := os.Mkdir(projectDir, 0o755); err != nil {
		if errors.Is(err, fs.ErrExist) {
			ok, statErr := IsProjectRoot(dir)
			if statErr != nil {
				return fmt.Errorf("checking %s: %w", projectDir, statErr)
			}
			if ok {
				return fmt.Errorf("%w: %s", ErrAlreadyInitialized, projectDir)
			}
			return fmt.Errorf("creating %s: exists and is not a directory", projectDir)
		}
		return fmt.Errorf("creating %s: %w", projectDir, err)
	}

	if err := writeProjectFiles(projectDir, answers); err != nil {
		_ = os.RemoveAll(projectDir)
		return err
	}
	return nil
}

func writeProjectFiles(projectDir string, answers InitAnswers) error {
	configPath := filepath.Join(projectDir, config.ConfigFileName)
	if err := os.WriteFile(configPath, []byte(renderConfigToml(answers)), 0o644); err != nil {
		return fmt.Errorf("writing %s: %w", configPath, err)
	}

	ignorePath := filepath.Join(projectDir, config.IgnoreFileName)
	var ignoreContent string
	if len(answers.Ignore) > 0 {
		ignoreContent = strings.Join(answers.Ignore, "\n") + "\n"
	}
	if err := os.WriteFile(ignorePath, []byte(ignoreContent), 0o644); err != nil {
		return fmt.Errorf("writing %s: %w", ignorePath, err)
	}
	return nil
}

const configTomlHeader = `# ttanic project configuration.
#
# Settings here override the global config
# ($XDG_CONFIG_HOME/ttanic/config.toml, or ~/.config/ttanic/config.toml) for
# this project only. A commented key shows the value currently in effect;
# uncomment (or edit) a key to change it for this project.
`

// renderConfigToml builds config.toml's initial content: every key defaults
// to commented (showing config.Default()'s value), except the three the init
// wizard/flags can set (level, workers, on_symlink), which are written
// active when the matching answers field is set.
func renderConfigToml(answers InitAnswers) string {
	d := config.Default()

	level, levelSet := d.Compression.Level, answers.Level != nil
	if levelSet {
		level = *answers.Level
	}
	workers, workersSet := d.Compression.Workers, answers.Workers != nil
	if workersSet {
		workers = *answers.Workers
	}
	onSymlink, onSymlinkSet := d.Archive.OnSymlink, answers.OnSymlink != nil
	if onSymlinkSet {
		onSymlink = *answers.OnSymlink
	}

	var b strings.Builder
	b.WriteString(configTomlHeader)
	b.WriteString("\n[compression]\n")
	b.WriteString(tomlField(tomlFieldOpts{Key: "level", Active: levelSet, Quoted: true, Value: string(level), Comment: "fastest | default | better | best"}))
	b.WriteString(tomlField(tomlFieldOpts{Key: "workers", Active: workersSet, Value: strconv.Itoa(workers), Comment: "0 = GOMAXPROCS"}))
	b.WriteString("\n[archive]\n")
	b.WriteString(tomlField(tomlFieldOpts{Key: "on_symlink", Active: onSymlinkSet, Quoted: true, Value: string(onSymlink), Comment: "error | skip"}))
	b.WriteString("\n[ui]\n")
	b.WriteString(tomlField(tomlFieldOpts{Key: "theme", Quoted: true, Value: d.UI.Theme}))
	b.WriteString(tomlField(tomlFieldOpts{Key: "show_hidden", Value: strconv.FormatBool(d.UI.ShowHidden)}))
	b.WriteString(tomlField(tomlFieldOpts{Key: "sort", Quoted: true, Value: string(d.UI.Sort), Comment: "name | size | mtime"}))
	b.WriteString(tomlField(tomlFieldOpts{Key: "editor", Quoted: true, Value: d.UI.Editor, Comment: `"" -> $VISUAL -> $EDITOR`}))
	b.WriteString(tomlField(tomlFieldOpts{Key: "icons", Quoted: true, Value: string(d.UI.Icons), Comment: "unicode | nerd | ascii"}))
	b.WriteString("\n[keys]\n")
	b.WriteString(tomlField(tomlFieldOpts{Key: "leader", Quoted: true, Value: d.Keys.Leader}))
	return b.String()
}

// tomlFieldOpts describes one config.toml line for tomlField.
type tomlFieldOpts struct {
	Key     string
	Active  bool // write uncommented (an answer overrides the default); default false: commented
	Quoted  bool // wrap Value in TOML string quotes
	Value   string
	Comment string // appended as "  # Comment" when non-empty
}

// tomlField renders one config.toml line: commented (prefixed "#") unless
// Active is set.
func tomlField(f tomlFieldOpts) string {
	prefix := "#"
	if f.Active {
		prefix = ""
	}
	v := f.Value
	if f.Quoted {
		v = strconv.Quote(f.Value)
	}
	line := fmt.Sprintf("%s%s = %s", prefix, f.Key, v)
	if f.Comment != "" {
		line += "  # " + f.Comment
	}
	return line + "\n"
}
