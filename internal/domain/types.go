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
	JobRollbackCode JobType = "rollback_code"
)

type AgentMode string

const (
	AgentModeSingle   AgentMode = "single"
	AgentModeParallel AgentMode = "parallel"
	AgentModeChain    AgentMode = "chain"
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
	CommitPattern  string    `json:"commit_pattern,omitempty"`
	TagPattern     string    `json:"tag_pattern,omitempty"`
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
	Required       bool              `json:"required"`
}

type GitRef struct {
	URL    string `json:"url"`
	Ref    string `json:"ref"`
	Commit string `json:"commit,omitempty"`
}

type AgentInfo struct {
	ID            string      `json:"id"`
	Name          string      `json:"name"`
	Token         string      `json:"token,omitempty"`
	Version       string      `json:"version"`
	Labels        []string    `json:"labels"`
	Status        AgentStatus `json:"status"`
	CurrentRun    int         `json:"current_run"`
	MaxRunning    int         `json:"max_running"`
	SSHEnabled    bool        `json:"ssh_enabled"`
	SSHHost       string      `json:"ssh_host,omitempty"`
	SSHPort       int         `json:"ssh_port,omitempty"`
	SSHUser       string      `json:"ssh_user,omitempty"`
	ReverseSSHURL string      `json:"reverse_ssh_url,omitempty"`
	CreatedAt     time.Time   `json:"created_at"`
	LastSeenAt    time.Time   `json:"last_seen_at"`
}

type SystemSettings struct {
	ServerBaseURL string `json:"server_base_url"`
	ReverseSSHURL string `json:"reverse_ssh_url"`
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
