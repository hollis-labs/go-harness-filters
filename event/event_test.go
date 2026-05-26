package event

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestEventJSONRoundTrip(t *testing.T) {
	ev := Event{
		Kind:     KindRepaired,
		Time:     time.Date(2026, 5, 26, 12, 0, 0, 0, time.UTC),
		RuleID:   "json.missing-closing-brace",
		Original: []byte(`{"k":"v"`),
		Result:   []byte(`{"k":"v"}`),
		Notes:    []string{"appended missing closing brace"},
		Context:  map[string]string{"session_id": "ses_abc", "tool_name": "fs_read"},
	}

	raw, err := json.Marshal(ev)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var got Event
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.Kind != ev.Kind || got.RuleID != ev.RuleID {
		t.Errorf("round-trip lost identity:\n  want %+v\n  got  %+v", ev, got)
	}
}

func TestEventOmitsUnsetOptionals(t *testing.T) {
	ev := Event{Kind: KindWarning, Time: time.Date(2026, 5, 26, 0, 0, 0, 0, time.UTC)}
	raw, _ := json.Marshal(ev)
	s := string(raw)
	for _, key := range []string{"rule_id", "original", "result", "notes", "context"} {
		if strings.Contains(s, `"`+key+`"`) {
			t.Errorf("unset optional %q appears in JSON: %s", key, s)
		}
	}
}

func TestUnknownKindRoundTrips(t *testing.T) {
	ev := Event{Kind: Kind("filter.future-thing-we-dont-know"), Time: time.Date(2026, 5, 26, 0, 0, 0, 0, time.UTC)}
	raw, _ := json.Marshal(ev)
	var got Event
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.Kind != ev.Kind {
		t.Errorf("unknown Kind not preserved: got %q want %q", got.Kind, ev.Kind)
	}
}
