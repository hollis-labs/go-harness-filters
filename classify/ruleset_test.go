package classify

import (
	"reflect"
	"regexp"
	"testing"
)

func TestNewRuleSetPanicsOnInvalidMatcher(t *testing.T) {
	cases := []struct {
		name string
		rule Rule
	}{
		{
			name: "neither matcher set",
			rule: Rule{ID: "x"},
		},
		{
			name: "both matchers set",
			rule: Rule{
				ID:          "y",
				ExactMatch:  "ls",
				RegexpMatch: regexp.MustCompile(`^ls`),
			},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r == nil {
					t.Errorf("NewRuleSet did not panic; want panic for %s", c.name)
				}
			}()
			_ = NewRuleSet(c.rule)
		})
	}
}

func TestRuleSetFirstMatchWins(t *testing.T) {
	// Two rules that both match "danger" — the first one declared
	// should win. This is what makes specific-before-general ordering
	// meaningful for callers.
	specific := Rule{
		ID:          "test.specific",
		Intent:      "danger.specific",
		ExactMatch:  "rm -rf /",
		Confidence:  ConfidenceExact,
		Recommended: []string{"trash /tmp/foo"},
	}
	general := Rule{
		ID:          "test.general",
		Intent:      "danger.general",
		RegexpMatch: regexp.MustCompile(`rm`),
		Confidence:  ConfidenceDerived,
		Recommended: []string{"think twice"},
	}
	rs := NewRuleSet(specific, general)

	res := rs.Classify(Input{Kind: "command", Content: []byte("rm -rf /")})
	if res.Match == nil {
		t.Fatal("expected a match")
	}
	if res.Match.RuleID != "test.specific" {
		t.Errorf("first-match-wins violated: got %q, want test.specific", res.Match.RuleID)
	}
}

func TestRuleSetReturnsEmptyOnNoMatch(t *testing.T) {
	rs := NewRuleSet(Rule{
		ID:         "test.exact",
		Intent:     "x",
		ExactMatch: "ls",
		Confidence: ConfidenceExact,
	})
	res := rs.Classify(Input{Kind: "command", Content: []byte("pwd")})
	if res.Match != nil {
		t.Errorf("expected no match for 'pwd' against ls rule; got %+v", res.Match)
	}
}

func TestRuleSetExactMatchNormalizesWhitespace(t *testing.T) {
	rs := NewRuleSet(Rule{
		ID:         "test.exact",
		Intent:     "x",
		ExactMatch: "git push --force",
		Confidence: ConfidenceExact,
	})
	cases := []string{
		"git push --force",
		"  git push --force  ",
		"git  push   --force",
		"git\tpush --force",
	}
	for _, c := range cases {
		t.Run(c, func(t *testing.T) {
			res := rs.Classify(Input{Kind: "command", Content: []byte(c)})
			if res.Match == nil {
				t.Errorf("input %q did not match (whitespace normalization failed)", c)
			}
		})
	}
}

func TestRuleSetRegexMatchPopulatesSourceSpan(t *testing.T) {
	rs := NewRuleSet(Rule{
		ID:          "test.regex",
		Intent:      "danger",
		RegexpMatch: regexp.MustCompile(`rm\s+-rf\s+/\S*`),
		Confidence:  ConfidenceDerived,
	})
	res := rs.Classify(Input{
		Kind:    "agent_text",
		Content: []byte("the agent said: rm -rf /tmp/junk would clean that up"),
	})
	if res.Match == nil {
		t.Fatal("expected a match")
	}
	if res.Match.Source != "rm -rf /tmp/junk" {
		t.Errorf("Source = %q, want %q", res.Match.Source, "rm -rf /tmp/junk")
	}
}

func TestRuleSetClassifyCopiesRecommendedSlice(t *testing.T) {
	// Recommended slice in the returned Match must NOT alias the
	// rule's slice — otherwise a caller mutating the result would
	// poison subsequent classifications.
	original := []string{"alternative"}
	rs := NewRuleSet(Rule{
		ID:          "test.alias",
		Intent:      "x",
		ExactMatch:  "trigger",
		Confidence:  ConfidenceExact,
		Recommended: original,
	})
	res := rs.Classify(Input{Kind: "command", Content: []byte("trigger")})
	if res.Match == nil {
		t.Fatal("expected a match")
	}
	res.Match.Recommended[0] = "MUTATED"
	if original[0] != "alternative" {
		t.Errorf("rule's Recommended slice was aliased and got mutated to %q", original[0])
	}
}

func TestRuleSetRulesReturnsCopy(t *testing.T) {
	r := Rule{
		ID: "x", Intent: "y", ExactMatch: "z", Confidence: ConfidenceExact,
	}
	rs := NewRuleSet(r)
	got := rs.Rules()
	got[0].ID = "MUTATED"
	if rs.Rules()[0].ID != "x" {
		t.Error("Rules() returned a slice aliasing internal state; mutation leaked back")
	}
}

func TestNormalizeCommand(t *testing.T) {
	cases := map[string]string{
		"  ls  ":           "ls",
		"git   push":       "git push",
		"git\tpush\n--all": "git push --all",
		"":                 "",
		"   \t  ":          "",
	}
	for in, want := range cases {
		if got := normalizeCommand(in); got != want {
			t.Errorf("normalizeCommand(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestSatisfiesClassifierInterface(t *testing.T) {
	// Compile-time assertion that *RuleSet satisfies Classifier.
	var _ Classifier = NewRuleSet()
}

func TestEmptyRuleSetMatchesNothing(t *testing.T) {
	rs := NewRuleSet()
	res := rs.Classify(Input{Kind: "command", Content: []byte("anything")})
	if res.Match != nil {
		t.Errorf("empty RuleSet produced match: %+v", res.Match)
	}
}

func TestPanicMessageNamesRule(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic")
		}
		msg, _ := r.(string)
		if !contains(msg, "named-but-broken") {
			t.Errorf("panic message %q should include the rule's ID for diagnosis", msg)
		}
	}()
	_ = NewRuleSet(Rule{ID: "named-but-broken"})
}

func contains(haystack, needle string) bool {
	for i := 0; i+len(needle) <= len(haystack); i++ {
		if haystack[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}

func TestRuleSetClassifyMatchFields(t *testing.T) {
	// Verify every Match field flows through correctly.
	rule := Rule{
		ID:          "test.full",
		Intent:      "danger.deletetree",
		RegexpMatch: regexp.MustCompile(`^rm -rf /`),
		Confidence:  ConfidenceExact,
		Recommended: []string{"trash <path>", "git rm <path>"},
		Reversible:  false,
	}
	rs := NewRuleSet(rule)
	res := rs.Classify(Input{Kind: "command", Content: []byte("rm -rf /")})
	if res.Match == nil {
		t.Fatal("expected a match")
	}
	want := &Match{
		Intent:      "danger.deletetree",
		RuleID:      "test.full",
		Confidence:  ConfidenceExact,
		Source:      "rm -rf /",
		Recommended: []string{"trash <path>", "git rm <path>"},
		Reversible:  false,
	}
	if !reflect.DeepEqual(res.Match, want) {
		t.Errorf("Match = %#v\n  want %#v", res.Match, want)
	}
}
