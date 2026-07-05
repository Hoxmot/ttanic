package project

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/Hoxmot/ttanic/internal/config"
)

// lockFileName is project-specific: it has no config.toml counterpart.
const lockFileName = "lock"

// ErrNoProject is returned by Find when no ancestor of cwd contains a
// .ttanic directory.
var ErrNoProject = errors.New("not inside a ttanic project")

// IsProjectRoot reports whether dir itself is a project root, i.e. contains
// a .ttanic directory. Find uses it as the walk-up stopping condition; a
// directory walker (scan, recursive compress, tree) reuses the same check on
// each subdirectory it visits within an already-found root to detect a
// foreign nested project, per the LLD's nested-project rule: dir != root and
// IsProjectRoot(dir) means "skip this subtree and warn."
func IsProjectRoot(dir string) (bool, error) {
	info, err := os.Stat(filepath.Join(dir, config.ProjectDirName))
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return info.IsDir(), nil
}

// Find walks up from cwd looking for the nearest ancestor (inclusive) that
// is a project root. It returns ErrNoProject if it reaches the filesystem
// root without finding one, or any other error encountered while checking an
// ancestor (e.g. permission denied) wrapped with the offending path.
func Find(cwd string) (string, error) {
	dir, err := filepath.Abs(cwd)
	if err != nil {
		return "", fmt.Errorf("resolving cwd: %w", err)
	}
	for {
		ok, err := IsProjectRoot(dir)
		if err != nil {
			return "", fmt.Errorf("checking %s: %w", dir, err)
		}
		if ok {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", ErrNoProject
		}
		dir = parent
	}
}
