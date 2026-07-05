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

// Init creates dir's .ttanic directory, a config.toml seeded from answers
// (unset keys commented out, showing the built-in default), and a .ttanic/
// ignore file from answers.Ignore. Manifest creation is delegated to the
// store (M1.11). It returns ErrAlreadyInitialized if dir is already a
// project root.
func Init(dir string, answers InitAnswers) error {
	projectDir := filepath.Join(dir, config.ProjectDirName)
	if err := os.Mkdir(projectDir, 0o755); err != nil {
		if errors.Is(err, fs.ErrExist) {
			return fmt.Errorf("%w: %s", ErrAlreadyInitialized, projectDir)
		}
		return fmt.Errorf("creating %s: %w", projectDir, err)
	}

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
	b.WriteString(tomlField("level", levelSet, true, string(level), "fastest | default | better | best"))
	b.WriteString(tomlField("workers", workersSet, false, strconv.Itoa(workers), "0 = GOMAXPROCS"))
	b.WriteString("\n[archive]\n")
	b.WriteString(tomlField("on_symlink", onSymlinkSet, true, string(onSymlink), "error | skip"))
	b.WriteString("\n[ui]\n")
	b.WriteString(tomlField("theme", false, true, d.UI.Theme, ""))
	b.WriteString(tomlField("show_hidden", false, false, strconv.FormatBool(d.UI.ShowHidden), ""))
	b.WriteString(tomlField("sort", false, true, string(d.UI.Sort), "name | size | mtime"))
	b.WriteString(tomlField("editor", false, true, d.UI.Editor, `"" -> $VISUAL -> $EDITOR`))
	b.WriteString(tomlField("icons", false, true, string(d.UI.Icons), "unicode | nerd | ascii"))
	b.WriteString("\n[keys]\n")
	b.WriteString(tomlField("leader", false, true, d.Keys.Leader, ""))
	return b.String()
}

// tomlField renders one config.toml line: commented (prefixed "#") unless
// active, quoted per TOML string syntax when quoted is set.
func tomlField(key string, active, quoted bool, value, comment string) string {
	prefix := "#"
	if active {
		prefix = ""
	}
	v := value
	if quoted {
		v = strconv.Quote(value)
	}
	line := fmt.Sprintf("%s%s = %s", prefix, key, v)
	if comment != "" {
		line += "  # " + comment
	}
	return line + "\n"
}
