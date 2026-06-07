package server

import (
	"net/http"
	"testing"
)

func TestWebhookEventGitHubComment(t *testing.T) {
	event := webhookEvent(http.Header{"X-Github-Event": []string{"issue_comment"}}, map[string]any{
		"comment": map[string]any{"body": "/deploy staging"},
	})
	if event.Type != "comment" || event.Comment != "/deploy staging" {
		t.Fatalf("unexpected event: %#v", event)
	}
}

func TestWebhookEventGitHubPing(t *testing.T) {
	event := webhookEvent(http.Header{"X-Github-Event": []string{"ping"}}, map[string]any{})
	if event.Type != "ping" {
		t.Fatalf("unexpected event: %#v", event)
	}
}

func TestWebhookEventPush(t *testing.T) {
	event := webhookEvent(http.Header{}, map[string]any{"ref": "refs/heads/main"})
	if event.Type != "push" || event.Branch != "main" {
		t.Fatalf("unexpected event: %#v", event)
	}
}
