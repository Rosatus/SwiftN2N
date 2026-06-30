package edge

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"SwiftN2N/internal/platform"
	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

const (
	EventStatus = "edge:status"
	EventLog    = "edge:log"
	EventExit   = "edge:exit"
)

var ErrAlreadyRunning = errors.New("edge is already running")

type Manager struct {
	mu sync.Mutex

	ctx         context.Context
	cancel      context.CancelFunc
	cmd         *exec.Cmd
	done        chan processExit
	usingHelper bool

	status              Status
	secrets             []string
	supernodeNoResponse int

	logMu     sync.Mutex
	noisyLogs map[string]*noisyLogBucket
}

type noisyLogBucket struct {
	count       int
	lastSummary time.Time
}

func NewManager() *Manager {
	now := time.Now().Format(time.RFC3339)
	return &Manager{
		status: Status{
			State:     StateIdle,
			Message:   "edge is idle",
			UpdatedAt: now,
		},
	}
}

func (m *Manager) SetContext(ctx context.Context) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ctx = ctx
}

func (m *Manager) Status() Status {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.status
}

func (m *Manager) Start(config Config) (Status, error) {
	edgePath, spec, err := m.prepareCommand(config)
	if err != nil {
		status := m.setStatus(StateFailed, err.Error(), 0, err.Error(), edgePath)
		return status, err
	}
	if !platform.IsElevated() && platform.CanUsePrivilegedHelper() {
		return m.startWithHelper(config, edgePath, spec)
	}
	if !platform.IsElevated() {
		err := errors.New("root privileges or a working pkexec helper are required to start edge")
		status := m.setStatus(StateFailed, err.Error(), 0, err.Error(), edgePath)
		return status, err
	}
	return m.startDirect(edgePath, spec)
}

func (m *Manager) startDirect(edgePath string, spec CommandSpec) (Status, error) {
	m.mu.Lock()
	if m.cmd != nil && m.cmd.Process != nil {
		status := m.status
		m.mu.Unlock()
		return status, ErrAlreadyRunning
	}

	ctx, cancel := context.WithCancel(context.Background())
	cmd := exec.CommandContext(ctx, edgePath, spec.Args...)
	cmd.Env = append(os.Environ(), spec.Env...)
	platform.ConfigureProcess(cmd)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		cancel()
		m.mu.Unlock()
		status := m.setStatus(StateFailed, err.Error(), 0, err.Error(), edgePath)
		return status, err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		cancel()
		m.mu.Unlock()
		status := m.setStatus(StateFailed, err.Error(), 0, err.Error(), edgePath)
		return status, err
	}

	if err := cmd.Start(); err != nil {
		cancel()
		m.mu.Unlock()
		status := m.setStatus(StateFailed, err.Error(), 0, err.Error(), edgePath)
		return status, err
	}

	now := time.Now().Format(time.RFC3339)
	done := make(chan processExit, 1)
	m.cancel = cancel
	m.cmd = cmd
	m.done = done
	m.usingHelper = false
	m.secrets = append([]string(nil), spec.Secrets...)
	m.supernodeNoResponse = 0
	m.resetLogNoiseFilter()
	m.status = Status{
		State:     StateStarting,
		Message:   "edge process started",
		PID:       cmd.Process.Pid,
		EdgePath:  edgePath,
		StartedAt: now,
		UpdatedAt: now,
	}
	status := m.status
	runtimeCtx := m.ctx
	m.mu.Unlock()

	if runtimeCtx != nil {
		wailsruntime.EventsEmit(runtimeCtx, EventStatus, status)
	}
	m.emitLog("system", "starting edge "+edgePath+" "+strings.Join(RedactArgs(spec.Args), " "))
	go m.scanPipe("stdout", stdout)
	go m.scanPipe("stderr", stderr)
	go m.waitForExit(cmd, cancel, done)

	return status, nil
}

