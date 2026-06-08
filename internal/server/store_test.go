package server

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/haiwo-ci/haiwo/internal/domain"
)

func TestSaveRunPreservesExistingMetadata(t *testing.T) {
	now := time.Date(2026, 6, 7, 16, 30, 0, 0, time.UTC)
	store := NewStore(func() time.Time { return now }, StorageOptions{})

	run := domain.Run{
		ID:         "run_test",
		ProjectID:  "prj_test",
		PipelineID: "pipe_test",
		Status:     domain.RunQueued,
		Source:     "manual",
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	store.SaveRun(run)
	store.AppendRunLog(run.ID, "error", "no matching online agent", now)

	run.Status = domain.RunFailed
	run.UpdatedAt = now.Add(time.Second)
	store.SaveRun(run)

	got, ok := store.GetRun(run.ID)
	if !ok {
		t.Fatal("run was not saved")
	}
	if got.Metadata["error"] != "no matching online agent" {
		t.Fatalf("expected metadata error to be preserved, got %#v", got.Metadata)
	}
	if got.Metadata["logs"] == "" {
		t.Fatal("expected logs to be preserved")
	}
}

func TestFileStoragePersistsAcrossRestart(t *testing.T) {
	now := time.Date(2026, 6, 7, 16, 30, 0, 0, time.UTC)
	dataFile := filepath.Join(t.TempDir(), "haiwo.json")
	opts := StorageOptions{DataFile: dataFile}

	store := NewStore(func() time.Time { return now }, opts)
	store.SaveProject(domain.Project{ID: "prj_1", Name: "demo", Provider: "github", CreatedAt: now})
	store.SaveTrigger(domain.Trigger{ID: "trg_1", ProjectID: "prj_1", PipelineID: "pipe_1", Type: "push", CreatedAt: now})
	store.SaveAgent(domain.AgentInfo{ID: "agt_1", Name: "runner", Status: domain.AgentOnline, CurrentRun: 2, CreatedAt: now})
	store.SaveSettings(domain.SystemSettings{ServerBaseURL: "https://ci.example.com"})

	// Simulate a restart: a fresh store loading the same file.
	reloaded := NewStore(func() time.Time { return now }, opts)

	if got := reloaded.ListProjects(); len(got) != 1 || got[0].Name != "demo" {
		t.Fatalf("expected project to be reloaded, got %#v", got)
	}
	if got := reloaded.ListTriggers("prj_1"); len(got) != 1 {
		t.Fatalf("expected trigger to be reloaded, got %#v", got)
	}
	if got := reloaded.GetSettings(); got.ServerBaseURL != "https://ci.example.com" {
		t.Fatalf("expected settings to be reloaded, got %#v", got)
	}
	agent, ok := reloaded.GetAgent("agt_1")
	if !ok {
		t.Fatal("expected agent to be reloaded")
	}
	if agent.Status != domain.AgentOffline || agent.CurrentRun != 0 {
		t.Fatalf("expected reloaded agent to be reset offline, got status=%q currentRun=%d", agent.Status, agent.CurrentRun)
	}

	// Deletions must also survive a restart.
	reloaded.DeleteTriggersForPipeline("pipe_1")
	reloaded.DeleteAgent("agt_1")
	final := NewStore(func() time.Time { return now }, opts)
	if got := final.ListTriggers("prj_1"); len(got) != 0 {
		t.Fatalf("expected triggers to be deleted, got %#v", got)
	}
	if _, ok := final.GetAgent("agt_1"); ok {
		t.Fatal("expected agent to be deleted")
	}
}

func TestFileBackendStreamsLogsToSeparateFile(t *testing.T) {
	now := time.Date(2026, 6, 7, 16, 30, 0, 0, time.UTC)
	dir := t.TempDir()
	stateFile := filepath.Join(dir, "haiwo.json")
	logFilePath := filepath.Join(dir, "haiwo.log")
	opts := StorageOptions{DataFile: stateFile, LogFile: logFilePath, LogMaxBytes: 1 << 20}

	store := NewStore(func() time.Time { return now }, opts)
	store.SaveRun(domain.Run{ID: "run_1", ProjectID: "prj_1", Status: domain.RunRunning, CreatedAt: now, UpdatedAt: now})

	// Capture the state file size before streaming logs.
	beforeInfo, err := os.Stat(stateFile)
	if err != nil {
		t.Fatalf("stat state file: %v", err)
	}
	for i := 0; i < 200; i++ {
		store.AppendRunLog("run_1", "stdout", "log line content", now)
	}

	// Logs must land in the separate log file, prefixed with the run id.
	logData, err := os.ReadFile(logFilePath)
	if err != nil {
		t.Fatalf("read log file: %v", err)
	}
	if !strings.Contains(string(logData), "[run_1] ") || !strings.Contains(string(logData), "log line content") {
		t.Fatalf("expected run logs in log file, got: %q", string(logData))
	}
	if got := strings.Count(string(logData), "log line content"); got != 200 {
		t.Fatalf("expected 200 log lines in log file, got %d", got)
	}

	// Streaming logs must NOT rewrite the state file (that was the O(N^2) bug).
	afterInfo, err := os.Stat(stateFile)
	if err != nil {
		t.Fatalf("stat state file: %v", err)
	}
	if !afterInfo.ModTime().Equal(beforeInfo.ModTime()) {
		t.Fatal("state file was rewritten while streaming logs; logs should bypass the state snapshot")
	}

	// The capped buffer is still available in memory for the UI.
	run, ok := store.GetRun("run_1")
	if !ok || run.Metadata["logs"] == "" {
		t.Fatalf("expected in-memory log buffer to be populated, got %#v", run.Metadata)
	}
}

func TestFileBackendLogRotation(t *testing.T) {
	now := time.Date(2026, 6, 7, 16, 30, 0, 0, time.UTC)
	dir := t.TempDir()
	logFilePath := filepath.Join(dir, "haiwo.log")
	// Tiny cap so a handful of lines forces rotation.
	opts := StorageOptions{DataFile: filepath.Join(dir, "haiwo.json"), LogFile: logFilePath, LogMaxBytes: 200}

	store := NewStore(func() time.Time { return now }, opts)
	store.SaveRun(domain.Run{ID: "run_1", Status: domain.RunRunning, CreatedAt: now, UpdatedAt: now})

	for i := 0; i < 100; i++ {
		store.AppendRunLog("run_1", "stdout", "some reasonably long log line to fill the file", now)
	}

	// Active file must stay within the cap (plus one line of slack).
	info, err := os.Stat(logFilePath)
	if err != nil {
		t.Fatalf("stat log file: %v", err)
	}
	if info.Size() > 400 {
		t.Fatalf("active log file exceeded cap after rotation: %d bytes", info.Size())
	}
	// A rotated segment must exist (old logs preserved as exactly one backup).
	if _, err := os.Stat(logFilePath + ".1"); err != nil {
		t.Fatalf("expected rotated log segment %s.1: %v", logFilePath, err)
	}
}
