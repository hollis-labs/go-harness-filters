package classify

import (
	"reflect"
	"testing"
)

func TestNaniteDeployRuleMatchesArchitectureDocExample(t *testing.T) {
	// The exact example from
	// chrispian/inbox/cli-runner-wrapper-architecture-2026-05-26.md:
	//
	//   go build -o nanite ./cmd/nanite
	rs := NewRuleSet(NaniteDeployRule)
	res := rs.Classify(Input{
		Kind:    "command",
		Content: []byte("go build -o nanite ./cmd/nanite"),
	})
	if res.Match == nil {
		t.Fatal("expected a match for the doc's worked example")
	}
	if res.Match.RuleID != "hollis.deploy.nanite.cerberus-required" {
		t.Errorf("RuleID = %q, want hollis.deploy.nanite.cerberus-required", res.Match.RuleID)
	}
	if res.Match.Intent != "deploy.nanite" {
		t.Errorf("Intent = %q, want deploy.nanite", res.Match.Intent)
	}
	want := []string{
		"cerberus_resource_deploy nanite-api-service",
		"cerberus_resource_reload nanite-api-service",
	}
	if !reflect.DeepEqual(res.Match.Recommended, want) {
		t.Errorf("Recommended = %v\n  want %v", res.Match.Recommended, want)
	}
	if res.Match.Reversible {
		t.Error("Reversible = true; want false (rewrites change semantics — policy should NUDGE)")
	}
	if res.Match.Confidence != ConfidenceExact {
		t.Errorf("Confidence = %q, want exact", res.Match.Confidence)
	}
}

func TestNaniteDeployRuleMatchesVariants(t *testing.T) {
	rs := NewRuleSet(NaniteDeployRule)
	cases := []string{
		"go build -o nanite ./cmd/nanite",
		"go build ./cmd/nanite",
		"go build -o /tmp/nanite ./cmd/nanite",
		"go build -o nanite ./cmd/nanite -tags prod",
		"  go   build   ./cmd/nanite  ",
	}
	for _, c := range cases {
		t.Run(c, func(t *testing.T) {
			res := rs.Classify(Input{Kind: "command", Content: []byte(c)})
			if res.Match == nil {
				t.Errorf("expected match for variant %q", c)
			}
		})
	}
}

func TestNaniteDeployRuleDoesNotMatchUnrelatedCommands(t *testing.T) {
	// Critical: the rule must NOT flag legitimate non-deploy
	// commands that mention cmd/nanite (tests, lint runs, etc.).
	rs := NewRuleSet(NaniteDeployRule)
	negatives := []string{
		"go test ./cmd/nanite",
		"go vet ./cmd/nanite",
		"golangci-lint run ./cmd/nanite",
		"cat cmd/nanite/main.go",
		"ls cmd/nanite",
		"go build ./cmd/other",
		"",
		"echo hello",
	}
	for _, c := range negatives {
		t.Run(c, func(t *testing.T) {
			res := rs.Classify(Input{Kind: "command", Content: []byte(c)})
			if res.Match != nil {
				t.Errorf("rule should NOT match %q; got %+v", c, res.Match)
			}
		})
	}
}

func TestNaniteDeployRuleMatchesJSONWrappedCommand(t *testing.T) {
	// The wrapper passes JSON-encoded tool_use payloads to the
	// policy engine. The rule must match the command embedded
	// inside a JSON string field.
	rs := NewRuleSet(NaniteDeployRule)
	jsonForms := []string{
		`{"command":"go build -o nanite ./cmd/nanite"}`,
		`{"id":"t","name":"Bash","input":{"command":"go build -o nanite ./cmd/nanite"}}`,
		`tool_use: go build ./cmd/nanite`,
	}
	for _, c := range jsonForms {
		t.Run(c, func(t *testing.T) {
			res := rs.Classify(Input{Kind: "tool_use", Content: []byte(c)})
			if res.Match == nil {
				t.Errorf("expected match for JSON-wrapped command in %q", c)
			}
		})
	}
}

func TestNaniteDeployRuleMatchesInsideAgentText(t *testing.T) {
	// Agent text may contain the command embedded in prose. The
	// rule's regex anchors at start-of-line OR whitespace so a
	// command embedded after a label still matches.
	rs := NewRuleSet(NaniteDeployRule)
	cases := []string{
		"I'll run: go build -o nanite ./cmd/nanite",
		"go build -o nanite ./cmd/nanite\nthen restart the service",
	}
	for _, c := range cases {
		t.Run(c, func(t *testing.T) {
			res := rs.Classify(Input{Kind: "agent_text", Content: []byte(c)})
			if res.Match == nil {
				t.Errorf("expected match for embedded command in %q", c)
			}
		})
	}
}
