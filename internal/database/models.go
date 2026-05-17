package database

import (
	"time"

	"gorm.io/datatypes"
)

type Project struct {
	ID            string    `gorm:"type:varchar(64);primaryKey"`
	Name          string    `gorm:"type:varchar(255);not null"`
	Provider      string    `gorm:"type:varchar(32);not null;index"`
	RepoURL       string    `gorm:"type:text;not null"`
	DefaultBranch string    `gorm:"type:varchar(255);not null;default:'main'"`
	CreatedAt     time.Time `gorm:"not null"`
}

type Agent struct {
	ID         string         `gorm:"type:varchar(128);primaryKey"`
	Name       string         `gorm:"type:varchar(255);not null"`
	Version    string         `gorm:"type:varchar(255);not null"`
	Labels     datatypes.JSON `gorm:"type:jsonb;not null;default:'[]'"`
	Status     string         `gorm:"type:varchar(32);not null;index"`
	CurrentRun int            `gorm:"not null;default:0"`
	MaxRunning int            `gorm:"not null;default:1"`
	LastSeenAt time.Time      `gorm:"not null;index"`
}

type Pipeline struct {
	ID         string         `gorm:"type:varchar(64);primaryKey"`
	ProjectID  string         `gorm:"type:varchar(64);not null;index"`
	Name       string         `gorm:"type:varchar(255);not null"`
	Definition datatypes.JSON `gorm:"type:jsonb;not null"`
	CreatedAt  time.Time      `gorm:"not null"`
}

type Trigger struct {
	ID             string    `gorm:"type:varchar(64);primaryKey"`
	ProjectID      string    `gorm:"type:varchar(64);not null;index"`
	PipelineID     string    `gorm:"type:varchar(64);not null;index"`
	Type           string    `gorm:"type:varchar(32);not null;index"`
	BranchPattern  string    `gorm:"type:varchar(255)"`
	CommentPattern string    `gorm:"type:varchar(255)"`
	Cron           string    `gorm:"type:varchar(255)"`
	CreatedAt      time.Time `gorm:"not null"`
}

type DatabaseTarget struct {
	ID                 string         `gorm:"type:varchar(64);primaryKey"`
	ProjectID          string         `gorm:"type:varchar(64);not null;index"`
	Type               string         `gorm:"type:varchar(32);not null;index"`
	Host               string         `gorm:"type:varchar(255);not null"`
	Port               int            `gorm:"not null"`
	Database           string         `gorm:"type:varchar(255);not null"`
	Username           string         `gorm:"type:varchar(255);not null"`
	PasswordSecretID   string         `gorm:"type:varchar(255)"`
	ExtraOptions       datatypes.JSON `gorm:"type:jsonb;not null;default:'{}'"`
	Environment        string         `gorm:"type:varchar(64);not null;index"`
	AllowedAgentLabels datatypes.JSON `gorm:"type:jsonb;not null;default:'[]'"`
	BackupFormat       string         `gorm:"type:varchar(32)"`
}

func (DatabaseTarget) TableName() string {
	return "databases"
}

type Run struct {
	ID         string         `gorm:"type:varchar(64);primaryKey"`
	ProjectID  string         `gorm:"type:varchar(64);not null;index"`
	PipelineID string         `gorm:"type:varchar(64);index"`
	Status     string         `gorm:"type:varchar(32);not null;index"`
	Source     string         `gorm:"type:varchar(32);not null;index"`
	Ref        string         `gorm:"type:varchar(255)"`
	Comment    string         `gorm:"type:text"`
	CommitSHA  string         `gorm:"type:varchar(128)"`
	Metadata   datatypes.JSON `gorm:"type:jsonb;not null;default:'{}'"`
	CreatedAt  time.Time      `gorm:"not null;index"`
	UpdatedAt  time.Time      `gorm:"not null;index"`
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

type Backup struct {
	ID          string    `gorm:"type:varchar(64);primaryKey"`
	ProjectID   string    `gorm:"type:varchar(64);index"`
	Environment string    `gorm:"type:varchar(64);index"`
	Type        string    `gorm:"type:varchar(32);index"`
	Database    string    `gorm:"type:varchar(255);index"`
	AgentID     string    `gorm:"type:varchar(128);index"`
	RunID       string    `gorm:"type:varchar(64);index"`
	JobID       string    `gorm:"type:varchar(128);index"`
	Path        string    `gorm:"type:text;not null"`
	Size        int64     `gorm:"not null;default:0"`
	Checksum    string    `gorm:"type:varchar(128)"`
	CreatedAt   time.Time `gorm:"not null;index"`
}
