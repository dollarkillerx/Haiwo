package agent

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/creack/pty"
	"github.com/gorilla/websocket"
	"github.com/haiwo-ci/haiwo/internal/domain"
	"github.com/haiwo-ci/haiwo/internal/jsonrpc"
)

type Config struct {
	ServerURL     string
	Token         string
	AgentID       string
	Name          string
	Version       string
	Labels        []string
	WorkDir       string
	MaxRunning    int
	SSHEnabled    bool
	ReverseSSHURL string
}

type Agent struct {
	cfg     Config
	peer    *jsonrpc.Peer
	mu      sync.Mutex
	running map[string]context.CancelFunc
	terms   map[string]*terminalSession
	cpuPrev *cpuSample
}

type cpuSample struct {
	idle  uint64
	total uint64
}

type terminalSession struct {
	cmd  *exec.Cmd
	file *os.File
	mu   sync.Mutex
}

func New(cfg Config) *Agent {
	if cfg.MaxRunning <= 0 {
		cfg.MaxRunning = 1
	}
	return &Agent{cfg: cfg, running: map[string]context.CancelFunc{}, terms: map[string]*terminalSession{}}
}

func (a *Agent) Run(ctx context.Context) error {
	backoff := time.Second
	for {
		if err := a.runOnce(ctx); err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			log.Printf("agent connection failed: %v; reconnecting in %s", err, backoff)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoff):
			}
			backoff = minDuration(backoff*2, 30*time.Second)
			continue
		}
		backoff = time.Second
	}
}

func (a *Agent) runOnce(ctx context.Context) error {
	headers := http.Header{}
	headers.Set("Authorization", "Bearer "+a.cfg.Token)
	conn, _, err := websocket.DefaultDialer.DialContext(ctx, a.cfg.ServerURL, headers)
	if err != nil {
		return err
	}
	defer a.closeTerminals()
	a.peer = jsonrpc.NewPeer(conn)
	a.peer.Handle("task.run", a.handleTaskRun)
	a.peer.Handle("task.cancel", a.handleTaskCancel)
	a.peer.Handle("terminal.open", a.handleTerminalOpen)
	a.peer.Handle("terminal.input", a.handleTerminalInput)
	a.peer.Handle("terminal.resize", a.handleTerminalResize)
	a.peer.Handle("terminal.close", a.handleTerminalClose)
	a.peer.Handle("agent.ping", func(context.Context, json.RawMessage) (any, *jsonrpc.Error) {
		return map[string]string{"status": "ok"}, nil
	})
	a.peer.Handle("agent.updateConfig", a.handleUpdateConfig)

	errCh := make(chan error, 1)
	go func() {
		errCh <- a.peer.Run(ctx)
	}()

	if err := a.peer.Call(ctx, "agent.register", a.info(), nil); err != nil {
		return err
	}
	go a.heartbeat(ctx)

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-errCh:
		return err
	}
}

func minDuration(a, b time.Duration) time.Duration {
	if a < b {
		return a
	}
	return b
}

func (a *Agent) heartbeat(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := a.peer.Notify("agent.heartbeat", a.info()); err != nil {
				log.Printf("heartbeat failed: %v", err)
			}
		}
	}
}

func (a *Agent) info() domain.AgentInfo {
	a.mu.Lock()
	running := len(a.running)
	cpuPercent := a.cpuPercentLocked()
	a.mu.Unlock()
	memUsed, memTotal, memPercent := memoryUsage()
	status := domain.AgentOnline
	if running >= a.cfg.MaxRunning {
		status = domain.AgentBusy
	}
	return domain.AgentInfo{
		ID:            a.cfg.AgentID,
		Name:          a.cfg.Name,
		Version:       a.cfg.Version + " " + runtime.GOOS + "/" + runtime.GOARCH,
		Labels:        a.cfg.Labels,
		Status:        status,
		CurrentRun:    running,
		MaxRunning:    a.cfg.MaxRunning,
		CPUPercent:    cpuPercent,
		MemoryUsed:    memUsed,
		MemoryTotal:   memTotal,
		MemoryPercent: memPercent,
		SSHEnabled:    a.cfg.SSHEnabled,
		ReverseSSHURL: a.cfg.ReverseSSHURL,
		LastSeenAt:    time.Now(),
	}
}

func (a *Agent) cpuPercentLocked() float64 {
	sample, ok := readCPUSample()
	if !ok {
		return 0
	}
	prev := a.cpuPrev
	a.cpuPrev = &sample
	if prev == nil || sample.total <= prev.total || sample.idle < prev.idle {
		return 0
	}
	totalDelta := sample.total - prev.total
	if totalDelta == 0 {
		return 0
	}
	idleDelta := sample.idle - prev.idle
	return float64(totalDelta-idleDelta) * 100 / float64(totalDelta)
}

