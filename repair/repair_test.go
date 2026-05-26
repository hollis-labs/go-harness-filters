package repair

import "testing"

type fakeRepairer struct{}

func (fakeRepairer) Repair(in Input) Result {
	if string(in.Content) != `{"k":"v"` {
		return Result{}
	}
	return Result{
		Repaired:    true,
		Original:    in.Content,
		Replacement: []byte(`{"k":"v"}`),
		RuleID:      "json.missing-closing-brace",
		Reason:      "appended missing closing brace",
	}
}

func TestRepairerContract(t *testing.T) {
	var r Repairer = fakeRepairer{}

	good := r.Repair(Input{Kind: "json", Content: []byte(`{"k":"v"}`)})
	if good.Repaired {
		t.Errorf("repairer flagged well-formed input as repaired")
	}

	bad := r.Repair(Input{Kind: "json", Content: []byte(`{"k":"v"`)})
	if !bad.Repaired {
		t.Fatal("repairer missed the malformed input")
	}
	if string(bad.Replacement) != `{"k":"v"}` {
		t.Errorf("Replacement = %q", bad.Replacement)
	}
	if string(bad.Original) != `{"k":"v"` {
		t.Errorf("Original not preserved: %q", bad.Original)
	}
}
