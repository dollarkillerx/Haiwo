package server

import (
	"testing"
	"time"

	"github.com/haiwo-ci/haiwo/internal/domain"
)

func TestSaveRunPreservesExistingMetadata(t *testing.T) {
	now := time.Date(2026, 6, 7, 16, 30, 0, 0, time.UTC)
	store := NewStore(func() time.Time { return now }, nil)

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
