package project

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/Hoxmot/ttanic/internal/config"
)

func TestFindNestedCwd(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, config.ProjectDirName), 0o755); err != nil {
		t.Fatal(err)
	}
	nested := filepath.Join(root, "a", "b", "c")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}

	got, err := Find(nested)
	if err != nil {
		t.Fatalf("Find() error = %v", err)
	}
	want, err := filepath.Abs(root)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Errorf("Find() = %q, want %q", got, want)
	}
}

func TestFindAtRoot(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, config.ProjectDirName), 0o755); err != nil {
		t.Fatal(err)
	}

	got, err := Find(root)
	if err != nil {
		t.Fatalf("Find() error = %v", err)
	}
	want, err := filepath.Abs(root)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Errorf("Find() = %q, want %q", got, want)
	}
}

func TestFindNoProject(t *testing.T) {
	dir := t.TempDir()
	_, err := Find(dir)
	if !errors.Is(err, ErrNoProject) {
		t.Fatalf("Find() error = %v, want ErrNoProject", err)
	}
}

func TestIsProjectRootFile(t *testing.T) {
	dir := t.TempDir()
	// .ttanic exists but is a plain file, not a directory: not a marker.
	if err := os.WriteFile(filepath.Join(dir, config.ProjectDirName), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	ok, err := IsProjectRoot(dir)
	if err != nil {
		t.Fatalf("IsProjectRoot() error = %v", err)
	}
	if ok {
		t.Errorf("IsProjectRoot() = true, want false for a non-directory .ttanic")
	}
}
