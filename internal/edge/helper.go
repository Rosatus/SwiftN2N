package edge

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"SwiftN2N/internal/platform"
)

const (
	HelperFlag = "--helper"
	HelperEdge = "edge"
	HelperStop = "stop"
)

type HelperEvent struct {
	Type    string `json:"type"`
	Time    string `json:"time"`
	Stream  string `json:"stream,omitempty"`
	Line    string `json:"line,omitempty"`
	Code    int    `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

func RunHelper(args []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) int {
	if len(args) < 2 || args[0] != HelperFlag {
		_, _ = fmt.Fprintln(stderr, "unknown helper command")
		return 2
	}
	switch args[1] {
	case HelperEdge:
		return RunEdgeHelper(stdin, stdout)
	case HelperStop:
		if len(args) < 3 {
			_, _ = fmt.Fprintln(stderr, "missing helper pid")
			return 2
		}
		pid, err := strconv.Atoi(args[2])
		if err != nil {
			_, _ = fmt.Fprintln(stderr, "invalid helper pid")
			return 2
		}
		return RunStopHelper(pid, stdout, stderr)
	default:
		_, _ = fmt.Fprintln(stderr, "unknown helper command")
		return 2
	}
}

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

func RunEdgeHelper(stdin io.Reader, stdout io.Writer) int {
	encoder := json.NewEncoder(stdout)
	var emitMu sync.Mutex
	emit := func(event HelperEvent) {
		event.Time = time.Now().Format(time.RFC3339)
		emitMu.Lock()
		defer emitMu.Unlock()
		_ = encoder.Encode(event)
	}

	var config Config
	if err := json.NewDecoder(stdin).Decode(&config); err != nil {
		emit(HelperEvent{Type: "error", Message: "invalid helper config: " + err.Error()})
		return 2
	}

	edgePath, err := ResolvePath(config.EdgePath)
	if err != nil {
		emit(HelperEvent{Type: "error", Message: err.Error()})
		return 1
	}
	spec, err := BuildCommand(config)
	if err != nil {
		emit(HelperEvent{Type: "error", Message: err.Error()})
		return 1
	}
	if !platform.IsExecutable(edgePath) {
		emit(HelperEvent{Type: "error", Message: "edge binary is not executable"})
		return 1
	}

	cmd := exec.Command(edgePath, spec.Args...)
	cmd.Env = append(os.Environ(), spec.Env...)
	platform.ConfigureProcess(cmd)

	edgeStdout, err := cmd.StdoutPipe()
	if err != nil {
		emit(HelperEvent{Type: "error", Message: err.Error()})
		return 1
	}
	edgeStderr, err := cmd.StderrPipe()
	if err != nil {
		emit(HelperEvent{Type: "error", Message: err.Error()})
		return 1
	}

	if err := cmd.Start(); err != nil {
		emit(HelperEvent{Type: "error", Message: err.Error()})
		return 1
	}
	emit(HelperEvent{Type: "status", Message: fmt.Sprintf("helper started edge pid %d", cmd.Process.Pid)})

	var wg sync.WaitGroup
	wg.Add(2)
	go helperScanPipe(&wg, &emitMu, encoder, "stdout", edgeStdout)
	go helperScanPipe(&wg, &emitMu, encoder, "stderr", edgeStderr)

	signalCh := make(chan os.Signal, 2)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM)
	done := make(chan struct{})
	go func() {
		select {
		case <-signalCh:
			_ = platform.TerminateProcessTree(cmd)
			select {
			case <-done:
			case <-time.After(3 * time.Second):
				_ = platform.KillProcessTree(cmd)
			}
		case <-done:
		}
	}()

	err = cmd.Wait()
	close(done)
	wg.Wait()
	signal.Stop(signalCh)

	exit := processExit{code: 0, message: "edge exited"}
	if err != nil {
		exit.message = err.Error()
		exit.code = 1
		if cmd.ProcessState != nil {
			exit.code = cmd.ProcessState.ExitCode()
		}
	} else if cmd.ProcessState != nil {
		exit.code = cmd.ProcessState.ExitCode()
	}
	emit(HelperEvent{Type: "exit", Code: exit.code, Message: exit.message})
	return exit.code
}

func helperScanPipe(wg *sync.WaitGroup, emitMu *sync.Mutex, encoder *json.Encoder, stream string, reader io.Reader) {
	defer wg.Done()
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		emitMu.Lock()
		_ = encoder.Encode(HelperEvent{
			Type:   "log",
			Time:   time.Now().Format(time.RFC3339),
			Stream: stream,
			Line:   scanner.Text(),
		})
		emitMu.Unlock()
	}
	if err := scanner.Err(); err != nil && !errors.Is(err, io.EOF) {
		emitMu.Lock()
		defer emitMu.Unlock()
		_ = encoder.Encode(HelperEvent{
			Type:    "log",
			Time:    time.Now().Format(time.RFC3339),
			Stream:  "stderr",
			Line:    "helper scanner error: " + err.Error(),
			Message: err.Error(),
		})
	}
}