func (m *Manager) startWithHelper(config Config, edgePath string, spec CommandSpec) (Status, error) {
	exe, err := os.Executable()
	if err != nil {
		status := m.setStatus(StateFailed, err.Error(), 0, err.Error(), edgePath)
		return status, err
	}

	m.mu.Lock()
	if m.cmd != nil && m.cmd.Process != nil {
		status := m.status
		m.mu.Unlock()
		return status, ErrAlreadyRunning
	}

	ctx, cancel := context.WithCancel(context.Background())
	cmd, err := platform.PrivilegedHelperCommand(ctx, exe, HelperFlag, HelperEdge)
	if err != nil {
		cancel()
		m.mu.Unlock()
		status := m.setStatus(StateFailed, err.Error(), 0, err.Error(), edgePath)
		return status, err
	}
	platform.ConfigureProcess(cmd)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		cancel()
		m.mu.Unlock()
		status := m.setStatus(StateFailed, err.Error(), 0, err.Error(), edgePath)
		return status, err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		cancel()
		m.mu.Unlock()
		status := m.setStatus(StateFailed, err.Error(), 0, err.Error(), edgePath)
		return status, err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		cancel()
		m.mu.Unlock()
		status := m.setStatus(StateFailed, err.Error(), 0, err.Error(), edgePath)
		return status, err
	}

	if err := cmd.Start(); err != nil {
		cancel()
		m.mu.Unlock()
		status := m.setStatus(StateFailed, err.Error(), 0, err.Error(), edgePath)
		return status, err
	}
	if err := json.NewEncoder(stdin).Encode(config); err != nil {
		_ = stdin.Close()
		_ = platform.KillProcessTree(cmd)
		cancel()
		m.mu.Unlock()
		status := m.setStatus(StateFailed, err.Error(), 0, err.Error(), edgePath)
		return status, err
	}
	_ = stdin.Close()

	now := time.Now().Format(time.RFC3339)
	done := make(chan processExit, 1)
	m.cancel = cancel
	m.cmd = cmd
	m.done = done
	m.usingHelper = true
	m.secrets = append([]string(nil), spec.Secrets...)
	m.supernodeNoResponse = 0
	m.resetLogNoiseFilter()
	m.status = Status{
		State:     StateStarting,
		Message:   "privileged helper started",
		PID:       cmd.Process.Pid,
		EdgePath:  edgePath,
		StartedAt: now,
		UpdatedAt: now,
	}
	status := m.status
	runtimeCtx := m.ctx
	m.mu.Unlock()

	if runtimeCtx != nil {
		wailsruntime.EventsEmit(runtimeCtx, EventStatus, status)
	}
	m.emitLog("system", "starting edge through privileged helper "+edgePath+" "+strings.Join(RedactArgs(spec.Args), " "))
	go m.scanHelperPipe(stdout)
	go m.scanPipe("stderr", stderr)
	go m.waitForExit(cmd, cancel, done)

	return status, nil
}

func (m *Manager) Stop() (Status, error) {
	m.mu.Lock()
	cmd := m.cmd
	done := m.done
	cancel := m.cancel
	usingHelper := m.usingHelper
	if cmd == nil || cmd.Process == nil {
		m.mu.Unlock()
		return m.setStatus(StateIdle, "edge is not running", 0, "", ""), nil
	}
	pid := cmd.Process.Pid
	edgePath := m.status.EdgePath
	m.mu.Unlock()

	status := m.setStatus(StateStopping, "stopping edge process", pid, "", edgePath)
	if usingHelper && !platform.IsElevated() && platform.CanUsePrivilegedHelper() {
		if err := m.stopPrivilegedHelper(pid); err != nil {
			m.emitLog("stderr", "privileged stop failed: "+err.Error())
		}
	} else if err := platform.TerminateProcessTree(cmd); err != nil && cancel != nil {
		cancel()
	}

	select {
	case <-done:
	case <-time.After(3 * time.Second):
		if usingHelper && !platform.IsElevated() && platform.CanUsePrivilegedHelper() {
			if err := m.stopPrivilegedHelper(pid); err != nil {
				m.emitLog("stderr", "privileged forced stop failed: "+err.Error())
			}
		} else {
			_ = platform.KillProcessTree(cmd)
		}
		select {
		case <-done:
		case <-time.After(7 * time.Second):
			status = m.setStatus(StateFailed, "edge process did not exit after forced kill", pid, "stop timeout", edgePath)
			return status, errors.New("edge process did not exit after forced kill")
		}
	}

	return m.setStatus(StateIdle, "edge stopped", 0, "", edgePath), nil
}

func (m *Manager) stopPrivilegedHelper(pid int) error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	cmd, err := platform.PrivilegedHelperCommand(ctx, exe, HelperFlag, HelperStop, strconv.Itoa(pid))
	if err != nil {
		return err
	}
	output, err := cmd.CombinedOutput()
	text := strings.TrimSpace(string(output))
	if ctx.Err() == context.DeadlineExceeded {
		return errors.New("privileged stop timed out")
	}
	if err != nil {
		if text != "" {
			return errors.New(text)
		}
		return err
	}
	if text != "" {
		m.emitLog("system", text)
	}
	return nil
}

func (m *Manager) Version(config Config) (string, error) {
	edgePath, err := ResolvePath(config.EdgePath)
	if err != nil {
		return "", err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, edgePath, "-h")
	platform.ConfigureProcess(cmd)
	output, err := cmd.CombinedOutput()
	if ctx.Err() == context.DeadlineExceeded {
		return "", errors.New("edge version command timed out")
	}
	text := strings.TrimSpace(string(output))
	if err != nil && text == "" {
		return "", err
	}
	return text, nil
}

