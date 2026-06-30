//go:build !windows

package platform

import (
	"context"
	"errors"
	"os"
	"os/exec"
)

func CanUsePrivilegedHelper() bool {
	if _, err := exec.LookPath("pkexec"); err != nil {
		return false
	}
	hasDisplay := os.Getenv("DISPLAY") != "" || os.Getenv("WAYLAND_DISPLAY") != ""
	hasSessionBus := os.Getenv("DBUS_SESSION_BUS_ADDRESS") != ""
	return hasDisplay && hasSessionBus
}

func PrivilegedHelperCommand(ctx context.Context, exe string, helperArgs ...string) (*exec.Cmd, error) {
	if !CanUsePrivilegedHelper() {
		return nil, errors.New("pkexec or desktop authentication session is not available")
	}
	args := []string{"env"}
	for _, key := range []string{"DISPLAY", "XAUTHORITY", "WAYLAND_DISPLAY", "XDG_RUNTIME_DIR", "DBUS_SESSION_BUS_ADDRESS"} {
		if value := os.Getenv(key); value != "" {
			args = append(args, key+"="+value)
		}
	}
	args = append(args, exe)
	args = append(args, helperArgs...)
	return exec.CommandContext(ctx, "pkexec", args...), nil
}
