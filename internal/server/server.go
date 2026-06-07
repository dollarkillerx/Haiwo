package server

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	dbmodel "github.com/haiwo-ci/haiwo/internal/database"
	"github.com/haiwo-ci/haiwo/internal/domain"
	"github.com/haiwo-ci/haiwo/internal/jsonrpc"
	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
)

type Config struct {
	Addr        string
	AgentToken  string
	WebPassword string
	DB          *gorm.DB
	Now         func() time.Time
}

type App struct {
	cfg      Config
	store    *Store
	hub      *AgentHub
	cron     *cron.Cron
	upgrader websocket.Upgrader
}

func New(cfg Config) *App {
	if cfg.Now == nil {
		cfg.Now = time.Now
	}
	app := &App{
		cfg:   cfg,
		store: NewStore(cfg.Now, cfg.DB),
		cron:  cron.New(),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(*http.Request) bool { return true },
		},
	}
	app.hub = NewAgentHub(app.store)
	app.cron.Start()
	return app
}

func (a *App) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("GET /login", a.loginPage)
	mux.HandleFunc("POST /login", a.login)
	mux.HandleFunc("POST /logout", a.logout)
	mux.Handle("GET /logo.png", staticHandler())
	mux.HandleFunc("GET /rpc/agent/ws", a.handleAgentWS)
	mux.HandleFunc("GET /api/projects", a.listProjects)
	mux.HandleFunc("POST /api/projects", a.createProject)
	mux.HandleFunc("GET /api/pipelines", a.listPipelines)
	mux.HandleFunc("POST /api/projects/{id}/pipelines", a.createPipeline)
	mux.HandleFunc("POST /api/projects/{id}/triggers", a.createTrigger)
	mux.HandleFunc("POST /api/webhooks/{provider}/{project_id}", a.handleWebhook)
	mux.HandleFunc("GET /api/runs", a.listRuns)
	mux.HandleFunc("GET /api/runs/{id}", a.getRun)
	mux.HandleFunc("POST /api/pipelines/{id}/runs", a.startPipelineRun)
	mux.HandleFunc("POST /api/runs/{id}/cancel", a.cancelRun)
	mux.HandleFunc("POST /api/runs/{id}/rollback", a.rollbackRun)
	mux.HandleFunc("GET /api/settings", a.getSettings)
	mux.HandleFunc("PUT /api/settings", a.updateSettings)
	mux.HandleFunc("POST /api/agents", a.createAgent)
	mux.HandleFunc("DELETE /api/agents/{id}", a.deleteAgent)
	mux.HandleFunc("GET /api/agents/{id}/deploy-script", a.agentDeployScript)
	mux.HandleFunc("GET /api/agents/{id}/deploy.sh", a.agentDeployShell)
	mux.HandleFunc("GET /api/agents/{id}/ssh-command", a.agentSSHCommand)
	mux.HandleFunc("GET /api/agents", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, a.store.ListAgents())
	})
	mux.Handle("GET /", a.requireWebAuth(staticHandler()))
	return mux
}

func (a *App) handleAgentWS(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Authorization")
	token = strings.TrimPrefix(token, "Bearer ")
	if token == "" {
		token = r.URL.Query().Get("token")
	}
	authAgentID := ""
	if agent, ok := a.store.GetAgentByToken(token); ok {
		authAgentID = agent.ID
	} else if a.cfg.AgentToken == "" || token != a.cfg.AgentToken {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	conn, err := a.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	peer := jsonrpc.NewPeer(conn)
	session := &AgentSession{peer: peer, authAgentID: authAgentID, token: token}
	peer.Handle("agent.register", a.hub.RegisterHandler(session))
	peer.Handle("agent.heartbeat", a.hub.HeartbeatHandler(session))
	peer.Handle("task.log", a.taskLog)
	peer.Handle("task.progress", a.taskProgress)
	peer.Handle("task.complete", a.taskComplete)
	peer.Handle("artifact.uploadComplete", a.artifactUploadComplete)
	if err := peer.Run(r.Context()); err != nil {
		log.Printf("agent rpc disconnected: %v", err)
	}
	a.hub.Disconnect(session)
}

func (a *App) createProject(w http.ResponseWriter, r *http.Request) {
	var p domain.Project
	if !decodeJSON(w, r, &p) {
		return
	}
	p.ID = id("prj")
	p.CreatedAt = a.cfg.Now()
	a.store.SaveProject(p)
	writeJSON(w, http.StatusCreated, p)
}

func (a *App) listProjects(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, a.store.ListProjects())
}

func (a *App) createPipeline(w http.ResponseWriter, r *http.Request) {
	var p domain.Pipeline
	if !decodeJSON(w, r, &p) {
		return
	}
	p.ID = id("pipe")
	p.ProjectID = r.PathValue("id")
	p.CreatedAt = a.cfg.Now()
	a.store.SavePipeline(p)
	writeJSON(w, http.StatusCreated, p)
}

func (a *App) listPipelines(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, a.store.ListPipelines(r.URL.Query().Get("project_id")))
}

