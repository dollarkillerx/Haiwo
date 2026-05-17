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
