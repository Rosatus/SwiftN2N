//go:build windows

package platform

import (
	"os/exec"
	"strconv"
	"syscall"

	"golang.org/x/sys/windows"
)

func ConfigureProcess(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP,
		HideWindow:    true,
	}
}

func TerminateProcessTree(cmd *exec.Cmd) error {
	return taskkill(cmd, false)
}

func KillProcessTree(cmd *exec.Cmd) error {
	return taskkill(cmd, true)
}

func taskkill(cmd *exec.Cmd, force bool) error {
	if cmd == nil || cmd.Process == nil {
		return nil
	}
	args := []string{"/PID", strconv.Itoa(cmd.Process.Pid), "/T"}
	if force {
		args = append(args, "/F")
	}
	return exec.Command("taskkill", args...).Run()
}

func EdgeExecutableName() string {
	return "edge.exe"
}

func SupportsForegroundFlag() bool {
	return false
}

func IsElevated() bool {
	proc := windows.NewLazySystemDLL("shell32.dll").NewProc("IsUserAnAdmin")
	result, _, _ := proc.Call()
	return result != 0
}

func IsExecutable(path string) bool {
	return path != ""
}