func (a *App) createTrigger(w http.ResponseWriter, r *http.Request) {
	var t domain.Trigger
	if !decodeJSON(w, r, &t) {
		return
	}
	t.ID = id("trg")
	t.ProjectID = r.PathValue("id")
	t.CreatedAt = a.cfg.Now()
	a.store.SaveTrigger(t)
	if t.Type == "schedule" && t.Cron != "" {
		_, err := a.cron.AddFunc(t.Cron, func() {
			_, err := a.startRun(context.Background(), t.PipelineID, "schedule", domain.TriggerEvent{Type: "schedule"})
			if err != nil {
				log.Printf("scheduled run failed: %v", err)
			}
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}
	writeJSON(w, http.StatusCreated, t)
}

func (a *App) handleWebhook(w http.ResponseWriter, r *http.Request) {
	var raw map[string]any
	if !decodeJSON(w, r, &raw) {
		return
	}
	event := webhookEvent(r.Header, raw)
	projectID := r.PathValue("project_id")
	triggers := a.store.ListTriggers(projectID)
	started := []domain.Run{}
	for _, t := range triggers {
		if t.Matches(event) {
			run, err := a.startRun(r.Context(), t.PipelineID, "webhook", event)
			if err == nil {
				started = append(started, run)
			}
		}
	}
	writeJSON(w, http.StatusAccepted, map[string]any{"matched": len(started), "runs": started})
}

func (a *App) startPipelineRun(w http.ResponseWriter, r *http.Request) {
	var ev domain.TriggerEvent
	_ = json.NewDecoder(r.Body).Decode(&ev)
	if ev.Type == "" {
		ev.Type = "manual"
	}
	run, err := a.startRun(r.Context(), r.PathValue("id"), "manual", ev)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusAccepted, run)
}

func (a *App) listRuns(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, a.store.ListRuns(r.URL.Query().Get("project_id")))
}

func (a *App) getRun(w http.ResponseWriter, r *http.Request) {
	run, ok := a.store.GetRun(r.PathValue("id"))
	if !ok {
		http.NotFound(w, r)
		return
	}
	writeJSON(w, http.StatusOK, run)
}

func (a *App) cancelRun(w http.ResponseWriter, r *http.Request) {
	run, ok := a.store.GetRun(r.PathValue("id"))
	if !ok {
		http.NotFound(w, r)
		return
	}
	run.Status = domain.RunCanceled
	run.UpdatedAt = a.cfg.Now()
	a.store.SaveRun(run)
	writeJSON(w, http.StatusOK, run)
}

func (a *App) rollbackRun(w http.ResponseWriter, r *http.Request) {
	var req struct {
		PipelineID string `json:"pipeline_id"`
		Ref        string `json:"ref"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	ev := domain.TriggerEvent{Type: "manual", Ref: req.Ref, Comment: "/rollback"}
	run, err := a.startRun(r.Context(), req.PipelineID, "rollback", ev)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	run.Metadata = map[string]string{"source_run_id": r.PathValue("id")}
	run.Status = domain.RunRollbackRunning
	a.store.SaveRun(run)
	writeJSON(w, http.StatusAccepted, run)
}

func (a *App) createAgent(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name       string   `json:"name"`
		Labels     []string `json:"labels"`
		MaxRunning int      `json:"max_running"`
		SSHEnabled bool     `json:"ssh_enabled"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}
	if req.MaxRunning <= 0 {
		req.MaxRunning = 1
	}
	if a.store.AgentNameExists(req.Name) {
		http.Error(w, "agent name already exists", http.StatusConflict)
		return
	}
	now := a.cfg.Now()
	settings := a.store.GetSettings()
	reverseSSHURL := reverseSSHURLFromBase(settings.ServerBaseURL)
	agent := domain.AgentInfo{
		ID:            id("agent"),
		Name:          req.Name,
		Token:         token(),
		Labels:        cleanStrings(req.Labels),
		Status:        domain.AgentOffline,
		MaxRunning:    req.MaxRunning,
		SSHEnabled:    req.SSHEnabled,
		SSHPort:       22,
		ReverseSSHURL: reverseSSHURL,
		CreatedAt:     now,
		LastSeenAt:    now,
	}
	a.store.SaveAgent(agent)
	writeJSON(w, http.StatusCreated, map[string]any{
		"agent":           agent,
		"deployScriptURL": a.deployScriptURL(r, agent),
	})
}

func (a *App) deleteAgent(w http.ResponseWriter, r *http.Request) {
	agentID := r.PathValue("id")
	if _, ok := a.store.GetAgent(agentID); !ok {
		http.NotFound(w, r)
		return
	}
	a.hub.Delete(agentID)
	a.store.DeleteAgent(agentID)
	writeJSON(w, http.StatusOK, map[string]bool{"deleted": true})
}

func (a *App) getSettings(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, a.store.GetSettings())
}

func (a *App) updateSettings(w http.ResponseWriter, r *http.Request) {
	var settings domain.SystemSettings
	if !decodeJSON(w, r, &settings) {
		return
	}
	if settings.ServerBaseURL == "" && settings.ReverseSSHURL != "" {
		settings.ServerBaseURL = baseURLFromReverseSSHURL(settings.ReverseSSHURL)
	}
	settings.ServerBaseURL = normalizeBaseURL(settings.ServerBaseURL)
	settings.ReverseSSHURL = reverseSSHURLFromBase(settings.ServerBaseURL)
	a.store.SaveSettings(settings)
	writeJSON(w, http.StatusOK, settings)
}

