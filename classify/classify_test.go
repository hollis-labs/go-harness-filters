package classify

import "testing"

type alwaysDeployClassifier struct{}

func (alwaysDeployClassifier) Classify(in Input) Result {
	if string(in.Content) != "go build -o nanite ./cmd/nanite" {
		return Result{}
	}
	return Result{Match: &Match{
		Intent:      "deploy.nanite",
		RuleID:      "hollis.deploy.nanite.cerberus-required",
		Confidence:  ConfidenceExact,
		Recommended: []string{"cerberus_resource_deploy nanite-api-service"},
		Reversible:  true,
	}}
}

func TestClassifierContract(t *testing.T) {
	var c Classifier = alwaysDeployClassifier{}

	hit := c.Classify(Input{Kind: "command", Content: []byte("go build -o nanite ./cmd/nanite")})
	if hit.Match == nil {
		t.Fatal("expected match for deploy-shaped command")
	}
	if hit.Match.Intent != "deploy.nanite" {
		t.Errorf("Intent = %q, want deploy.nanite", hit.Match.Intent)
	}
	if hit.Match.Confidence != ConfidenceExact {
		t.Errorf("Confidence = %q, want exact", hit.Match.Confidence)
	}

	miss := c.Classify(Input{Kind: "command", Content: []byte("ls")})
	if miss.Match != nil {
		t.Errorf("expected no match for benign command; got %+v", miss.Match)
	}
}
