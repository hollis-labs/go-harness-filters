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

func TestChainReturnsFirstRepair(t *testing.T) {
	c := Chain{fakeRepairer{}, fakeRepairer{}}
	got := c.Repair(Input{Kind: "json", Content: []byte(`{"k":"v"`)})
	if !got.Repaired {
		t.Fatal("chain did not return repair")
	}
	if got.RuleID != "json.missing-closing-brace" {
		t.Errorf("RuleID = %q", got.RuleID)
	}
}

func TestMissingClosingDelimiterJSONRepairsObject(t *testing.T) {
	r := MissingClosingDelimiterJSON{}
	got := r.Repair(Input{Kind: "json", Content: []byte(`{"k":{"nested":true}`)})
	if !got.Repaired {
		t.Fatal("expected repair")
	}
	if string(got.Replacement) != `{"k":{"nested":true}}` {
		t.Errorf("Replacement = %q", got.Replacement)
	}
	if got.SemanticChange {
		t.Error("missing delimiter repair should not be semantic-changing")
	}
}

func TestMissingClosingDelimiterJSONRepairsArray(t *testing.T) {
	r := MissingClosingDelimiterJSON{}
	got := r.Repair(Input{Kind: "envelope", Content: []byte(`{"items":[1,2,3]`)})
	if !got.Repaired {
		t.Fatal("expected repair")
	}
	if string(got.Replacement) != `{"items":[1,2,3]}` {
		t.Errorf("Replacement = %q", got.Replacement)
	}
}

func TestMissingClosingDelimiterJSONIgnoresSemanticRepairs(t *testing.T) {
	r := MissingClosingDelimiterJSON{}
	for _, in := range []string{
		`{"k":"unterminated}`,
		`{"a":1 "b":2}`,
		`{"a":1]`,
		`{"a":}`,
	} {
		if got := r.Repair(Input{Kind: "json", Content: []byte(in)}); got.Repaired {
			t.Errorf("Repair(%q) repaired to %q; want no repair", in, got.Replacement)
		}
	}
}

func TestMissingClosingDelimiterJSONIgnoresOtherKinds(t *testing.T) {
	r := MissingClosingDelimiterJSON{}
	got := r.Repair(Input{Kind: "command", Content: []byte(`{"k":"v"`)})
	if got.Repaired {
		t.Errorf("unexpected repair for command kind: %+v", got)
	}
}
