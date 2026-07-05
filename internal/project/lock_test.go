//go:build unix

package project

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/Hoxmot/ttanic/internal/config"
)

// lockHelperEnv, when set, tells this test binary to act as a subprocess
// that holds the lock instead of running the normal test suite. That's the
// only way to exercise "the OS releases the flock when the holding process
// dies" -- an in-process Close() call doesn't prove anything about crashes.
const lockHelperEnv = "TTANIC_LOCK_HELPER_DIR"

func TestMain(m *testing.M) {
	if dir := os.Getenv(lockHelperEnv); dir != "" {
		runLockHelper(dir)
		return
	}
	os.Exit(m.Run())
}

func runLockHelper(dir string) {
	lk, err := AcquireLock(dir)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	_ = lk
	fmt.Println("locked")
	time.Sleep(time.Hour) // parent kills us long before this elapses
}

func TestAcquireLockDouble(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, config.ProjectDirName), 0o755); err != nil {
		t.Fatal(err)
	}

	lk, err := AcquireLock(root)
	if err != nil {
		t.Fatalf("first AcquireLock() error = %v", err)
	}
	defer func() { _ = lk.Close() }()

	if _, err := AcquireLock(root); !errors.Is(err, ErrLocked) {
		t.Fatalf("second AcquireLock() error = %v, want ErrLocked", err)
	}
}

func TestLockReleasedAfterProcessExit(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, config.ProjectDirName), 0o755); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command(os.Args[0])
	cmd.Env = append(os.Environ(), lockHelperEnv+"="+root)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		t.Fatalf("starting helper process: %v", err)
	}
	defer func() {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
	}()

	scanner := bufio.NewScanner(stdout)
	if !scanner.Scan() {
		t.Fatalf("helper process produced no output: %v", scanner.Err())
	}
	if line := scanner.Text(); line != "locked" {
		t.Fatalf("helper process said %q, want %q", line, "locked")
	}

	if _, err := AcquireLock(root); !errors.Is(err, ErrLocked) {
		t.Fatalf("AcquireLock() while helper alive = %v, want ErrLocked", err)
	}

	if err := cmd.Process.Kill(); err != nil {
		t.Fatalf("killing helper process: %v", err)
	}
	_ = cmd.Wait()

	lk, err := AcquireLock(root)
	if err != nil {
		t.Fatalf("AcquireLock() after helper process exit = %v", err)
	}
	_ = lk.Close()
}
