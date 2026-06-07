package database

import "time"

type Project struct {
	ID            string    `gorm:"type:varchar(64);primaryKey"`
	Name          string    `gorm:"type:varchar(255);not null"`
	Provider      string    `gorm:"type:varchar(32);not null;index"`
	RepoURL       string    `gorm:"type:text;not null"`
	DefaultBranch string    `gorm:"type:varchar(255);not null;default:'main'"`
	CreatedAt     time.Time `gorm:"not null"`
}

type Agent struct {
	ID            string    `gorm:"type:varchar(128);primaryKey"`
	Name          string    `gorm:"type:varchar(255);not null"`
	Token         string    `gorm:"type:varchar(128);index"`
	Version       string    `gorm:"type:varchar(255);not null"`
	Labels        []byte    `gorm:"type:jsonb;not null;default:'[]'"`
	Status        string    `gorm:"type:varchar(32);not null;index"`
	CurrentRun    int       `gorm:"not null;default:0"`
	MaxRunning    int       `gorm:"not null;default:1"`
	SSHEnabled    bool      `gorm:"not null;default:false"`
	SSHHost       string    `gorm:"type:varchar(255)"`
	SSHPort       int       `gorm:"not null;default:22"`
	SSHUser       string    `gorm:"type:varchar(255)"`
	ReverseSSHURL string    `gorm:"type:text"`
	CreatedAt     time.Time `gorm:"not null;index"`
	LastSeenAt    time.Time `gorm:"not null;index"`
}

type Pipeline struct {
	ID         string    `gorm:"type:varchar(64);primaryKey"`
	ProjectID  string    `gorm:"type:varchar(64);not null;index"`
	Name       string    `gorm:"type:varchar(255);not null"`
	Definition []byte    `gorm:"type:jsonb;not null"`
	CreatedAt  time.Time `gorm:"not null"`
}

type Trigger struct {
	ID             string    `gorm:"type:varchar(64);primaryKey"`
	ProjectID      string    `gorm:"type:varchar(64);not null;index"`
	PipelineID     string    `gorm:"type:varchar(64);not null;index"`
	Type           string    `gorm:"type:varchar(32);not null;index"`
	BranchPattern  string    `gorm:"type:varchar(255)"`
	CommitPattern  string    `gorm:"type:varchar(255)"`
	TagPattern     string    `gorm:"type:varchar(255)"`
	CommentPattern string    `gorm:"type:varchar(255)"`
	Cron           string    `gorm:"type:varchar(255)"`
	CreatedAt      time.Time `gorm:"not null"`
}

type Run struct {
	ID         string    `gorm:"type:varchar(64);primaryKey"`
	ProjectID  string    `gorm:"type:varchar(64);not null;index"`
	PipelineID string    `gorm:"type:varchar(64);index"`
	Status     string    `gorm:"type:varchar(32);not null;index"`
	Source     string    `gorm:"type:varchar(32);not null;index"`
	Ref        string    `gorm:"type:varchar(255)"`
	Comment    string    `gorm:"type:text"`
	CommitSHA  string    `gorm:"type:varchar(128)"`
	Metadata   []byte    `gorm:"type:jsonb;not null;default:'{}'"`
	CreatedAt  time.Time `gorm:"not null;index"`
	UpdatedAt  time.Time `gorm:"not null;index"`
}

type JobRun struct {
	ID         string     `gorm:"type:varchar(64);primaryKey"`
	RunID      string     `gorm:"type:varchar(64);not null;index"`
	JobID      string     `gorm:"type:varchar(128);not null;index"`
	AgentID    string     `gorm:"type:varchar(128);index"`
	Status     string     `gorm:"type:varchar(32);not null;index"`
	StartedAt  *time.Time `gorm:"index"`
	FinishedAt *time.Time `gorm:"index"`
	ExitCode   *int
	Error      string `gorm:"type:text"`
}

type TaskLog struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement"`
	RunID     string    `gorm:"type:varchar(64);index"`
	JobID     string    `gorm:"type:varchar(128);index"`
	TaskID    string    `gorm:"type:varchar(64);not null;index"`
	Stream    string    `gorm:"type:varchar(16);not null"`
	Line      string    `gorm:"type:text;not null"`
	CreatedAt time.Time `gorm:"not null;index"`
}

type SystemSetting struct {
	Key       string    `gorm:"type:varchar(128);primaryKey"`
	Value     []byte    `gorm:"type:jsonb;not null;default:'{}'"`
	UpdatedAt time.Time `gorm:"not null"`
}
