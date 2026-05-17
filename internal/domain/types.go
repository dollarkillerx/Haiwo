package domain

import "time"

type AgentStatus string

const (
	AgentOnline   AgentStatus = "online"
	AgentOffline  AgentStatus = "offline"
	AgentBusy     AgentStatus = "busy"
	AgentDisabled AgentStatus = "disabled"
)

type RunStatus string

const (
	RunQueued          RunStatus = "queued"
	RunRunning         RunStatus = "running"
	RunSuccess         RunStatus = "success"
	RunFailed          RunStatus = "failed"
	RunCanceled        RunStatus = "canceled"
	RunRollbackRunning RunStatus = "rollback_running"
	RunRollbackSuccess RunStatus = "rollback_success"
	RunRollbackFailed  RunStatus = "rollback_failed"
)

type JobStatus string

const (
	JobQueued   JobStatus = "queued"
	JobAssigned JobStatus = "assigned"
	JobRunning  JobStatus = "running"
	JobSuccess  JobStatus = "success"
	JobFailed   JobStatus = "failed"
	JobSkipped  JobStatus = "skipped"
	JobTimeout  JobStatus = "timeout"
)

type JobType string

const (
	JobCommand      JobType = "command"
	JobGitCheckout  JobType = "git_checkout"
	JobDBBackup     JobType = "db_backup"
	JobDBRestore    JobType = "db_restore"
	JobRollbackCode JobType = "rollback_code"
	JobRollbackDB   JobType = "rollback_db"
)

type AgentMode string

const (
	AgentModeSingle   AgentMode = "single"
	AgentModeParallel AgentMode = "parallel"
	AgentModeChain    AgentMode = "chain"
)

type DatabaseType string

const (
	DatabaseMySQL      DatabaseType = "mysql"
	DatabasePostgreSQL DatabaseType = "postgresql"
)

type Project struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Provider      string    `json:"provider"`
	RepoURL       string    `json:"repo_url"`
	DefaultBranch string    `json:"default_branch"`
	CreatedAt     time.Time `json:"created_at"`
}

type Trigger struct {
	ID             string    `json:"id"`
	ProjectID      string    `json:"project_id"`
	PipelineID     string    `json:"pipeline_id"`
	Type           string    `json:"type"`
	BranchPattern  string    `json:"branch_pattern,omitempty"`
	CommentPattern string    `json:"comment_pattern,omitempty"`
	Cron           string    `json:"cron,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}

type Pipeline struct {
	ID        string    `json:"id"`
	ProjectID string    `json:"project_id"`
	Name      string    `json:"name"`
	Stages    []Stage   `json:"stages"`
	CreatedAt time.Time `json:"created_at"`
}

type Stage struct {
	Name string `json:"name"`
	Jobs []Job  `json:"jobs"`
}

type Job struct {
	ID             string            `json:"id"`
	Name           string            `json:"name"`
	Type           JobType           `json:"job_type"`
	AgentMode      AgentMode         `json:"agent_mode"`
	AgentIDs       []string          `json:"agent_ids,omitempty"`
	AgentLabels    []string          `json:"agent_labels,omitempty"`
	WorkingDir     string            `json:"working_dir,omitempty"`
	Commands       []string          `json:"commands,omitempty"`
	Repo           *GitRef           `json:"repo,omitempty"`
	Env            map[string]string `json:"env,omitempty"`
	Secrets        map[string]string `json:"secrets,omitempty"`
	TimeoutSeconds int               `json:"timeout_seconds,omitempty"`
	DatabaseTarget *DatabaseTarget   `json:"database_target,omitempty"`
	BackupID       string            `json:"backup_id,omitempty"`
	Required       bool              `json:"required"`
}

type GitRef struct {
	URL    string `json:"url"`
	Ref    string `json:"ref"`
	Commit string `json:"commit,omitempty"`
}

type DatabaseTarget struct {
	ID                 string            `json:"id,omitempty"`
	Type               DatabaseType      `json:"type"`
	Host               string            `json:"host"`
	Port               int               `json:"port"`
	Database           string            `json:"database"`
	Username           string            `json:"username"`
	Password           string            `json:"password,omitempty"`
	PasswordSecretID   string            `json:"password_secret_id,omitempty"`
	ExtraOptions       map[string]string `json:"extra_options,omitempty"`
	Environment        string            `json:"environment"`
	AllowedAgentLabels []string          `json:"allowed_agent_labels,omitempty"`
	BackupFormat       string            `json:"backup_format,omitempty"`
	Confirm            bool              `json:"confirm,omitempty"`
}

type AgentInfo struct {
	ID         string      `json:"id"`
	Name       string      `json:"name"`
	Version    string      `json:"version"`
	Labels     []string    `json:"labels"`
	Status     AgentStatus `json:"status"`
	CurrentRun int         `json:"current_run"`
	MaxRunning int         `json:"max_running"`
	LastSeenAt time.Time   `json:"last_seen_at"`
}

type TaskPayload struct {
	TaskID         string            `json:"task_id"`
	RunID          string            `json:"run_id"`
	JobID          string            `json:"job_id"`
	JobType        JobType           `json:"job_type"`
	WorkingDir     string            `json:"working_dir,omitempty"`
	Commands       []string          `json:"commands,omitempty"`
	Repo           *GitRef           `json:"repo,omitempty"`
	Ref            string            `json:"ref,omitempty"`
	Env            map[string]string `json:"env,omitempty"`
	Secrets        map[string]string `json:"secrets,omitempty"`
	TimeoutSeconds int               `json:"timeout_seconds,omitempty"`
	DatabaseTarget *DatabaseTarget   `json:"database_target,omitempty"`
	BackupID       string            `json:"backup_id,omitempty"`
}

type TaskLog struct {
	TaskID string    `json:"task_id"`
	Stream string    `json:"stream"`
	Line   string    `json:"line"`
	Time   time.Time `json:"time"`
}

type TaskProgress struct {
	TaskID  string `json:"task_id"`
	Message string `json:"message"`
	Percent int    `json:"percent"`
}

type TaskComplete struct {
	TaskID    string            `json:"task_id"`
	RunID     string            `json:"run_id"`
	JobID     string            `json:"job_id"`
	Status    JobStatus         `json:"status"`
	ExitCode  int               `json:"exit_code"`
	Error     string            `json:"error,omitempty"`
	Artifacts map[string]string `json:"artifacts,omitempty"`
	BackupID  string            `json:"backup_id,omitempty"`
}

type BackupRecord struct {
	ID          string       `json:"id"`
	ProjectID   string       `json:"project_id"`
	Environment string       `json:"environment"`
	Type        DatabaseType `json:"type"`
	Database    string       `json:"database"`
	AgentID     string       `json:"agent_id"`
	RunID       string       `json:"run_id"`
	JobID       string       `json:"job_id"`
	Path        string       `json:"path"`
	Size        int64        `json:"size"`
	Checksum    string       `json:"checksum"`
	CreatedAt   time.Time    `json:"created_at"`
}

type Run struct {
	ID         string            `json:"id"`
	ProjectID  string            `json:"project_id"`
	PipelineID string            `json:"pipeline_id"`
	Status     RunStatus         `json:"status"`
	Source     string            `json:"source"`
	Ref        string            `json:"ref,omitempty"`
	Comment    string            `json:"comment,omitempty"`
	CommitSHA  string            `json:"commit_sha,omitempty"`
	Metadata   map[string]string `json:"metadata,omitempty"`
	CreatedAt  time.Time         `json:"created_at"`
	UpdatedAt  time.Time         `json:"updated_at"`
}
