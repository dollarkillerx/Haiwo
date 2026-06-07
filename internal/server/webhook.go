package server

import (
	"net/http"
	"strings"

	"github.com/haiwo-ci/haiwo/internal/domain"
)

func webhookEvent(headers http.Header, raw map[string]any) domain.TriggerEvent {
	if ev := headers.Get("X-GitHub-Event"); ev == "ping" {
		return domain.TriggerEvent{Type: "ping"}
	}
	if comment := nestedString(raw, "comment", "body"); comment != "" {
		return domain.TriggerEvent{Type: "comment", Comment: comment, Branch: branchFromRaw(raw), Ref: refFromRaw(raw)}
	}
	if note := nestedString(raw, "object_attributes", "note"); note != "" {
		return domain.TriggerEvent{Type: "comment", Comment: note, Branch: branchFromRaw(raw), Ref: refFromRaw(raw)}
	}
	ref := refFromRaw(raw)
	if strings.HasPrefix(ref, "refs/tags/") {
		return domain.TriggerEvent{Type: "tag", Tag: strings.TrimPrefix(ref, "refs/tags/"), Ref: ref, CommitSHA: commitSHAFromRaw(raw), CommitMessage: commitMessageFromRaw(raw)}
	}
	if ref != "" {
		return domain.TriggerEvent{Type: "push", Branch: strings.TrimPrefix(ref, "refs/heads/"), Ref: ref, CommitSHA: commitSHAFromRaw(raw), CommitMessage: commitMessageFromRaw(raw)}
	}
	if ev := headers.Get("X-GitHub-Event"); ev == "issue_comment" || ev == "pull_request_review_comment" {
		return domain.TriggerEvent{Type: "comment", Comment: nestedString(raw, "comment", "body")}
	}
	return domain.TriggerEvent{Type: "webhook"}
}

func commitSHAFromRaw(raw map[string]any) string {
	if v, _ := raw["after"].(string); v != "" {
		return v
	}
	if v, _ := raw["checkout_sha"].(string); v != "" {
		return v
	}
	if v := nestedString(raw, "head_commit", "id"); v != "" {
		return v
	}
	if v := nestedString(raw, "commits", "0", "id"); v != "" {
		return v
	}
	return ""
}

func commitMessageFromRaw(raw map[string]any) string {
	messages := []string{}
	if v := nestedString(raw, "head_commit", "message"); v != "" {
		messages = append(messages, v)
	}
	if commits, ok := raw["commits"].([]any); ok {
		for _, item := range commits {
			if commit, ok := item.(map[string]any); ok {
				if message, _ := commit["message"].(string); message != "" {
					messages = append(messages, message)
				}
			}
		}
	}
	return strings.Join(messages, "\n")
}

func refFromRaw(raw map[string]any) string {
	if v, _ := raw["ref"].(string); v != "" {
		return v
	}
	if v := nestedString(raw, "object_attributes", "source_branch"); v != "" {
		return "refs/heads/" + v
	}
	if v := nestedString(raw, "pull_request", "head", "ref"); v != "" {
		return "refs/heads/" + v
	}
	return ""
}

func branchFromRaw(raw map[string]any) string {
	ref := refFromRaw(raw)
	return strings.TrimPrefix(ref, "refs/heads/")
}

func nestedString(raw map[string]any, path ...string) string {
	var cur any = raw
	for _, key := range path {
		m, ok := cur.(map[string]any)
		if !ok {
			return ""
		}
		cur = m[key]
	}
	v, _ := cur.(string)
	return v
}
