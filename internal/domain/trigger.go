package domain

import (
	"path/filepath"
	"strings"
)

type TriggerEvent struct {
	Type    string `json:"type"`
	Branch  string `json:"branch,omitempty"`
	Tag     string `json:"tag,omitempty"`
	Comment string `json:"comment,omitempty"`
	Ref     string `json:"ref,omitempty"`
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
	return true
}
