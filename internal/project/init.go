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
// flags. A nil pointer field means "not set": the key is omitted from the
// generated config.toml entirely, so the effective value keeps coming from
// the global config or built-in defaults under config.Load's merge rules.
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

// Init creates dir's .ttanic directory, a config.toml containing only the
// keys set in answers (see renderConfigToml), and a .ttanic/ignore file from
// answers.Ignore. Manifest creation is delegated to the store (M1.11). It returns ErrAlreadyInitialized if dir is already a
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
# this project only. Only keys explicitly set appear here; everything else
# comes from the global config or ttanic's built-in defaults, and copying
# those values here would silently pin them. Use "ttanic config" to list
# available keys, allowed values, and the values currently in effect.
`

// renderConfigToml builds config.toml's initial content: the explanatory
// header plus one active line per answer the wizard/flags actually set.
// Unset keys are omitted entirely -- no commented schema, no default values.
// In a layered config anything baked into a generated file either goes
// stale (a default or the global config changes) or, once uncommented,
// silently pins a value the user never chose. Key discovery is "ttanic
// config"'s job (M3.13), not this file's.
func renderConfigToml(answers InitAnswers) string {
	var b strings.Builder
	b.WriteString(configTomlHeader)
	if answers.Level != nil || answers.Workers != nil {
		b.WriteString("\n[compression]\n")
		if answers.Level != nil {
			fmt.Fprintf(&b, "level = %s\n", strconv.Quote(string(*answers.Level)))
		}
		if answers.Workers != nil {
			fmt.Fprintf(&b, "workers = %d\n", *answers.Workers)
		}
	}
	if answers.OnSymlink != nil {
		b.WriteString("\n[archive]\n")
		fmt.Fprintf(&b, "on_symlink = %s\n", strconv.Quote(string(*answers.OnSymlink)))
	}
	return b.String()
}
