package agent

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/haiwo-ci/haiwo/internal/domain"
)

func TestRunCommandsKeepsShellState(t *testing.T) {
	workDir := t.TempDir()
	a := New(Config{WorkDir: workDir})
	task := domain.TaskPayload{
		TaskID:  "task_test",
		JobType: domain.JobCommand,
		Commands: []string{
			"mkdir -p target",
			"cd target",
			"pwd > ../pwd.txt",
		},
	}

	if err := a.runCommands(context.Background(), task); err != nil {
		t.Fatalf("run commands: %v", err)
	}

	got, err := os.ReadFile(filepath.Join(workDir, "tasks", task.TaskID, "pwd.txt"))
	if err != nil {
		t.Fatalf("read pwd output: %v", err)
	}
	want, err := filepath.EvalSymlinks(filepath.Join(workDir, "tasks", task.TaskID, "target"))
	if err != nil {
		t.Fatalf("resolve target path: %v", err)
	}
	actual, err := filepath.EvalSymlinks(strings.TrimSpace(string(got)))
	if err != nil {
		t.Fatalf("resolve pwd output: %v", err)
	}
	if actual != want {
		t.Fatalf("expected pwd %q, got %q", want, actual)
	}
}
