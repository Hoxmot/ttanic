//go:build unix

package project

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"syscall"

	"github.com/Hoxmot/ttanic/internal/config"
)

// ErrLocked is returned by AcquireLock when another process already holds
// the project's lock.
var ErrLocked = errors.New("project is locked by another ttanic process")

// Lock is the flock held on a project's .ttanic/lock for the lifetime of a
// process. The lock lives on the open file descriptor, so the kernel
// releases it the instant the holding process dies -- crash, SIGKILL, panic,
// anything -- with no stale-lock detection needed: a leftover lock file is
// inert, and the next AcquireLock simply flocks it again.
type Lock struct {
	f *os.File
}

// AcquireLock takes the exclusive, non-blocking flock on root's
// .ttanic/lock and writes this process's PID into it for diagnostics. It
// returns ErrLocked if another process already holds the lock.
func AcquireLock(root string) (*Lock, error) {
	path := filepath.Join(root, config.ProjectDirName, lockFileName)
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return nil, fmt.Errorf("opening %s: %w", path, err)
	}
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		_ = f.Close()
		if errors.Is(err, syscall.EWOULDBLOCK) {
			return nil, ErrLocked
		}
		return nil, fmt.Errorf("locking %s: %w", path, err)
	}
	if err := writePID(f); err != nil {
		_ = syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
		_ = f.Close()
		return nil, fmt.Errorf("writing pid to %s: %w", path, err)
	}
	return &Lock{f: f}, nil
}

func writePID(f *os.File) error {
	if err := f.Truncate(0); err != nil {
		return err
	}
	_, err := f.WriteAt([]byte(strconv.Itoa(os.Getpid())), 0)
	return err
}

// Close releases the lock and closes the underlying file.
func (l *Lock) Close() error {
	defer func() { _ = l.f.Close() }()
	return syscall.Flock(int(l.f.Fd()), syscall.LOCK_UN)
}
