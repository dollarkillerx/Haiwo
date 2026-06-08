package server

import (
	"encoding/json"
	"errors"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	dbmodel "github.com/haiwo-ci/haiwo/internal/database"
	"github.com/haiwo-ci/haiwo/internal/domain"
	"gorm.io/gorm"
)

// persistState is the full, serializable snapshot of the control-plane state the
// store keeps durable. Run logs are NOT part of this snapshot for the file
// backend — they stream to a separate append-only log file (see logFile) so that
// high-frequency log writes never trigger a full state rewrite.
type persistState struct {
	Projects  []domain.Project      `json:"projects"`
	Pipelines []domain.Pipeline     `json:"pipelines"`
	Triggers  []domain.Trigger      `json:"triggers"`
	Agents    []domain.AgentInfo    `json:"agents"`
	Runs      []domain.Run          `json:"runs"`
	Settings  domain.SystemSettings `json:"settings"`
}

// Persister is the durable backend behind the in-memory Store. The store stays
// the source of truth for reads; every mutation is written through to the
// Persister so it survives a restart. Two implementations exist: a file backend
// (default) and a PostgreSQL backend.
type Persister interface {
	Load() (persistState, error)
	SaveProject(domain.Project) error
	SavePipeline(domain.Pipeline) error
	SaveTrigger(domain.Trigger) error
	DeleteTriggersForPipeline(pipelineID string) error
	SaveAgent(domain.AgentInfo) error
	DeleteAgent(id string) error
	SaveRun(domain.Run) error
	// AppendRunLog persists a single run log line. It is on a hot path (one call
	// per streamed line), so implementations must be cheap and must not rewrite
	// the whole state.
	AppendRunLog(run domain.Run, entry string) error
	SaveSettings(domain.SystemSettings) error
}

// logFile is an append-only writer with size-based rotation. When the active
// file would exceed maxBytes it is rotated to "<path>.1" (deleting any previous
// ".1"), keeping the most recent segment plus one older segment.
type logFile struct {
	mu       sync.Mutex
	path     string
	maxBytes int64
	f        *os.File
	size     int64
}

func newLogFile(path string, maxBytes int64) (*logFile, error) {
	if dir := filepath.Dir(path); dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, err
		}
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return nil, err
	}
	info, err := f.Stat()
	if err != nil {
		_ = f.Close()
		return nil, err
	}
	return &logFile{path: path, maxBytes: maxBytes, f: f, size: info.Size()}, nil
}

func (l *logFile) append(line string) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	data := []byte(line + "\n")
	if l.maxBytes > 0 && l.size+int64(len(data)) > l.maxBytes {
		if err := l.rotate(); err != nil {
			return err
		}
	}
	n, err := l.f.Write(data)
	l.size += int64(n)
	return err
}

// rotate must be called with l.mu held.
func (l *logFile) rotate() error {
	if err := l.f.Close(); err != nil {
		return err
	}
	old := l.path + ".1"
	_ = os.Remove(old) // delete the previous old segment
	if err := os.Rename(l.path, old); err != nil {
		return err
	}
	f, err := os.OpenFile(l.path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	l.f = f
	l.size = 0
	return nil
}

func (l *logFile) close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.f.Close()
}

// filePersister stores the control-plane state as a single JSON file (rewritten
// atomically on each state change) and streams run logs to a separate log file.
// The snapshot func returns the store's current state; the store always holds
// its lock while a Save*/Delete* method runs, so reading the maps without an
// additional lock is safe.
type filePersister struct {
	path     string
	snapshot func() persistState
	mu       sync.Mutex
	logs     *logFile
}

func newFilePersister(path, logPath string, logMaxBytes int64, snapshot func() persistState) (*filePersister, error) {
	if dir := filepath.Dir(path); dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, err
		}
	}
	fp := &filePersister{path: path, snapshot: snapshot}
	if logPath != "" {
		lf, err := newLogFile(logPath, logMaxBytes)
		if err != nil {
			return nil, err
		}
		fp.logs = lf
	}
	return fp, nil
}