func readCPUSample() (cpuSample, bool) {
	data, err := os.ReadFile("/proc/stat")
	if err != nil {
		return cpuSample{}, false
	}
	line, _, _ := strings.Cut(string(data), "\n")
	fields := strings.Fields(line)
	if len(fields) < 5 || fields[0] != "cpu" {
		return cpuSample{}, false
	}
	var values []uint64
	for _, field := range fields[1:] {
		value, err := strconv.ParseUint(field, 10, 64)
		if err != nil {
			return cpuSample{}, false
		}
		values = append(values, value)
	}
	var total uint64
	for _, value := range values {
		total += value
	}
	idle := values[3]
	if len(values) > 4 {
		idle += values[4]
	}
	return cpuSample{idle: idle, total: total}, true
}

func memoryUsage() (uint64, uint64, float64) {
	data, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return 0, 0, 0
	}
	values := map[string]uint64{}
	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		value, err := strconv.ParseUint(fields[1], 10, 64)
		if err != nil {
			continue
		}
		values[strings.TrimSuffix(fields[0], ":")] = value * 1024
	}
	total := values["MemTotal"]
	available := values["MemAvailable"]
	if total == 0 || available > total {
		return 0, total, 0
	}
	used := total - available
	return used, total, float64(used) * 100 / float64(total)
}

func (a *Agent) handleTaskRun(ctx context.Context, params json.RawMessage) (any, *jsonrpc.Error) {
	var task domain.TaskPayload
	if err := json.Unmarshal(params, &task); err != nil {
		return nil, &jsonrpc.Error{Code: -32602, Message: err.Error()}
	}
	if task.TaskID == "" {
		return nil, &jsonrpc.Error{Code: -32602, Message: "task_id is required"}
	}
	taskCtx, cancel := context.WithCancel(ctx)
	if task.TimeoutSeconds > 0 {
		taskCtx, cancel = context.WithTimeout(ctx, time.Duration(task.TimeoutSeconds)*time.Second)
	}
	a.mu.Lock()
	if len(a.running) >= a.cfg.MaxRunning {
		a.mu.Unlock()
		cancel()
		return nil, &jsonrpc.Error{Code: -32000, Message: "agent is busy"}
	}
	a.running[task.TaskID] = cancel
	a.mu.Unlock()
	defer func() {
		a.mu.Lock()
		delete(a.running, task.TaskID)
		a.mu.Unlock()
		cancel()
	}()

	complete := a.execute(taskCtx, task)
	_ = a.peer.Notify("task.complete", complete)
	return complete, nil
}

func (a *Agent) handleTaskCancel(_ context.Context, params json.RawMessage) (any, *jsonrpc.Error) {
	var req struct {
		TaskID string `json:"task_id"`
	}
	_ = json.Unmarshal(params, &req)
	a.mu.Lock()
	cancel := a.running[req.TaskID]
	a.mu.Unlock()
	if cancel != nil {
		cancel()
	}
	return map[string]bool{"canceled": cancel != nil}, nil
}

func (a *Agent) handleUpdateConfig(_ context.Context, params json.RawMessage) (any, *jsonrpc.Error) {
	var req struct {
		Labels     []string `json:"labels"`
		MaxRunning int      `json:"max_running"`
		WorkDir    string   `json:"work_dir"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, &jsonrpc.Error{Code: -32602, Message: err.Error()}
	}
	a.mu.Lock()
	if req.Labels != nil {
		a.cfg.Labels = req.Labels
	}
	if req.MaxRunning > 0 {
		a.cfg.MaxRunning = req.MaxRunning
	}
	if req.WorkDir != "" {
		a.cfg.WorkDir = req.WorkDir
	}
	a.mu.Unlock()
	return a.info(), nil
}

func (a *Agent) handleTerminalOpen(ctx context.Context, params json.RawMessage) (any, *jsonrpc.Error) {
	var req struct {
		SessionID string `json:"session_id"`
		Shell     string `json:"shell"`
		Cols      uint16 `json:"cols"`
		Rows      uint16 `json:"rows"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, &jsonrpc.Error{Code: -32602, Message: err.Error()}
	}
	if req.SessionID == "" {
		return nil, &jsonrpc.Error{Code: -32602, Message: "session_id is required"}
	}
	shell := req.Shell
	if shell == "" {
		shell = defaultShell()
	}
	args := []string{}
	switch filepath.Base(shell) {
	case "bash", "zsh":
		args = append(args, "-l")
	}
	cmd := exec.CommandContext(ctx, shell, args...)
	cmd.Dir = a.cfg.WorkDir
	if cmd.Dir == "" {
		cmd.Dir = "."
	}
	f, err := pty.Start(cmd)
	if err != nil {
		return nil, &jsonrpc.Error{Code: -32000, Message: err.Error()}
	}
	if req.Cols == 0 {
		req.Cols = 100
	}
	if req.Rows == 0 {
		req.Rows = 30
	}
	_ = pty.Setsize(f, &pty.Winsize{Cols: req.Cols, Rows: req.Rows})

	session := &terminalSession{cmd: cmd, file: f}
	a.mu.Lock()
	a.terms[req.SessionID] = session
	a.mu.Unlock()

	go a.streamTerminal(req.SessionID, session)
	return map[string]bool{"opened": true}, nil
}

