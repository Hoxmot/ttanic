package project

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Hoxmot/ttanic/internal/config"
)

func TestInitIdempotent(t *testing.T) {
	dir := t.TempDir()
	if err := Init(dir, InitAnswers{}); err != nil {
		t.Fatalf("first Init() error = %v", err)
	}
	err := Init(dir, InitAnswers{})
	if !errors.Is(err, ErrAlreadyInitialized) {
		t.Fatalf("second Init() error = %v, want ErrAlreadyInitialized", err)
	}
}

func TestInitDefaultsRoundTrip(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir()) // isolate from any real global config.toml

	dir := t.TempDir()
	if err := Init(dir, InitAnswers{}); err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	cfg, err := config.Load(dir, config.Overrides{})
	if err != nil {
		t.Fatalf("config.Load() error = %v", err)
	}
	if cfg != config.Default() {
		t.Errorf("config.Load() = %+v, want config.Default() = %+v", cfg, config.Default())
	}

	raw, err := os.ReadFile(filepath.Join(dir, config.ProjectDirName, config.ConfigFileName))
	if err != nil {
		t.Fatalf("reading config.toml: %v", err)
	}
	for _, line := range strings.Split(string(raw), "\n") {
		if line != "" && !strings.HasPrefix(line, "#") {
			t.Errorf("config.toml line %q, want only the comment header when no answers are set", line)
		}
	}

	data, err := os.ReadFile(filepath.Join(dir, config.ProjectDirName, config.IgnoreFileName))
	if err != nil {
		t.Fatalf("reading ignore file: %v", err)
	}
	if len(data) != 0 {
		t.Errorf("ignore file = %q, want empty", data)
	}
}

func TestInitAppliesAnswers(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir()) // isolate from any real global config.toml

	dir := t.TempDir()
	level := config.LevelBest
	workers := 4
	onSymlink := config.SymlinkSkip
	answers := InitAnswers{
		Level:     &level,
		Workers:   &workers,
		OnSymlink: &onSymlink,
		Ignore:    []string{"*.log", "!keep.log"},
	}
	if err := Init(dir, answers); err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	cfg, err := config.Load(dir, config.Overrides{})
	if err != nil {
		t.Fatalf("config.Load() error = %v", err)
	}
	if cfg.Compression.Level != level {
		t.Errorf("Compression.Level = %q, want %q", cfg.Compression.Level, level)
	}
	if cfg.Compression.Workers != workers {
		t.Errorf("Compression.Workers = %d, want %d", cfg.Compression.Workers, workers)
	}
	if cfg.Archive.OnSymlink != onSymlink {
		t.Errorf("Archive.OnSymlink = %q, want %q", cfg.Archive.OnSymlink, onSymlink)
	}

	raw, err := os.ReadFile(filepath.Join(dir, config.ProjectDirName, config.ConfigFileName))
	if err != nil {
		t.Fatalf("reading config.toml: %v", err)
	}
	content := string(raw)
	for _, active := range []string{`level = "best"`, "workers = 4", `on_symlink = "skip"`} {
		if !strings.Contains(content, active+"\n") {
			t.Errorf("config.toml missing active line %q:\n%s", active, content)
		}
	}
	// Unset keys must not appear at all -- not even commented.
	for _, unset := range []string{"theme", "show_hidden", "sort", "editor", "icons", "leader"} {
		if strings.Contains(content, unset) {
			t.Errorf("config.toml mentions unset key %q:\n%s", unset, content)
		}
	}

	data, err := os.ReadFile(filepath.Join(dir, config.ProjectDirName, config.IgnoreFileName))
	if err != nil {
		t.Fatalf("reading ignore file: %v", err)
	}
	want := strings.Join(answers.Ignore, "\n")
	if got := strings.TrimRight(string(data), "\n"); got != want {
		t.Errorf("ignore file = %q, want %q", got, want)
	}
}

func TestInitAlreadyInitializedNoOverwrite(t *testing.T) {
	dir := t.TempDir()
	if err := Init(dir, InitAnswers{}); err != nil {
		t.Fatalf("first Init() error = %v", err)
	}
	before, err := os.ReadFile(filepath.Join(dir, config.ProjectDirName, config.ConfigFileName))
	if err != nil {
		t.Fatal(err)
	}

	level := config.LevelBest
	if err := Init(dir, InitAnswers{Level: &level}); !errors.Is(err, ErrAlreadyInitialized) {
		t.Fatalf("second Init() error = %v, want ErrAlreadyInitialized", err)
	}

	after, err := os.ReadFile(filepath.Join(dir, config.ProjectDirName, config.ConfigFileName))
	if err != nil {
		t.Fatal(err)
	}
	if string(before) != string(after) {
		t.Errorf("config.toml changed after a failed re-Init")
	}
}

func TestInitRejectsInvalidAnswers(t *testing.T) {
	dir := t.TempDir()
	badLevel := config.Level("turbo")
	err := Init(dir, InitAnswers{Level: &badLevel})
	if !errors.Is(err, config.ErrUnknownLevel) {
		t.Fatalf("Init() error = %v, want errors.Is(..., ErrUnknownLevel)", err)
	}
	if _, statErr := os.Stat(filepath.Join(dir, config.ProjectDirName)); !errors.Is(statErr, os.ErrNotExist) {
		t.Errorf(".ttanic was created despite invalid answers (stat error = %v)", statErr)
	}
}

func TestInitStrayFileNotAlreadyInitialized(t *testing.T) {
	dir := t.TempDir()
	// A plain file named .ttanic (not a directory) is not a project root, so
	// it shouldn't be mistaken for one.
	if err := os.WriteFile(filepath.Join(dir, config.ProjectDirName), []byte("oops"), 0o644); err != nil {
		t.Fatal(err)
	}

	err := Init(dir, InitAnswers{})
	if err == nil {
		t.Fatal("Init() error = nil, want a non-nil error for a stray .ttanic file")
	}
	if errors.Is(err, ErrAlreadyInitialized) {
		t.Errorf("Init() error = %v, want anything but ErrAlreadyInitialized for a non-directory .ttanic", err)
	}
}