func (f *filePersister) Load() (persistState, error) {
	var state persistState
	data, err := os.ReadFile(f.path)
	if errors.Is(err, os.ErrNotExist) {
		return state, nil
	}
	if err != nil {
		return state, err
	}
	if len(data) == 0 {
		return state, nil
	}
	if err := json.Unmarshal(data, &state); err != nil {
		// Preserve the unreadable file instead of letting the next write
		// overwrite it, so the data can be recovered manually.
		if renameErr := os.Rename(f.path, f.path+".corrupt"); renameErr != nil {
			log.Printf("failed to preserve corrupt state file: %v", renameErr)
		} else {
			log.Printf("state file was corrupt; moved to %s.corrupt", f.path)
		}
		return persistState{}, err
	}
	return state, nil
}

func (f *filePersister) flush() error {
	state := f.snapshot()
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	tmp := f.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, f.path)
}

// State changes rewrite the (small, control-plane) snapshot. These are
// low-frequency: project/pipeline/trigger/agent CRUD and run status changes.
func (f *filePersister) SaveProject(domain.Project) error         { return f.flush() }
func (f *filePersister) SavePipeline(domain.Pipeline) error       { return f.flush() }
func (f *filePersister) SaveTrigger(domain.Trigger) error         { return f.flush() }
func (f *filePersister) DeleteTriggersForPipeline(string) error   { return f.flush() }
func (f *filePersister) SaveAgent(domain.AgentInfo) error         { return f.flush() }
func (f *filePersister) DeleteAgent(string) error                 { return f.flush() }
func (f *filePersister) SaveRun(domain.Run) error                 { return f.flush() }
func (f *filePersister) SaveSettings(domain.SystemSettings) error { return f.flush() }

// AppendRunLog streams the line to the separate log file (cheap append) instead
// of rewriting the whole state.
func (f *filePersister) AppendRunLog(run domain.Run, entry string) error {
	if f.logs == nil {
		return nil
	}
	return f.logs.append("[" + run.ID + "] " + entry)
}

func (f *filePersister) Close() error {
	if f.logs == nil {
		return nil
	}
	return f.logs.close()
}

// postgresPersister writes each entity through to PostgreSQL via GORM. Targeted
// upserts/deletes keep per-change cost low.
type postgresPersister struct {
	db  *gorm.DB
	now func() time.Time
}

func newPostgresPersister(db *gorm.DB, now func() time.Time) *postgresPersister {
	return &postgresPersister{db: db, now: now}
}

func (p *postgresPersister) SaveProject(v domain.Project) error {
	return p.db.Save(&dbmodel.Project{
		ID:            v.ID,
		Name:          v.Name,
		Provider:      v.Provider,
		RepoURL:       v.RepoURL,
		DefaultBranch: v.DefaultBranch,
		CreatedAt:     v.CreatedAt,
	}).Error
}

func (p *postgresPersister) SavePipeline(v domain.Pipeline) error {
	definition, err := json.Marshal(v.Stages)
	if err != nil {
		return err
	}
	return p.db.Save(&dbmodel.Pipeline{
		ID:         v.ID,
		ProjectID:  v.ProjectID,
		Name:       v.Name,
		Definition: definition,
		CreatedAt:  v.CreatedAt,
	}).Error
}

func (p *postgresPersister) SaveTrigger(v domain.Trigger) error {
	return p.db.Save(&dbmodel.Trigger{
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
	}).Error
}

func (p *postgresPersister) DeleteTriggersForPipeline(pipelineID string) error {
	return p.db.Delete(&dbmodel.Trigger{}, "pipeline_id = ?", pipelineID).Error
}

func (p *postgresPersister) SaveAgent(v domain.AgentInfo) error {
	labels, err := json.Marshal(v.Labels)
	if err != nil {
		return err
	}
	return p.db.Save(&dbmodel.Agent{
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
	}).Error
}

func (p *postgresPersister) DeleteAgent(id string) error {
	return p.db.Delete(&dbmodel.Agent{ID: id}).Error
}

