package agent

import (
	"bufio"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
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
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/haiwo-ci/haiwo/internal/domain"
	"github.com/haiwo-ci/haiwo/internal/jsonrpc"
)

type Config struct {
	ServerURL  string
	Token      string
	AgentID    string
	Name       string
	Version    string
	Labels     []string
	WorkDir    string
	MaxRunning int
}

type Agent struct {
	cfg     Config
	peer    *jsonrpc.Peer
	mu      sync.Mutex
	running map[string]context.CancelFunc
}

func New(cfg Config) *Agent {
	if cfg.MaxRunning <= 0 {
		cfg.MaxRunning = 1
	}
	return &Agent{cfg: cfg, running: map[string]context.CancelFunc{}}
}

func (a *Agent) Run(ctx context.Context) error {
	headers := http.Header{}
	headers.Set("Authorization", "Bearer "+a.cfg.Token)
	conn, _, err := websocket.DefaultDialer.DialContext(ctx, a.cfg.ServerURL, headers)
	if err != nil {
		return err
	}
	a.peer = jsonrpc.NewPeer(conn)
	a.peer.Handle("task.run", a.handleTaskRun)
	a.peer.Handle("task.cancel", a.handleTaskCancel)
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
	a.mu.Unlock()
	status := domain.AgentOnline
	if running >= a.cfg.MaxRunning {
		status = domain.AgentBusy
	}
	return domain.AgentInfo{
		ID:         a.cfg.AgentID,
		Name:       a.cfg.Name,
		Version:    a.cfg.Version + " " + runtime.GOOS + "/" + runtime.GOARCH,
		Labels:     a.cfg.Labels,
		Status:     status,
		CurrentRun: running,
		MaxRunning: a.cfg.MaxRunning,
		LastSeenAt: time.Now(),
	}
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
	case domain.JobDBBackup:
		result.BackupID, result.Artifacts, err = a.backupDatabase(ctx, task)
	case domain.JobDBRestore, domain.JobRollbackDB:
		err = a.restoreDatabase(ctx, task)
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
		if err := a.runShell(ctx, task, dir, command, taskEnv(task)); err != nil {
			return err
		}
	}
	return nil
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

func (a *Agent) backupDatabase(ctx context.Context, task domain.TaskPayload) (string, map[string]string, error) {
	db := task.DatabaseTarget
	if db == nil {
		return "", nil, errors.New("database_target is required")
	}
	backupID := "bak_" + time.Now().Format("20060102150405")
	dir := filepath.Join(a.cfg.WorkDir, "backups", backupID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", nil, err
	}
	switch db.Type {
	case domain.DatabaseMySQL:
		path := filepath.Join(dir, db.Database+".sql.gz")
		err := a.mysqlDump(ctx, task, *db, path)
		return backupID, artifact(path), err
	case domain.DatabasePostgreSQL:
		path := filepath.Join(dir, db.Database+".dump")
		err := a.pgDump(ctx, task, *db, path)
		return backupID, artifact(path), err
	default:
		return "", nil, fmt.Errorf("unsupported database type %q", db.Type)
	}
}

func (a *Agent) restoreDatabase(ctx context.Context, task domain.TaskPayload) error {
	db := task.DatabaseTarget
	if db == nil {
		return errors.New("database_target is required")
	}
	if db.Environment == "prod" && !db.Confirm {
		return errors.New("production restore requires confirm=true")
	}
	if task.BackupID == "" {
		return errors.New("backup_id is required")
	}
	backupDir := filepath.Join(a.cfg.WorkDir, "backups", task.BackupID)
	switch db.Type {
	case domain.DatabaseMySQL:
		return a.mysqlRestore(ctx, task, *db, filepath.Join(backupDir, db.Database+".sql.gz"))
	case domain.DatabasePostgreSQL:
		return a.pgRestore(ctx, task, *db, filepath.Join(backupDir, db.Database+".dump"))
	default:
		return fmt.Errorf("unsupported database type %q", db.Type)
	}
}

func (a *Agent) mysqlDump(ctx context.Context, task domain.TaskPayload, db domain.DatabaseTarget, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	gz := gzip.NewWriter(file)
	defer gz.Close()
	args := []string{"--single-transaction", "--routines", "--triggers", "--events", "-h", db.Host, "-P", fmt.Sprint(port(db.Port, 3306)), "-u", db.Username, db.Database}
	cmd := exec.CommandContext(ctx, "mysqldump", args...)
	cmd.Env = append(os.Environ(), "MYSQL_PWD="+db.Password)
	cmd.Stdout = gz
	cmd.Stderr = logWriter{fn: func(line string) { a.sendLog(task, "stderr", line) }}
	return cmd.Run()
}

func (a *Agent) mysqlRestore(ctx context.Context, task domain.TaskPayload, db domain.DatabaseTarget, path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	gz, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gz.Close()
	args := []string{"-h", db.Host, "-P", fmt.Sprint(port(db.Port, 3306)), "-u", db.Username, db.Database}
	cmd := exec.CommandContext(ctx, "mysql", args...)
	cmd.Env = append(os.Environ(), "MYSQL_PWD="+db.Password)
	cmd.Stdin = gz
	cmd.Stdout = logWriter{fn: func(line string) { a.sendLog(task, "stdout", line) }}
	cmd.Stderr = logWriter{fn: func(line string) { a.sendLog(task, "stderr", line) }}
	return cmd.Run()
}

func (a *Agent) pgDump(ctx context.Context, task domain.TaskPayload, db domain.DatabaseTarget, path string) error {
	args := []string{"-Fc", "-h", db.Host, "-p", fmt.Sprint(port(db.Port, 5432)), "-U", db.Username, "-d", db.Database, "-f", path}
	cmd := exec.CommandContext(ctx, "pg_dump", args...)
	cmd.Env = append(os.Environ(), "PGPASSWORD="+db.Password)
	cmd.Stdout = logWriter{fn: func(line string) { a.sendLog(task, "stdout", line) }}
	cmd.Stderr = logWriter{fn: func(line string) { a.sendLog(task, "stderr", line) }}
	return cmd.Run()
}

func (a *Agent) pgRestore(ctx context.Context, task domain.TaskPayload, db domain.DatabaseTarget, path string) error {
	args := []string{"--clean", "--if-exists", "-h", db.Host, "-p", fmt.Sprint(port(db.Port, 5432)), "-U", db.Username, "-d", db.Database, path}
	cmd := exec.CommandContext(ctx, "pg_restore", args...)
	cmd.Env = append(os.Environ(), "PGPASSWORD="+db.Password)
	cmd.Stdout = logWriter{fn: func(line string) { a.sendLog(task, "stdout", line) }}
	cmd.Stderr = logWriter{fn: func(line string) { a.sendLog(task, "stderr", line) }}
	return cmd.Run()
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

func artifact(path string) map[string]string {
	info, err := os.Stat(path)
	size := int64(0)
	if err == nil {
		size = info.Size()
	}
	return map[string]string{
		"path":     path,
		"size":     fmt.Sprint(size),
		"checksum": checksum(path),
	}
}

func checksum(path string) string {
	file, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer file.Close()
	h := sha256.New()
	_, _ = io.Copy(h, file)
	return hex.EncodeToString(h.Sum(nil))
}

func port(value, fallback int) int {
	if value == 0 {
		return fallback
	}
	return value
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

type logWriter struct {
	fn func(string)
}

func (w logWriter) Write(p []byte) (int, error) {
	text := strings.TrimRight(string(p), "\n")
	if text != "" {
		for _, line := range strings.Split(text, "\n") {
			w.fn(line)
		}
	}
	return len(p), nil
}
