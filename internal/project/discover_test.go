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

func TestFindNonexistentCwd(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, config.ProjectDirName), 0o755); err != nil {
		t.Fatal(err)
	}

	// A nonexistent cwd must error, not resolve to the nearest existing
	// ancestor's project.
	_, err := Find(filepath.Join(root, "does", "not", "exist"))
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("Find() error = %v, want errors.Is(..., os.ErrNotExist)", err)
	}
}

func TestFindFileCwd(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, config.ProjectDirName), 0o755); err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(root, "plain.txt")
	if err := os.WriteFile(file, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := Find(file); err == nil {
		t.Fatal("Find() error = nil, want an error for a non-directory cwd")
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