func (a *App) agentDeployScript(w http.ResponseWriter, r *http.Request) {
	agent, ok := a.store.GetAgent(r.PathValue("id"))
	if !ok {
		http.NotFound(w, r)
		return
	}
	url := a.deployScriptURL(r, agent)
	writeJSON(w, http.StatusOK, map[string]string{
		"url":     url,
		"command": "curl -fsSL " + shellQuote(url) + " | bash",
	})
}

func (a *App) agentDeployShell(w http.ResponseWriter, r *http.Request) {
	agent, ok := a.store.GetAgent(r.PathValue("id"))
	if !ok {
		http.NotFound(w, r)
		return
	}
	if r.URL.Query().Get("token") != agent.Token {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	w.Header().Set("Content-Type", "text/x-shellscript; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="haiwo-agent-`+safeFilename(agent.Name)+`.sh"`)
	_, _ = w.Write([]byte(a.deployScript(r, agent)))
}

func (a *App) agentSSHCommand(w http.ResponseWriter, r *http.Request) {
	agent, ok := a.store.GetAgent(r.PathValue("id"))
	if !ok {
		http.NotFound(w, r)
		return
	}
	command := ""
	if agent.SSHEnabled && agent.SSHHost != "" {
		user := agent.SSHUser
		if user == "" {
			user = "root"
		}
		port := agent.SSHPort
		if port <= 0 {
			port = 22
		}
		command = "ssh -p " + strconv.Itoa(port) + " " + shellQuote(user+"@"+agent.SSHHost)
	} else if agent.ReverseSSHURL != "" {
		command = "reverse SSH WebSocket configured at " + agent.ReverseSSHURL
	}
	if command == "" {
		http.Error(w, "ssh is not configured for this agent", http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"command": command})
}

func (a *App) startRun(ctx context.Context, pipelineID, source string, event domain.TriggerEvent) (domain.Run, error) {
	pipeline, ok := a.store.GetPipeline(pipelineID)
	if !ok {
		return domain.Run{}, errors.New("pipeline not found")
	}
	run := domain.Run{
		ID:         id("run"),
		ProjectID:  pipeline.ProjectID,
		PipelineID: pipeline.ID,
		Status:     domain.RunQueued,
		Source:     source,
		Ref:        event.Ref,
		Comment:    event.Comment,
		CommitSHA:  event.CommitSHA,
		CreatedAt:  a.cfg.Now(),
		UpdatedAt:  a.cfg.Now(),
	}
	a.store.SaveRun(run)
	go a.runPipeline(context.Background(), run, pipeline)
	return run, nil
}

func (a *App) runPipeline(ctx context.Context, run domain.Run, pipeline domain.Pipeline) {
	run.Status = domain.RunRunning
	run.UpdatedAt = a.cfg.Now()
	a.store.SaveRun(run)
	for _, stage := range pipeline.Stages {
		errCh := make(chan error, len(stage.Jobs))
		for _, job := range stage.Jobs {
			job := job
			go func() { errCh <- a.runJob(ctx, run, job) }()
		}
		for range stage.Jobs {
			if err := <-errCh; err != nil {
				run.Status = domain.RunFailed
				run.UpdatedAt = a.cfg.Now()
				a.store.SaveRun(run)
				return
			}
		}
	}
	run.Status = domain.RunSuccess
	run.UpdatedAt = a.cfg.Now()
	a.store.SaveRun(run)
}

func (a *App) runJob(ctx context.Context, run domain.Run, job domain.Job) error {
	agents := a.hub.Match(job.AgentIDs, job.AgentLabels)
	if len(agents) == 0 {
		return errors.New("no matching online agent")
	}
	if job.AgentMode == "" {
		job.AgentMode = domain.AgentModeSingle
	}
	if job.AgentMode == domain.AgentModeSingle && len(agents) > 1 {
		agents = agents[:1]
	}
	if job.AgentMode == domain.AgentModeChain {
		for _, ag := range agents {
			if err := a.dispatchTask(ctx, ag, run, job); err != nil {
				return err
			}
		}
		return nil
	}
	errCh := make(chan error, len(agents))
	for _, ag := range agents {
		ag := ag
		go func() { errCh <- a.dispatchTask(ctx, ag, run, job) }()
	}
	for range agents {
		if err := <-errCh; err != nil && job.Required {
			return err
		}
	}
	return nil
}

func (a *App) dispatchTask(ctx context.Context, session *AgentSession, run domain.Run, job domain.Job) error {
	payload := domain.TaskPayload{
		TaskID:         id("task"),
		RunID:          run.ID,
		JobID:          job.ID,
		JobType:        job.Type,
		WorkingDir:     job.WorkingDir,
		Commands:       job.Commands,
		Repo:           job.Repo,
		Env:            job.Env,
		Secrets:        job.Secrets,
		TimeoutSeconds: job.TimeoutSeconds,
	}
	if payload.Repo != nil {
		payload.Ref = payload.Repo.Ref
	}
	var result domain.TaskComplete
	if err := session.peer.Call(ctx, "task.run", payload, &result); err != nil {
		return err
	}
	if result.Status != domain.JobSuccess {
		return errors.New(result.Error)
	}
	return nil
}

func (a *App) taskLog(_ context.Context, params json.RawMessage) (any, *jsonrpc.Error) {
	var msg domain.TaskLog
	_ = json.Unmarshal(params, &msg)
	log.Printf("[%s][%s] %s", msg.TaskID, msg.Stream, msg.Line)
	return map[string]bool{"ok": true}, nil
}

func (a *App) taskProgress(_ context.Context, params json.RawMessage) (any, *jsonrpc.Error) {
	var msg domain.TaskProgress
	_ = json.Unmarshal(params, &msg)
	log.Printf("[%s] %d%% %s", msg.TaskID, msg.Percent, msg.Message)
	return map[string]bool{"ok": true}, nil
}

func (a *App) taskComplete(_ context.Context, params json.RawMessage) (any, *jsonrpc.Error) {
	var msg domain.TaskComplete
	_ = json.Unmarshal(params, &msg)
	return map[string]bool{"ok": true}, nil
}

func (a *App) artifactUploadComplete(_ context.Context, _ json.RawMessage) (any, *jsonrpc.Error) {
	return map[string]bool{"ok": true}, nil
}

func (a *App) deployScript(r *http.Request, agent domain.AgentInfo) string {
	serverURL := agentRPCURL(a.baseURL(r))
	labels := strings.Join(agent.Labels, ",")
	lines := []string{
		"#!/usr/bin/env bash",
		"set -euo pipefail",
		"",
		"# Haiwo agent deployment script.",
		": ${HAIWO_AGENT_BINARY_URL:=https://fileoss.hacksnews.top/haiwo-agent-linux-amd64}",
		"export HAIWO_SERVER_URL=" + shellQuote(serverURL),
		"export HAIWO_AGENT_TOKEN=" + shellQuote(agent.Token),
		"export HAIWO_AGENT_ID=" + shellQuote(agent.ID),
		"export HAIWO_AGENT_NAME=" + shellQuote(agent.Name),
		"export HAIWO_AGENT_LABELS=" + shellQuote(labels),
		"export HAIWO_AGENT_MAX_RUNNING=" + shellQuote(strconv.Itoa(max(agent.MaxRunning, 1))),
		": ${HAIWO_AGENT_HOME:=${HOME:-/opt/haiwo}/.haiwo/agents/${HAIWO_AGENT_ID}}",
		": ${HAIWO_AGENT_BIN:=${HAIWO_AGENT_HOME}/haiwo-agent}",
		": ${HAIWO_AGENT_PID_FILE:=${HAIWO_AGENT_HOME}/haiwo-agent.pid}",
		"export HAIWO_AGENT_WORKDIR=${HAIWO_AGENT_WORKDIR:-${HAIWO_AGENT_HOME}/workdir}",
	}
	if agent.ReverseSSHURL != "" {
		lines = append(lines, "export HAIWO_AGENT_REVERSE_SSH_URL="+shellQuote(agent.ReverseSSHURL))
	}
	if agent.SSHEnabled {
		lines = append(lines, "export HAIWO_AGENT_SSH_ENABLED=true")
	}
	lines = append(lines,
		"",
		"mkdir -p \"$HAIWO_AGENT_HOME\" \"$HAIWO_AGENT_WORKDIR\"",
		"",
		"stop_agent() {",
		"  if [ -f \"$HAIWO_AGENT_PID_FILE\" ]; then",
		"    old_pid=\"$(cat \"$HAIWO_AGENT_PID_FILE\" 2>/dev/null || true)\"",
		"    if [ -n \"$old_pid\" ] && kill -0 \"$old_pid\" >/dev/null 2>&1; then",
		"      echo \"Stopping existing Haiwo agent pid=${old_pid}\"",
		"      kill \"$old_pid\" >/dev/null 2>&1 || true",
		"      for _ in {1..20}; do",
		"        kill -0 \"$old_pid\" >/dev/null 2>&1 || break",
		"        sleep 0.2",
		"      done",
		"      kill -0 \"$old_pid\" >/dev/null 2>&1 && kill -9 \"$old_pid\" >/dev/null 2>&1 || true",
		"    fi",
		"    rm -f \"$HAIWO_AGENT_PID_FILE\"",
		"  fi",
		"  if command -v pgrep >/dev/null 2>&1; then",
		"    while IFS= read -r old_pid; do",
		"      [ -n \"$old_pid\" ] || continue",
		"      [ \"$old_pid\" = \"$$\" ] && continue",
		"      echo \"Stopping existing Haiwo agent process pid=${old_pid}\"",
		"      kill \"$old_pid\" >/dev/null 2>&1 || true",
		"    done <<EOF",
		"$(pgrep -f \"$HAIWO_AGENT_BIN\" || true)",
		"EOF",
		"  fi",
		"}",
		"",
		"download_agent() {",
		"  tmp_bin=\"${HAIWO_AGENT_BIN}.tmp\"",
		"  rm -f \"$tmp_bin\"",
		"  if command -v curl >/dev/null 2>&1; then",
		"    curl -fsSL \"$HAIWO_AGENT_BINARY_URL\" -o \"$tmp_bin\"",
		"  elif command -v wget >/dev/null 2>&1; then",
		"    wget -qO \"$tmp_bin\" \"$HAIWO_AGENT_BINARY_URL\"",
		"  else",
		"    echo \"curl or wget is required to download haiwo agent\" >&2",
		"    exit 1",
		"  fi",
		"  chmod +x \"$tmp_bin\"",
		"  mv \"$tmp_bin\" \"$HAIWO_AGENT_BIN\"",
		"}",
		"",
		"if [ -e \"$HAIWO_AGENT_BIN\" ]; then",
		"  stop_agent",
		"  rm -f \"$HAIWO_AGENT_BIN\"",
		"fi",
		"",
		"download_agent",
		"nohup \"$HAIWO_AGENT_BIN\" >>\"${HAIWO_AGENT_HOME}/agent.log\" 2>&1 &",
		"echo $! > \"$HAIWO_AGENT_PID_FILE\"",
		"echo \"Haiwo agent started pid=$(cat \"$HAIWO_AGENT_PID_FILE\")\"",
		"echo \"Log: ${HAIWO_AGENT_HOME}/agent.log\"",
	)
	return strings.Join(lines, "\n")
}

func (a *App) deployScriptURL(r *http.Request, agent domain.AgentInfo) string {
	return a.baseURL(r) + "/api/agents/" + agent.ID + "/deploy.sh?token=" + agent.Token
}

func (a *App) baseURL(r *http.Request) string {
	if baseURL := normalizeBaseURL(a.store.GetSettings().ServerBaseURL); baseURL != "" {
		return baseURL
	}
	scheme := "http"
	if r.TLS != nil || strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https") {
		scheme = "https"
	}
	host := r.Host
	if forwardedHost := r.Header.Get("X-Forwarded-Host"); forwardedHost != "" {
		host = forwardedHost
	}
	return scheme + "://" + host
}

func websocketURL(r *http.Request) string {
	scheme := "ws"
	if r.TLS != nil || strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https") {
		scheme = "wss"
	}
	host := r.Host
	if forwardedHost := r.Header.Get("X-Forwarded-Host"); forwardedHost != "" {
		host = forwardedHost
	}
	return scheme + "://" + host + "/rpc/agent/ws"
}