func (a *Agent) handleTerminalInput(_ context.Context, params json.RawMessage) (any, *jsonrpc.Error) {
	var req struct {
		SessionID string `json:"session_id"`
		Data      string `json:"data"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, &jsonrpc.Error{Code: -32602, Message: err.Error()}
	}
	session := a.terminal(req.SessionID)
	if session == nil {
		return nil, &jsonrpc.Error{Code: -32004, Message: "terminal session not found"}
	}
	session.mu.Lock()
	_, err := session.file.Write([]byte(req.Data))
	session.mu.Unlock()
	if err != nil {
		return nil, &jsonrpc.Error{Code: -32000, Message: err.Error()}
	}
	return map[string]bool{"ok": true}, nil
}

func (a *Agent) handleTerminalResize(_ context.Context, params json.RawMessage) (any, *jsonrpc.Error) {
	var req struct {
		SessionID string `json:"session_id"`
		Cols      uint16 `json:"cols"`
		Rows      uint16 `json:"rows"`
	}
	if err := json.Unmarshal(params, &req); err != nil {
		return nil, &jsonrpc.Error{Code: -32602, Message: err.Error()}
	}
	session := a.terminal(req.SessionID)
	if session == nil {
		return nil, &jsonrpc.Error{Code: -32004, Message: "terminal session not found"}
	}
	if req.Cols == 0 || req.Rows == 0 {
		return map[string]bool{"ok": true}, nil
	}
	if err := pty.Setsize(session.file, &pty.Winsize{Cols: req.Cols, Rows: req.Rows}); err != nil {
		return nil, &jsonrpc.Error{Code: -32000, Message: err.Error()}
	}
	return map[string]bool{"ok": true}, nil
}

func (a *Agent) handleTerminalClose(_ context.Context, params json.RawMessage) (any, *jsonrpc.Error) {
	var req struct {
		SessionID string `json:"session_id"`
	}
	_ = json.Unmarshal(params, &req)
	a.closeTerminal(req.SessionID)
	return map[string]bool{"closed": true}, nil
}

func (a *Agent) streamTerminal(sessionID string, session *terminalSession) {
	buf := make([]byte, 4096)
	for {
		n, err := session.file.Read(buf)
		if n > 0 && a.peer != nil {
			_ = a.peer.Notify("terminal.output", map[string]string{
				"session_id": sessionID,
				"data":       string(buf[:n]),
			})
		}
		if err != nil {
			a.closeTerminal(sessionID)
			if a.peer != nil {
				_ = a.peer.Notify("terminal.closed", map[string]string{"session_id": sessionID})
			}
			return
		}
	}
}

func (a *Agent) terminal(sessionID string) *terminalSession {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.terms[sessionID]
}

func (a *Agent) closeTerminal(sessionID string) {
	a.mu.Lock()
	session := a.terms[sessionID]
	delete(a.terms, sessionID)
	a.mu.Unlock()
	if session == nil {
		return
	}
	_ = session.file.Close()
	if session.cmd.Process != nil {
		_ = session.cmd.Process.Kill()
	}
}

func (a *Agent) closeTerminals() {
	a.mu.Lock()
	ids := make([]string, 0, len(a.terms))
	for id := range a.terms {
		ids = append(ids, id)
	}
	a.mu.Unlock()
	for _, id := range ids {
		a.closeTerminal(id)
	}
}

func defaultShell() string {
	for _, shell := range []string{os.Getenv("SHELL"), "/bin/bash", "/bin/sh"} {
		if shell == "" {
			continue
		}
		if _, err := os.Stat(shell); err == nil {
			return shell
		}
	}
	return "/bin/sh"
}

func (a *Agent) execute(ctx context.Context, task domain.TaskPayload) domain.TaskComplete {
	result := domain.TaskComplete{
		TaskID:    task.TaskID,
		RunID:     task.RunID,
		JobID:     task.JobID,
		Status:    domain.JobSuccess,
		Artifacts: map[string]string{},
	}
	var err error
	switch task.JobType {
	case domain.JobCommand:
		err = a.runCommands(ctx, task)
	case domain.JobGitCheckout, domain.JobRollbackCode:
		err = a.gitCheckout(ctx, task)
	default:
		err = fmt.Errorf("unsupported job_type %q", task.JobType)
	}
	if err != nil {
		result.Status = domain.JobFailed
		result.ExitCode = exitCode(err)
		result.Error = err.Error()
	}
	return result
}

func (a *Agent) runCommands(ctx context.Context, task domain.TaskPayload) error {
	if len(task.Commands) == 0 {
		return errors.New("commands are required")
	}
	dir := a.taskDir(task)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	for _, command := range task.Commands {
		a.sendLog(task, "stdout", "$ "+command)
	}
	return a.runShellScript(ctx, task, dir, task.Commands, taskEnv(task))
}

func (a *Agent) gitCheckout(ctx context.Context, task domain.TaskPayload) error {
	if task.Repo == nil || task.Repo.URL == "" {
		return errors.New("repo.url is required")
	}
	dir := a.taskDir(task)
	if err := os.MkdirAll(filepath.Dir(dir), 0o755); err != nil {
		return err
	}
	if _, err := os.Stat(filepath.Join(dir, ".git")); err != nil {
		if err := a.runShell(ctx, task, filepath.Dir(dir), "git clone "+shellQuote(task.Repo.URL)+" "+shellQuote(dir), taskEnv(task)); err != nil {
			return err
		}
	}
	ref := task.Repo.Ref
	if ref == "" {
		ref = task.Ref
	}
	if ref == "" {
		ref = "main"
	}
	if err := a.runShell(ctx, task, dir, "git fetch --all --tags --prune", taskEnv(task)); err != nil {
		return err
	}
	if task.Repo.Commit != "" {
		ref = task.Repo.Commit
	}
	return a.runShell(ctx, task, dir, "git checkout "+shellQuote(ref), taskEnv(task))
}

func (a *Agent) runShell(ctx context.Context, task domain.TaskPayload, dir, command string, env []string) error {
	a.sendLog(task, "stdout", "$ "+command)
	cmd := exec.CommandContext(ctx, "sh", "-c", command)
	cmd.Dir = dir
	cmd.Env = env
	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()
	if err := cmd.Start(); err != nil {
		return err
	}
	go a.scan(task, "stdout", stdout)
	go a.scan(task, "stderr", stderr)
	return cmd.Wait()
}

func (a *Agent) runShellScript(ctx context.Context, task domain.TaskPayload, dir string, commands []string, env []string) error {
	scriptPath := filepath.Join(dir, ".haiwo-task.sh")
	script := "#!/bin/sh\nset -e\n" + strings.Join(commands, "\n") + "\n"
	if err := os.WriteFile(scriptPath, []byte(script), 0o700); err != nil {
		return err
	}
	cmd := exec.CommandContext(ctx, "sh", scriptPath)
	cmd.Dir = dir
	cmd.Env = env
	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()
	if err := cmd.Start(); err != nil {
		return err
	}
	go a.scan(task, "stdout", stdout)
	go a.scan(task, "stderr", stderr)
	return cmd.Wait()
}

func (a *Agent) scan(task domain.TaskPayload, stream string, r io.Reader) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		a.sendLog(task, stream, scanner.Text())
	}
}

func (a *Agent) sendLog(task domain.TaskPayload, stream, line string) {
	if a.peer == nil {
		return
	}
	_ = a.peer.Notify("task.log", domain.TaskLog{TaskID: task.TaskID, Stream: stream, Line: line, Time: time.Now()})
}

func (a *Agent) taskDir(task domain.TaskPayload) string {
	if task.WorkingDir != "" {
		return task.WorkingDir
	}
	return filepath.Join(a.cfg.WorkDir, "tasks", task.TaskID)
}

func taskEnv(task domain.TaskPayload) []string {
	env := os.Environ()
	for k, v := range task.Env {
		env = append(env, k+"="+v)
	}
	for k, v := range task.Secrets {
		env = append(env, k+"="+v)
	}
	return env
}

func exitCode(err error) int {
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return exitErr.ExitCode()
	}
	return 1
}

func shellQuote(v string) string {
	return "'" + strings.ReplaceAll(v, "'", "'\\''") + "'"
}
