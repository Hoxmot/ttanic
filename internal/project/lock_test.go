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

// helperMaxLifetime bounds how long a stray helper process could survive if
// the parent test somehow failed to kill it -- the parent always kills it
// within milliseconds under normal operation, so this only caps the
// worst-case orphan lifetime.
const helperMaxLifetime = 30 * time.Second

func runLockHelper(dir string) {
	lk, err := AcquireLock(dir)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	// The defer keeps lk reachable for the whole sleep. Without it the GC
	// could collect the Lock (os.File's cleanup closes the fd, releasing the
	// flock) while the parent still expects the lock to be held.
	defer func() { _ = lk.Close() }()
	fmt.Println("locked")
	time.Sleep(helperMaxLifetime)
}

// waitTimeout waits for cmd to exit, reporting a timeout error instead of
// blocking forever if it doesn't within d. The exit error itself (expected
// and non-nil when the process was killed) is discarded -- only whether it
// exited at all is interesting here.
func waitTimeout(cmd *exec.Cmd, d time.Duration) error {
	done := make(chan struct{})
	go func() {
		_ = cmd.Wait()
		close(done)
	}()
	select {
	case <-done:
		return nil
	case <-time.After(d):
		return fmt.Errorf("process %d did not exit within %s", cmd.Process.Pid, d)
	}
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

	if _, err := AcquireLock(root); !errors.Is(err, ErrLocked) {
		t.Fatalf("second AcquireLock() error = %v, want ErrLocked", err)
	}

	// An explicit Close releases the lock: re-acquiring must now succeed.
	if err := lk.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	lk2, err := AcquireLock(root)
	if err != nil {
		t.Fatalf("AcquireLock() after Close() error = %v", err)
	}
	defer func() { _ = lk2.Close() }()
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
		_ = waitTimeout(cmd, 5*time.Second)
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
	if err := waitTimeout(cmd, 5*time.Second); err != nil {
		t.Fatalf("helper process did not exit after SIGKILL: %v", err)
	}

	lk, err := AcquireLock(root)
	if err != nil {
		t.Fatalf("AcquireLock() after helper process exit = %v", err)
	}
	_ = lk.Close()
}