func agentRPCURL(baseURL string) string {
	baseURL = normalizeBaseURL(baseURL)
	if strings.HasPrefix(baseURL, "https://") {
		return "wss://" + strings.TrimPrefix(baseURL, "https://") + "/rpc/agent/ws"
	}
	return "ws://" + strings.TrimPrefix(baseURL, "http://") + "/rpc/agent/ws"
}

func safeFilename(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "agent"
	}
	var b strings.Builder
	for _, r := range value {
		if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '-' || r == '_' {
			b.WriteRune(r)
		}
	}
	if b.Len() == 0 {
		return "agent"
	}
	return b.String()
}

func normalizeBaseURL(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	value = strings.TrimRight(value, "/")
	if strings.HasPrefix(value, "ws://") {
		value = "http://" + strings.TrimPrefix(value, "ws://")
	}
	if strings.HasPrefix(value, "wss://") {
		value = "https://" + strings.TrimPrefix(value, "wss://")
	}
	if !strings.HasPrefix(value, "http://") && !strings.HasPrefix(value, "https://") {
		value = "https://" + value
	}
	for _, suffix := range []string{"/ssh/agent", "/rpc/agent/ws"} {
		if strings.HasSuffix(value, suffix) {
			value = strings.TrimSuffix(value, suffix)
		}
	}
	return strings.TrimRight(value, "/")
}

