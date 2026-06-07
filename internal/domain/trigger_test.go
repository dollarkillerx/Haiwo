package domain

import "testing"

func TestTriggerMatchesBranchAndComment(t *testing.T) {
	trigger := Trigger{Type: "webhook", BranchPattern: "main", CommentPattern: "/deploy prod"}
	if !trigger.Matches(TriggerEvent{Type: "comment", Branch: "main", Comment: "/deploy prod"}) {
		t.Fatal("expected trigger to match branch + comment")
	}
	if trigger.Matches(TriggerEvent{Type: "comment", Branch: "feature/x", Comment: "/deploy prod"}) {
		t.Fatal("expected branch mismatch")
	}
	if trigger.Matches(TriggerEvent{Type: "comment", Branch: "main", Comment: "/deploy staging"}) {
		t.Fatal("expected comment mismatch")
	}
}

func TestTriggerMatchesWildcardBranch(t *testing.T) {
	trigger := Trigger{Type: "webhook", BranchPattern: "release/*"}
	if !trigger.Matches(TriggerEvent{Type: "push", Ref: "refs/heads/release/1.0"}) {
		t.Fatal("expected wildcard branch to match")
	}
}

func TestTriggerMatchesPushCommitMessage(t *testing.T) {
	trigger := Trigger{Type: "push", BranchPattern: "main", CommitPattern: "deploy"}
	if !trigger.Matches(TriggerEvent{Type: "push", Branch: "main", CommitMessage: "deploy api service"}) {
		t.Fatal("expected commit text to match")
	}
	if trigger.Matches(TriggerEvent{Type: "push", Branch: "main", CommitMessage: "fix tests"}) {
		t.Fatal("expected commit text mismatch")
	}
}

func TestTriggerMatchesTagPattern(t *testing.T) {
	trigger := Trigger{Type: "tag", TagPattern: "dev*"}
	if !trigger.Matches(TriggerEvent{Type: "tag", Tag: "dev-2026"}) {
		t.Fatal("expected tag pattern to match")
	}
	if trigger.Matches(TriggerEvent{Type: "tag", Tag: "prod-2026"}) {
		t.Fatal("expected tag pattern mismatch")
	}
}