func (m *Manager) Validate(config Config) error {
	if _, err := BuildCommand(config); err != nil {
		return err
	}
	path, err := ResolvePath(config.EdgePath)
	if err != nil {
		return err
	}
	if !platform.IsExecutable(path) {
		return errors.New("edge binary is not executable")
	}
	return nil
}

func (m *Manager) Environment(config Config) EnvironmentStatus {
	helperAvailable := platform.CanUsePrivilegedHelper()
	status := EnvironmentStatus{
		OS:              runtime.GOOS,
		Arch:            runtime.GOARCH,
		Elevated:        platform.IsElevated(),
		HelperAvailable: helperAvailable,
		Missing:         []string{},
	}
	if !status.Elevated && !status.HelperAvailable {
		status.Missing = append(status.Missing, "administrator/root privileges or pkexec helper")
	}
	if path, err := ResolvePath(config.EdgePath); err == nil {
		status.EdgeFound = true
		status.EdgePath = path
		if !platform.IsExecutable(path) {
			status.Missing = append(status.Missing, "edge executable permission")
		}
		if version, err := m.Version(config); err == nil {
			status.EdgeVersion = firstMeaningfulLine(version)
		}
	} else {
		status.Missing = append(status.Missing, "edge binary")
	}

	switch {
	case len(status.Missing) == 0:
		if status.Elevated {
			status.Message = "environment ready"
		} else {
			status.Message = "privileged helper ready"
		}
	case !status.EdgeFound:
		status.Message = "edge binary not found"
	default:
		status.Message = "edge requires root privileges or a working pkexec helper"
	}
	return status
}

func (m *Manager) prepareCommand(config Config) (string, CommandSpec, error) {
	edgePath, err := ResolvePath(config.EdgePath)
	if err != nil {
		return "", CommandSpec{}, err
	}
	spec, err := BuildCommand(config)
	if err != nil {
		return edgePath, CommandSpec{}, err
	}
	return edgePath, spec, nil
}

func (m *Manager) scanPipe(stream string, reader io.Reader) {
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		line := scanner.Text()
		m.emitLog(stream, line)
		m.parseLogLine(line)
	}
	if err := scanner.Err(); err != nil {
		m.emitLog(stream, "log scanner error: "+err.Error())
	}
}

func (m *Manager) scanHelperPipe(reader io.Reader) {
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		line := scanner.Text()
		var event HelperEvent
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			m.emitLog("system", line)
			continue
		}
		switch event.Type {
		case "log":
			stream := event.Stream
			if stream == "" {
				stream = "stdout"
			}
			m.emitLog(stream, event.Line)
			m.parseLogLine(event.Line)
		case "status":
			m.emitLog("system", event.Message)
		case "error":
			m.emitLog("stderr", event.Message)
			m.setStatus(StateFailed, event.Message, 0, event.Message, "")
		case "exit":
			m.emitLog("system", event.Message)
		default:
			m.emitLog("system", line)
		}
	}
	if err := scanner.Err(); err != nil {
		m.emitLog("stderr", "helper scanner error: "+err.Error())
	}
}

func (m *Manager) waitForExit(cmd *exec.Cmd, cancel context.CancelFunc, done chan processExit) {
	err := cmd.Wait()
	cancel()
	m.flushSuppressedLogs()

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

	done <- exit
	close(done)

	m.mu.Lock()
	wasCurrent := m.cmd == cmd
	previous := m.status.State
	edgePath := m.status.EdgePath
	if wasCurrent {
		m.cmd = nil
		m.cancel = nil
		m.done = nil
		m.usingHelper = false
	}
	m.mu.Unlock()

	if !wasCurrent {
		return
	}

	if previous == StateStopping || exit.code == 0 {
		m.setStatus(StateIdle, "edge stopped", 0, "", edgePath)
	} else {
		m.setStatus(StateFailed, exit.message, 0, exit.message, edgePath)
	}
	m.emitExit(exit)

	m.mu.Lock()
	if m.cmd == nil {
		m.secrets = nil
		m.usingHelper = false
	}
	m.mu.Unlock()
}

func (m *Manager) emitLog(stream string, line string) {
	if m.shouldSuppressNoisyLog(stream, line) {
		return
	}
	m.emitLogDirect(stream, line)
}

func (m *Manager) emitLogDirect(stream string, line string) {
	ctx, secrets := m.currentContextAndSecrets()
	if ctx == nil {
		return
	}
	wailsruntime.EventsEmit(ctx, EventLog, LogEvent{
		Time:   time.Now().Format(time.RFC3339),
		Stream: stream,
		Line:   RedactSecrets(line, secrets),
	})
}