func reverseSSHURLFromBase(baseURL string) string {
	baseURL = normalizeBaseURL(baseURL)
	if baseURL == "" {
		return ""
	}
	if strings.HasPrefix(baseURL, "https://") {
		return "wss://" + strings.TrimPrefix(baseURL, "https://") + "/ssh/agent"
	}
	return "ws://" + strings.TrimPrefix(baseURL, "http://") + "/ssh/agent"
}

func baseURLFromReverseSSHURL(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if strings.HasPrefix(value, "wss://") {
		value = "https://" + strings.TrimPrefix(value, "wss://")
	}
	if strings.HasPrefix(value, "ws://") {
		value = "http://" + strings.TrimPrefix(value, "ws://")
	}
	return normalizeBaseURL(value)
}

type Store struct {
	mu        sync.RWMutex
	now       func() time.Time
	db        *gorm.DB
	projects  map[string]domain.Project
	pipelines map[string]domain.Pipeline
	triggers  map[string]domain.Trigger
	agents    map[string]domain.AgentInfo
	runs      map[string]domain.Run
	settings  domain.SystemSettings
}

func NewStore(now func() time.Time, db *gorm.DB) *Store {
	s := &Store{
		now:       now,
		db:        db,
		projects:  map[string]domain.Project{},
		pipelines: map[string]domain.Pipeline{},
		triggers:  map[string]domain.Trigger{},
		agents:    map[string]domain.AgentInfo{},
		runs:      map[string]domain.Run{},
		settings:  domain.SystemSettings{},
	}
	s.loadFromDB()
	return s
}

