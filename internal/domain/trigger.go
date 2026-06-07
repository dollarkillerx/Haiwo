package domain

import (
	"path/filepath"
	"strings"
)

type TriggerEvent struct {
	Type          string `json:"type"`
	Branch        string `json:"branch,omitempty"`
	Tag           string `json:"tag,omitempty"`
	Comment       string `json:"comment,omitempty"`
	CommitMessage string `json:"commit_message,omitempty"`
	CommitSHA     string `json:"commit_sha,omitempty"`
	Ref           string `json:"ref,omitempty"`
}

func (t Trigger) Matches(event TriggerEvent) bool {
	if t.Type != "" && event.Type != "" && t.Type != event.Type {
		if !(t.Type == "webhook" && (event.Type == "push" || event.Type == "tag" || event.Type == "comment")) {
			return false
		}
	}
	if t.BranchPattern != "" {
		branch := event.Branch
		if branch == "" {
			branch = strings.TrimPrefix(event.Ref, "refs/heads/")
		}
		ok, err := filepath.Match(t.BranchPattern, branch)
		if err != nil || !ok {
			return false
		}
	}
	if t.CommentPattern != "" {
		ok, err := filepath.Match(t.CommentPattern, event.Comment)
		if err != nil || !ok {
			return false
		}
	}
	if t.TagPattern != "" {
		if !matchExactOrGlob(t.TagPattern, event.Tag) {
			return false
		}
	}
	if t.CommitPattern != "" {
		if !matchContainsOrGlob(t.CommitPattern, event.CommitMessage) {
			return false
		}
	}
	return true
}

func matchExactOrGlob(pattern, value string) bool {
	if pattern == "" {
		return true
	}
	if hasGlob(pattern) {
		ok, err := filepath.Match(pattern, value)
		return err == nil && ok
	}
	return pattern == value
}

func matchContainsOrGlob(pattern, value string) bool {
	if pattern == "" {
		return true
	}
	if hasGlob(pattern) {
		ok, err := filepath.Match(pattern, value)
		return err == nil && ok
	}
	return strings.Contains(value, pattern)
}

func hasGlob(pattern string) bool {
	return strings.ContainsAny(pattern, "*?[")
}
