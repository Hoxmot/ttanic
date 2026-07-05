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

	data, err := os.ReadFile(filepath.Join(dir, config.ProjectDirName, config.IgnoreFileName))
	if err != nil {
		t.Fatalf("reading ignore file: %v", err)
	}
	if len(data) != 0 {
		t.Errorf("ignore file = %q, want empty", data)
	}
}

func TestInitAppliesAnswers(t *testing.T) {
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