func (s *Store) SaveProject(v domain.Project) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.projects[v.ID] = v
	if s.db != nil {
		if err := s.db.Save(&dbmodel.Project{
			ID:            v.ID,
			Name:          v.Name,
			Provider:      v.Provider,
			RepoURL:       v.RepoURL,
			DefaultBranch: v.DefaultBranch,
			CreatedAt:     v.CreatedAt,
		}).Error; err != nil {
			log.Printf("save project failed: %v", err)
		}
	}
}

func (s *Store) SavePipeline(v domain.Pipeline) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pipelines[v.ID] = v
	if s.db != nil {
		definition, err := json.Marshal(v.Stages)
		if err != nil {
			log.Printf("marshal pipeline failed: %v", err)
			return
		}
		if err := s.db.Save(&dbmodel.Pipeline{
			ID:         v.ID,
			ProjectID:  v.ProjectID,
			Name:       v.Name,
			Definition: definition,
			CreatedAt:  v.CreatedAt,
		}).Error; err != nil {
			log.Printf("save pipeline failed: %v", err)
		}
	}
}

func (s *Store) SaveTrigger(v domain.Trigger) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.triggers[v.ID] = v
	if s.db != nil {
		if err := s.db.Save(&dbmodel.Trigger{
			ID:             v.ID,
			ProjectID:      v.ProjectID,
			PipelineID:     v.PipelineID,
			Type:           v.Type,
			BranchPattern:  v.BranchPattern,
			CommitPattern:  v.CommitPattern,
			TagPattern:     v.TagPattern,
			CommentPattern: v.CommentPattern,
			Cron:           v.Cron,
			CreatedAt:      v.CreatedAt,
		}).Error; err != nil {
			log.Printf("save trigger failed: %v", err)
		}
	}
}

func (s *Store) SaveAgent(v domain.AgentInfo) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.agents[v.ID] = v
	if s.db != nil {
		labels, err := json.Marshal(v.Labels)
		if err != nil {
			log.Printf("marshal agent labels failed: %v", err)
			return
		}
		if err := s.db.Save(&dbmodel.Agent{
			ID:            v.ID,
			Name:          v.Name,
			Token:         v.Token,
			Version:       v.Version,
			Labels:        labels,
			Status:        string(v.Status),
			CurrentRun:    v.CurrentRun,
			MaxRunning:    v.MaxRunning,
			SSHEnabled:    v.SSHEnabled,
			SSHHost:       v.SSHHost,
			SSHPort:       v.SSHPort,
			SSHUser:       v.SSHUser,
			ReverseSSHURL: v.ReverseSSHURL,
			CreatedAt:     v.CreatedAt,
			LastSeenAt:    v.LastSeenAt,
		}).Error; err != nil {
			log.Printf("save agent failed: %v", err)
		}
	}
}

func (s *Store) DeleteAgent(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.agents, id)
	if s.db != nil {
		if err := s.db.Delete(&dbmodel.Agent{ID: id}).Error; err != nil {
			log.Printf("delete agent failed: %v", err)
		}
	}
}

func (s *Store) AgentNameExists(name string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, agent := range s.agents {
		if strings.EqualFold(agent.Name, name) {
			return true
		}
	}
	return false
}

func (s *Store) SaveRun(v domain.Run) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.runs[v.ID] = v
	if s.db != nil {
		metadata, err := json.Marshal(v.Metadata)
		if err != nil {
			log.Printf("marshal run metadata failed: %v", err)
			return
		}
		if err := s.db.Save(&dbmodel.Run{
			ID:         v.ID,
			ProjectID:  v.ProjectID,
			PipelineID: v.PipelineID,
			Status:     string(v.Status),
			Source:     v.Source,
			Ref:        v.Ref,
			Comment:    v.Comment,
			CommitSHA:  v.CommitSHA,
			Metadata:   metadata,
			CreatedAt:  v.CreatedAt,
			UpdatedAt:  v.UpdatedAt,
		}).Error; err != nil {
			log.Printf("save run failed: %v", err)
		}
	}
}

func (s *Store) GetAgent(id string) (domain.AgentInfo, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.agents[id]
	return v, ok
}

func (s *Store) GetAgentByToken(token string) (domain.AgentInfo, bool) {
	if token == "" {
		return domain.AgentInfo{}, false
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, v := range s.agents {
		if v.Token != "" && v.Token == token {
			return v, true
		}
	}
	return domain.AgentInfo{}, false
}

func (s *Store) GetSettings() domain.SystemSettings {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.settings
}

func (s *Store) SaveSettings(v domain.SystemSettings) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.settings = v
	if s.db != nil {
		value, err := json.Marshal(v)
		if err != nil {
			log.Printf("marshal settings failed: %v", err)
			return
		}
		if err := s.db.Save(&dbmodel.SystemSetting{Key: "system", Value: value, UpdatedAt: s.now()}).Error; err != nil {
			log.Printf("save settings failed: %v", err)
		}
	}
}

func (s *Store) ListProjects() []domain.Project {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]domain.Project, 0, len(s.projects))
	for _, v := range s.projects {
		out = append(out, v)
	}
	return out
}

