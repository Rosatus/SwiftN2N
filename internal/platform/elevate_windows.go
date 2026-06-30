//go:build windows

package platform

import (
	"context"
	"errors"
	"os/exec"
)

func CanUsePrivilegedHelper() bool {
	return false
}

func PrivilegedHelperCommand(ctx context.Context, exe string, helperArgs ...string) (*exec.Cmd, error) {
	return nil, errors.New("Windows uses the application manifest for administrator privileges")
}