func (m *Manager) shouldSuppressNoisyLog(stream string, line string) bool {
	if stream != "stdout" && stream != "stderr" {
		return false
	}
	label, ok := noisyLogLabel(line)
	if !ok {
		return false
	}

	now := time.Now()
	m.logMu.Lock()
	if m.noisyLogs == nil {
		m.noisyLogs = map[string]*noisyLogBucket{}
	}
	bucket, exists := m.noisyLogs[label]
	if !exists {
		m.noisyLogs[label] = &noisyLogBucket{lastSummary: now}
		m.logMu.Unlock()
		return false
	}

	bucket.count++
	if now.Sub(bucket.lastSummary) < 5*time.Second {
		m.logMu.Unlock()
		return true
	}

	count := bucket.count
	bucket.count = 0
	bucket.lastSummary = now
	m.logMu.Unlock()

	m.emitLogDirect("system", fmt.Sprintf("suppressed %d repeated noisy log lines: %s", count, label))
	return true
}

func (m *Manager) resetLogNoiseFilter() {
	m.logMu.Lock()
	defer m.logMu.Unlock()
	m.noisyLogs = map[string]*noisyLogBucket{}
}

func (m *Manager) flushSuppressedLogs() {
	var summaries []string
	m.logMu.Lock()
	for label, bucket := range m.noisyLogs {
		if bucket.count > 0 {
			summaries = append(summaries, fmt.Sprintf("suppressed %d repeated noisy log lines: %s", bucket.count, label))
			bucket.count = 0
			bucket.lastSummary = time.Now()
		}
	}
	m.logMu.Unlock()

	for _, summary := range summaries {
		m.emitLogDirect("system", summary)
	}
}

func noisyLogLabel(line string) (string, bool) {
	lower := strings.ToLower(line)
	switch {
	case strings.Contains(lower, "dropping tx multicast"):
		return "dropping Tx multicast", true
	case strings.Contains(lower, "drop packet before first registration with supernode"):
		return "DROP packet before first registration with supernode", true
	case strings.Contains(lower, "rx tap packet"):
		return "Rx TAP packet", true
	case strings.Contains(lower, "tx packet of") && strings.Contains(lower, "ff:ff:ff:ff:ff:ff"):
		return "broadcast Tx packet", true
	default:
		return "", false
	}
}

func (m *Manager) emitExit(exit processExit) {
	ctx, secrets := m.currentContextAndSecrets()
	if ctx == nil {
		return
	}
	wailsruntime.EventsEmit(ctx, EventExit, ExitEvent{
		Time:    time.Now().Format(time.RFC3339),
		Code:    exit.code,
		Message: RedactSecrets(exit.message, secrets),
	})
}

func (m *Manager) parseLogLine(line string) {
	m.mu.Lock()
	nextCount := m.supernodeNoResponse
	if strings.Contains(strings.ToLower(line), "no response from supernode") || strings.Contains(strings.ToLower(line), "supernode not responding") {
		nextCount++
	}
	classification := ClassifyLogLine(line, nextCount)
	if classification.NoReply {
		m.supernodeNoResponse = nextCount
	}
	if classification.Connected {
		m.supernodeNoResponse = 0
	}
	m.mu.Unlock()

	if classification.Matched {
		m.setStatus(classification.State, classification.Message, 0, classification.LastError, "")
	}
}

func (m *Manager) setStatus(state State, message string, pid int, lastError string, edgePath string) Status {
	m.mu.Lock()
	if pid == 0 && m.status.PID != 0 && state != StateIdle && state != StateFailed {
		pid = m.status.PID
	}
	if edgePath == "" {
		edgePath = m.status.EdgePath
	}
	if message == "" {
		message = m.status.Message
	}
	if lastError == "" && state != StateFailed {
		lastError = ""
	}
	now := time.Now().Format(time.RFC3339)
	m.status.State = state
	m.status.Message = message
	m.status.PID = pid
	m.status.EdgePath = edgePath
	m.status.UpdatedAt = now
	m.status.LastError = lastError
	status := m.status
	ctx := m.ctx
	m.mu.Unlock()

	if ctx != nil {
		wailsruntime.EventsEmit(ctx, EventStatus, status)
	}
	return status
}

func (m *Manager) currentContextAndSecrets() (context.Context, []string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.ctx, append([]string(nil), m.secrets...)
}

func firstMeaningfulLine(value string) string {
	line := strings.Split(strings.TrimSpace(value), "\n")
	if len(line) == 0 {
		return ""
	}
	return strings.TrimSpace(line[0])
}