func (s *Store) GetPipeline(id string) (domain.Pipeline, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.pipelines[id]
	return v, ok
}
func (s *Store) ListPipelines(projectID string) []domain.Pipeline {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]domain.Pipeline, 0, len(s.pipelines))
	for _, v := range s.pipelines {
		if projectID == "" || v.ProjectID == projectID {
			out = append(out, v)
		}
	}
	return out
}
func (s *Store) GetRun(id string) (domain.Run, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.runs[id]
	return v, ok
}
func (s *Store) ListRuns(projectID string) []domain.Run {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]domain.Run, 0, len(s.runs))
	for _, v := range s.runs {
		if projectID == "" || v.ProjectID == projectID {
			out = append(out, v)
		}
	}
	return out
}
func (s *Store) ListTriggers(projectID string) []domain.Trigger {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []domain.Trigger
	for _, v := range s.triggers {
		if v.ProjectID == projectID {
			out = append(out, v)
		}
	}
	return out
}
func (s *Store) ListAgents() []domain.AgentInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]domain.AgentInfo, 0, len(s.agents))
	for _, v := range s.agents {
		out = append(out, v)
	}
	return out
}

func (s *Store) loadFromDB() {
	if s.db == nil {
		return
	}
	var projects []dbmodel.Project
	if err := s.db.Find(&projects).Error; err != nil {
		log.Printf("load projects failed: %v", err)
	} else {
		for _, v := range projects {
			s.projects[v.ID] = domain.Project{
				ID:            v.ID,
				Name:          v.Name,
				Provider:      v.Provider,
				RepoURL:       v.RepoURL,
				DefaultBranch: v.DefaultBranch,
				CreatedAt:     v.CreatedAt,
			}
		}
	}
	var pipelines []dbmodel.Pipeline
	if err := s.db.Find(&pipelines).Error; err != nil {
		log.Printf("load pipelines failed: %v", err)
	} else {
		for _, v := range pipelines {
			var stages []domain.Stage
			if len(v.Definition) > 0 {
				_ = json.Unmarshal(v.Definition, &stages)
			}
			s.pipelines[v.ID] = domain.Pipeline{
				ID:        v.ID,
				ProjectID: v.ProjectID,
				Name:      v.Name,
				Stages:    stages,
				CreatedAt: v.CreatedAt,
			}
		}
	}
	var triggers []dbmodel.Trigger
	if err := s.db.Find(&triggers).Error; err != nil {
		log.Printf("load triggers failed: %v", err)
	} else {
		for _, v := range triggers {
			s.triggers[v.ID] = domain.Trigger{
				ID:             v.ID,
				ProjectID:      v.ProjectID,
				PipelineID:     v.PipelineID,
				Type:           v.Type,
				BranchPattern:  v.BranchPattern,
				CommitPattern:  v.CommitPattern,
				TagPattern:     v.TagPattern,
				CommentPattern: v.CommentPattern,
				Cron:           v.Cron,
				CreatedAt:      v.CreatedAt,
			}
		}
	}
	var agents []dbmodel.Agent
	if err := s.db.Find(&agents).Error; err != nil {
		log.Printf("load agents failed: %v", err)
	} else {
		for _, v := range agents {
			var labels []string
			if len(v.Labels) > 0 {
				_ = json.Unmarshal(v.Labels, &labels)
			}
			s.agents[v.ID] = domain.AgentInfo{
				ID:            v.ID,
				Name:          v.Name,
				Token:         v.Token,
				Version:       v.Version,
				Labels:        labels,
				Status:        domain.AgentOffline,
				CurrentRun:    0,
				MaxRunning:    v.MaxRunning,
				SSHEnabled:    v.SSHEnabled,
				SSHHost:       v.SSHHost,
				SSHPort:       v.SSHPort,
				SSHUser:       v.SSHUser,
				ReverseSSHURL: v.ReverseSSHURL,
				CreatedAt:     v.CreatedAt,
				LastSeenAt:    v.LastSeenAt,
			}
		}
	}
	var runs []dbmodel.Run
	if err := s.db.Find(&runs).Error; err != nil {
		log.Printf("load runs failed: %v", err)
	} else {
		for _, v := range runs {
			var metadata map[string]string
			if len(v.Metadata) > 0 {
				_ = json.Unmarshal(v.Metadata, &metadata)
			}
			s.runs[v.ID] = domain.Run{
				ID:         v.ID,
				ProjectID:  v.ProjectID,
				PipelineID: v.PipelineID,
				Status:     domain.RunStatus(v.Status),
				Source:     v.Source,
				Ref:        v.Ref,
				Comment:    v.Comment,
				CommitSHA:  v.CommitSHA,
				Metadata:   metadata,
				CreatedAt:  v.CreatedAt,
				UpdatedAt:  v.UpdatedAt,
			}
		}
	}
	var setting dbmodel.SystemSetting
	if err := s.db.First(&setting, "key = ?", "system").Error; err == nil && len(setting.Value) > 0 {
		_ = json.Unmarshal(setting.Value, &s.settings)
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Printf("load settings failed: %v", err)
	}
}

type AgentSession struct {
	id          string
	authAgentID string
	token       string
	peer        *jsonrpc.Peer
}

type AgentHub struct {
	mu       sync.RWMutex
	store    *Store
	sessions map[string]*AgentSession
}

func NewAgentHub(store *Store) *AgentHub {
	return &AgentHub{store: store, sessions: map[string]*AgentSession{}}
}

