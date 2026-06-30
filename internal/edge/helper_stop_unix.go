//go:build !windows

package edge

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"syscall"
	"time"
)

func RunStopHelper(pid int, stdout io.Writer, stderr io.Writer) int {
	if pid <= 1 {
		_, _ = fmt.Fprintln(stderr, "refusing to stop invalid helper pid")
		return 2
	}
	if err := validateStopTarget(pid); err != nil {
		_, _ = fmt.Fprintln(stderr, err)
		return 1
	}

	if err := syscall.Kill(pid, syscall.SIGTERM); err != nil && !errors.Is(err, syscall.ESRCH) {
		_, _ = fmt.Fprintf(stderr, "failed to signal helper: %v\n", err)
		return 1
	}
	if waitForProcessExit(pid, 6*time.Second) {
		_, _ = fmt.Fprintf(stdout, "helper %d stopped\n", pid)
		return 0
	}

	if err := syscall.Kill(pid, syscall.SIGKILL); err != nil && !errors.Is(err, syscall.ESRCH) {
		_, _ = fmt.Fprintf(stderr, "failed to kill helper: %v\n", err)
		return 1
	}
	if waitForProcessExit(pid, 2*time.Second) {
		_, _ = fmt.Fprintf(stdout, "helper %d killed\n", pid)
		return 0
	}

	_, _ = fmt.Fprintf(stderr, "helper %d did not exit\n", pid)
	return 1
}

func validateStopTarget(pid int) error {
	cmdline, err := os.ReadFile(fmt.Sprintf("/proc/%d/cmdline", pid))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("cannot inspect helper process: %w", err)
	}
	parts := strings.Split(strings.TrimRight(string(cmdline), "\x00"), "\x00")
	if len(parts) < 3 {
		return errors.New("target process is not a SwiftN2N edge helper")
	}
	if parts[1] != HelperFlag || parts[2] != HelperEdge {
		return errors.New("target process is not a SwiftN2N edge helper")
	}
	return nil
}

func waitForProcessExit(pid int, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if !processExists(pid) {
			return true
		}
		time.Sleep(100 * time.Millisecond)
	}
	return !processExists(pid)
}

func processExists(pid int) bool {
	err := syscall.Kill(pid, 0)
	return err == nil || errors.Is(err, syscall.EPERM)
}
