package server

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/haiwo-ci/haiwo/internal/domain"
	"github.com/haiwo-ci/haiwo/internal/jsonrpc"
	"github.com/robfig/cron/v3"
)

type Config struct {
	Addr        string
	AgentToken  string
	WebPassword string
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
		store: NewStore(cfg.Now),
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
	mux.HandleFunc("GET /rpc/agent/ws", a.handleAgentWS)
	mux.HandleFunc("GET /api/projects", a.listProjects)
	mux.HandleFunc("POST /api/projects", a.createProject)
	mux.HandleFunc("GET /api/pipelines", a.listPipelines)
	mux.HandleFunc("POST /api/projects/{id}/pipelines", a.createPipeline)
	mux.HandleFunc("POST /api/projects/{id}/triggers", a.createTrigger)
	mux.HandleFunc("POST /api/projects/{id}/databases", a.createDatabase)
	mux.HandleFunc("POST /api/webhooks/{provider}/{project_id}", a.handleWebhook)
	mux.HandleFunc("GET /api/runs", a.listRuns)
	mux.HandleFunc("GET /api/runs/{id}", a.getRun)
	mux.HandleFunc("POST /api/pipelines/{id}/runs", a.startPipelineRun)
	mux.HandleFunc("POST /api/runs/{id}/cancel", a.cancelRun)
	mux.HandleFunc("POST /api/runs/{id}/rollback", a.rollbackRun)
	mux.HandleFunc("GET /api/backups", a.listBackups)
	mux.HandleFunc("POST /api/backups/{id}/restore", a.restoreBackup)
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
	if a.cfg.AgentToken != "" && token != a.cfg.AgentToken {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	conn, err := a.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	peer := jsonrpc.NewPeer(conn)
	session := &AgentSession{peer: peer}
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

func (a *App) createDatabase(w http.ResponseWriter, r *http.Request) {
	var db domain.DatabaseTarget
	if !decodeJSON(w, r, &db) {
		return
	}
	db.ID = id("db")
	a.store.SaveDatabase(r.PathValue("id"), db)
	db.Password = ""
	writeJSON(w, http.StatusCreated, db)
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
		BackupID   string `json:"backup_id"`
		Confirm    bool   `json:"confirm"`
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
	run.Metadata = map[string]string{"source_run_id": r.PathValue("id"), "backup_id": req.BackupID}
	run.Status = domain.RunRollbackRunning
	a.store.SaveRun(run)
	writeJSON(w, http.StatusAccepted, run)
}

func (a *App) listBackups(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, a.store.ListBackups())
}

func (a *App) restoreBackup(w http.ResponseWriter, r *http.Request) {
	backup, ok := a.store.GetBackup(r.PathValue("id"))
	if !ok {
		http.NotFound(w, r)
		return
	}
	var req struct {
		DatabaseTarget domain.DatabaseTarget `json:"database_target"`
		AgentLabels    []string              `json:"agent_labels"`
		Confirm        bool                  `json:"confirm"`
	}
	if !decodeJSON(w, r, &req) {
		return
	}
	if req.DatabaseTarget.Environment == "prod" && !req.Confirm {
		http.Error(w, "production restore requires confirm=true", http.StatusBadRequest)
		return
	}
	job := domain.Job{
		ID:             id("job"),
		Type:           domain.JobDBRestore,
		AgentMode:      domain.AgentModeSingle,
		AgentLabels:    req.AgentLabels,
		DatabaseTarget: &req.DatabaseTarget,
		BackupID:       backup.ID,
		Required:       true,
	}
	run := domain.Run{ID: id("run"), ProjectID: backup.ProjectID, Status: domain.RunRollbackRunning, Source: "restore", CreatedAt: a.cfg.Now(), UpdatedAt: a.cfg.Now()}
	a.store.SaveRun(run)
	go a.runJob(context.Background(), run, job)
	writeJSON(w, http.StatusAccepted, run)
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
		DatabaseTarget: job.DatabaseTarget,
		BackupID:       job.BackupID,
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
	if msg.BackupID != "" && msg.Artifacts != nil {
		a.store.SaveBackup(domain.BackupRecord{
			ID:        msg.BackupID,
			RunID:     msg.RunID,
			JobID:     msg.JobID,
			Path:      msg.Artifacts["path"],
			Checksum:  msg.Artifacts["checksum"],
			CreatedAt: a.cfg.Now(),
		})
	}
	return map[string]bool{"ok": true}, nil
}

func (a *App) artifactUploadComplete(_ context.Context, _ json.RawMessage) (any, *jsonrpc.Error) {
	return map[string]bool{"ok": true}, nil
}

type Store struct {
	mu        sync.RWMutex
	now       func() time.Time
	projects  map[string]domain.Project
	pipelines map[string]domain.Pipeline
	triggers  map[string]domain.Trigger
	databases map[string][]domain.DatabaseTarget
	agents    map[string]domain.AgentInfo
	runs      map[string]domain.Run
	backups   map[string]domain.BackupRecord
}

func NewStore(now func() time.Time) *Store {
	return &Store{
		now:       now,
		projects:  map[string]domain.Project{},
		pipelines: map[string]domain.Pipeline{},
		triggers:  map[string]domain.Trigger{},
		databases: map[string][]domain.DatabaseTarget{},
		agents:    map[string]domain.AgentInfo{},
		runs:      map[string]domain.Run{},
		backups:   map[string]domain.BackupRecord{},
	}
}

func (s *Store) SaveProject(v domain.Project) { s.mu.Lock(); defer s.mu.Unlock(); s.projects[v.ID] = v }
func (s *Store) SavePipeline(v domain.Pipeline) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pipelines[v.ID] = v
}
func (s *Store) SaveTrigger(v domain.Trigger) { s.mu.Lock(); defer s.mu.Unlock(); s.triggers[v.ID] = v }
func (s *Store) SaveDatabase(projectID string, v domain.DatabaseTarget) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.databases[projectID] = append(s.databases[projectID], v)
}
func (s *Store) SaveAgent(v domain.AgentInfo) { s.mu.Lock(); defer s.mu.Unlock(); s.agents[v.ID] = v }
func (s *Store) SaveRun(v domain.Run)         { s.mu.Lock(); defer s.mu.Unlock(); s.runs[v.ID] = v }
func (s *Store) SaveBackup(v domain.BackupRecord) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.backups[v.ID] = v
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
func (s *Store) GetBackup(id string) (domain.BackupRecord, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.backups[id]
	return v, ok
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
func (s *Store) ListBackups() []domain.BackupRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]domain.BackupRecord, 0, len(s.backups))
	for _, v := range s.backups {
		out = append(out, v)
	}
	return out
}

type AgentSession struct {
	id   string
	peer *jsonrpc.Peer
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