func (h *AgentHub) RegisterHandler(session *AgentSession) jsonrpc.Handler {
	return func(_ context.Context, params json.RawMessage) (any, *jsonrpc.Error) {
		var info domain.AgentInfo
		if err := json.Unmarshal(params, &info); err != nil {
			return nil, &jsonrpc.Error{Code: -32602, Message: err.Error()}
		}
		if session.authAgentID != "" && info.ID != session.authAgentID {
			return nil, &jsonrpc.Error{Code: -32001, Message: "agent token does not match agent id"}
		}
		if existing, ok := h.store.GetAgent(info.ID); ok {
			info.Token = existing.Token
			info.SSHEnabled = existing.SSHEnabled
			info.SSHHost = existing.SSHHost
			info.SSHPort = existing.SSHPort
			info.SSHUser = existing.SSHUser
			info.ReverseSSHURL = existing.ReverseSSHURL
			info.CreatedAt = existing.CreatedAt
			if len(info.Labels) == 0 {
				info.Labels = existing.Labels
			}
			if info.MaxRunning <= 0 {
				info.MaxRunning = existing.MaxRunning
			}
		} else if session.token != "" {
			info.Token = session.token
		}
		if info.CreatedAt.IsZero() {
			info.CreatedAt = time.Now()
		}
		info.Status = domain.AgentOnline
		info.LastSeenAt = time.Now()
		session.id = info.ID
		h.mu.Lock()
		h.sessions[info.ID] = session
		h.mu.Unlock()
		h.store.SaveAgent(info)
		return map[string]bool{"accepted": true}, nil
	}
}

func (h *AgentHub) HeartbeatHandler(session *AgentSession) jsonrpc.Handler {
	return func(_ context.Context, params json.RawMessage) (any, *jsonrpc.Error) {
		var info domain.AgentInfo
		if err := json.Unmarshal(params, &info); err != nil {
			return nil, &jsonrpc.Error{Code: -32602, Message: err.Error()}
		}
		if info.ID == "" {
			info.ID = session.id
		}
		if session.authAgentID != "" && info.ID != session.authAgentID {
			return nil, &jsonrpc.Error{Code: -32001, Message: "agent token does not match agent id"}
		}
		if existing, ok := h.store.GetAgent(info.ID); ok {
			info.Token = existing.Token
			info.SSHEnabled = existing.SSHEnabled
			info.SSHHost = existing.SSHHost
			info.SSHPort = existing.SSHPort
			info.SSHUser = existing.SSHUser
			info.ReverseSSHURL = existing.ReverseSSHURL
			info.CreatedAt = existing.CreatedAt
			if len(info.Labels) == 0 {
				info.Labels = existing.Labels
			}
			if info.MaxRunning <= 0 {
				info.MaxRunning = existing.MaxRunning
			}
		}
		if info.CreatedAt.IsZero() {
			info.CreatedAt = time.Now()
		}
		info.Status = domain.AgentOnline
		info.LastSeenAt = time.Now()
		h.store.SaveAgent(info)
		return map[string]bool{"ok": true}, nil
	}
}

func (h *AgentHub) Disconnect(session *AgentSession) {
	if session == nil || session.id == "" {
		return
	}
	h.mu.Lock()
	delete(h.sessions, session.id)
	h.mu.Unlock()
}

func (h *AgentHub) Delete(agentID string) {
	h.mu.Lock()
	session := h.sessions[agentID]
	delete(h.sessions, agentID)
	h.mu.Unlock()
	if session != nil && session.peer != nil {
		_ = session.peer.Close()
	}
}

func (h *AgentHub) Match(ids, labels []string) []*AgentSession {
	h.mu.RLock()
	defer h.mu.RUnlock()
	var out []*AgentSession
	for id, session := range h.sessions {
		if len(ids) > 0 && !contains(ids, id) {
			continue
		}
		if len(labels) > 0 {
			agents := h.store.ListAgents()
			var info domain.AgentInfo
			for _, ag := range agents {
				if ag.ID == id {
					info = ag
					break
				}
			}
			if !hasAll(info.Labels, labels) {
				continue
			}
		}
		out = append(out, session)
	}
	return out
}

func decodeJSON(w http.ResponseWriter, r *http.Request, out any) bool {
	if err := json.NewDecoder(r.Body).Decode(out); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return false
	}
	return true
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func id(prefix string) string {
	var b [8]byte
	_, _ = rand.Read(b[:])
	return prefix + "_" + hex.EncodeToString(b[:])
}

func token() string {
	var b [24]byte
	_, _ = rand.Read(b[:])
	return "hwagt_" + hex.EncodeToString(b[:])
}

func cleanStrings(values []string) []string {
	out := make([]string, 0, len(values))
	seen := map[string]bool{}
	for _, value := range values {
		for _, part := range strings.Split(value, ",") {
			part = strings.TrimSpace(part)
			if part != "" && !seen[part] {
				out = append(out, part)
				seen[part] = true
			}
		}
	}
	return out
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
}

func contains(values []string, want string) bool {
	for _, v := range values {
		if v == want {
			return true
		}
	}
	return false
}

func hasAll(values, wants []string) bool {
	for _, want := range wants {
		if !contains(values, want) {
			return false
		}
	}
	return true
}