func (p *postgresPersister) SaveRun(v domain.Run) error {
	metadata, err := json.Marshal(v.Metadata)
	if err != nil {
		return err
	}
	return p.db.Save(&dbmodel.Run{
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
	}).Error
}

// AppendRunLog persists the updated run. For Postgres a per-line row upsert is
// cheap, so logs stay incrementally durable (run.Metadata carries the capped
// log buffer).
func (p *postgresPersister) AppendRunLog(run domain.Run, _ string) error {
	return p.SaveRun(run)
}

func (p *postgresPersister) SaveSettings(v domain.SystemSettings) error {
	value, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return p.db.Save(&dbmodel.SystemSetting{Key: "system", Value: value, UpdatedAt: p.now()}).Error
}

// Load is best-effort: a failure loading one entity type is logged and the rest
// are still loaded, so a single bad table does not start the server empty.
func (p *postgresPersister) Load() (persistState, error) {
	var state persistState

	var projects []dbmodel.Project
	if err := p.db.Find(&projects).Error; err != nil {
		log.Printf("load projects failed: %v", err)
	} else {
		for _, v := range projects {
			state.Projects = append(state.Projects, domain.Project{
				ID:            v.ID,
				Name:          v.Name,
				Provider:      v.Provider,
				RepoURL:       v.RepoURL,
				DefaultBranch: v.DefaultBranch,
				CreatedAt:     v.CreatedAt,
			})
		}
	}

	var pipelines []dbmodel.Pipeline
	if err := p.db.Find(&pipelines).Error; err != nil {
		log.Printf("load pipelines failed: %v", err)
	} else {
		for _, v := range pipelines {
			var stages []domain.Stage
			if len(v.Definition) > 0 {
				_ = json.Unmarshal(v.Definition, &stages)
			}
			state.Pipelines = append(state.Pipelines, domain.Pipeline{
				ID:        v.ID,
				ProjectID: v.ProjectID,
				Name:      v.Name,
				Stages:    stages,
				CreatedAt: v.CreatedAt,
			})
		}
	}

	var triggers []dbmodel.Trigger
	if err := p.db.Find(&triggers).Error; err != nil {
		log.Printf("load triggers failed: %v", err)
	} else {
		for _, v := range triggers {
			state.Triggers = append(state.Triggers, domain.Trigger{
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
			})
		}
	}

	var agents []dbmodel.Agent
	if err := p.db.Find(&agents).Error; err != nil {
		log.Printf("load agents failed: %v", err)
	} else {
		for _, v := range agents {
			var labels []string
			if len(v.Labels) > 0 {
				_ = json.Unmarshal(v.Labels, &labels)
			}
			state.Agents = append(state.Agents, domain.AgentInfo{
				ID:            v.ID,
				Name:          v.Name,
				Token:         v.Token,
				Version:       v.Version,
				Labels:        labels,
				Status:        domain.AgentStatus(v.Status),
				CurrentRun:    v.CurrentRun,
				MaxRunning:    v.MaxRunning,
				SSHEnabled:    v.SSHEnabled,
				SSHHost:       v.SSHHost,
				SSHPort:       v.SSHPort,
				SSHUser:       v.SSHUser,
				ReverseSSHURL: v.ReverseSSHURL,
				CreatedAt:     v.CreatedAt,
				LastSeenAt:    v.LastSeenAt,
			})
		}
	}

	var runs []dbmodel.Run
	if err := p.db.Find(&runs).Error; err != nil {
		log.Printf("load runs failed: %v", err)
	} else {
		for _, v := range runs {
			var metadata map[string]string
			if len(v.Metadata) > 0 {
				_ = json.Unmarshal(v.Metadata, &metadata)
			}
			state.Runs = append(state.Runs, domain.Run{
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
			})
		}
	}

	var setting dbmodel.SystemSetting
	if err := p.db.First(&setting, "key = ?", "system").Error; err == nil && len(setting.Value) > 0 {
		_ = json.Unmarshal(setting.Value, &state.Settings)
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Printf("load settings failed: %v", err)
	}

	return state, nil
}
