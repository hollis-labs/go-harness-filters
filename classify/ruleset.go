package classify

import (
	"regexp"
	"strings"
)

// Rule is one entry in a [RuleSet]. Each rule declares its matcher
// (exactly one of ExactMatch or RegexpMatch) plus the classification
// metadata to emit when the rule matches an [Input].
//
// RuleID values should be stable, hierarchical, lowercase identifiers
// ("hollis.deploy.nanite.cerberus-required", "git.force-push.main").
// Downstream policy engines and audit logs reference them directly;
// changing one breaks persisted decisions.
type Rule struct {
	// ID is the stable, hierarchical rule identifier.
	ID string

	// Intent is the canonical name this rule assigns when it matches
	// ("deploy.nanite", "git.force-push", "filesystem.delete-tree").
	Intent string

	// ExactMatch makes the rule match when the input content equals
	// this string after whitespace normalization (TrimSpace +
	// internal-whitespace collapse). Set EXACTLY one of ExactMatch
	// or RegexpMatch.
	ExactMatch string

	// RegexpMatch makes the rule match when this regex matches the
	// input content. Use for command families and patterned matches.
	// Set EXACTLY one of ExactMatch or RegexpMatch.
	RegexpMatch *regexp.Regexp

	// Confidence is the classifier's certainty that a match means
	// what the rule says it means. Use [ConfidenceExact] for
	// deterministic patterns (exact match, anchored regex);
	// [ConfidenceDerived] for fuzzy patterns; [ConfidenceInferred]
	// only when no deterministic signal is available.
	Confidence Confidence

	// Recommended is the canonical command(s) the agent should prefer
	// over the matched input. Surfaced verbatim through the [Match]
	// to the policy layer.
	Recommended []string

	// Reversible signals whether the proposed alternative is safe to
	// rewrite/block on without operator approval. False is the safe
	// default for any rule whose Recommended commands change observable
	// system state (deploy, push, drop, delete); the policy layer
	// should NUDGE rather than REWRITE for these.
	Reversible bool
}

// RuleSet is a deterministic command classifier driven by an ordered
// list of [Rule]s. The first rule that matches an [Input] wins;
// subsequent rules are not consulted.
//
// Order matters: place more specific patterns earlier so a broad
// catch-all regex doesn't shadow a precise exact-match rule.
type RuleSet struct {
	rules []Rule
}

// NewRuleSet returns a RuleSet that consults the supplied rules in
// order. Panics if any rule sets neither ExactMatch nor RegexpMatch,
// or sets both — that's a programming error and silent acceptance
// would let a misconfigured rule never match.
func NewRuleSet(rules ...Rule) *RuleSet {
	for i, r := range rules {
		hasExact := r.ExactMatch != ""
		hasRegex := r.RegexpMatch != nil
		if hasExact == hasRegex { // both or neither
			panic(panicMsg(i, r, hasExact, hasRegex))
		}
	}
	out := make([]Rule, len(rules))
	copy(out, rules)
	return &RuleSet{rules: out}
}

// Rules returns a copy of the configured rule list in declaration
// order. Useful for diagnostics and audit dumps.
func (rs *RuleSet) Rules() []Rule {
	out := make([]Rule, len(rs.rules))
	copy(out, rs.rules)
	return out
}

// Classify implements [Classifier]. Returns the first rule's [Result]
// whose matcher accepts in.Content; otherwise returns a zero Result
// (Match == nil).
//
// Whitespace normalization for ExactMatch rules: leading/trailing
// whitespace is trimmed and runs of internal whitespace collapse to a
// single space. RegexpMatch rules see the raw content.
func (rs *RuleSet) Classify(in Input) Result {
	raw := string(in.Content)
	normalized := normalizeCommand(raw)
	for _, r := range rs.rules {
		if span, ok := matchRule(r, raw, normalized); ok {
			return Result{Match: &Match{
				Intent:      r.Intent,
				RuleID:      r.ID,
				Confidence:  r.Confidence,
				Source:      span,
				Recommended: append([]string(nil), r.Recommended...),
				Reversible:  r.Reversible,
			}}
		}
	}
	return Result{}
}

func matchRule(r Rule, raw, normalized string) (string, bool) {
	if r.ExactMatch != "" {
		if normalized == normalizeCommand(r.ExactMatch) {
			return normalized, true
		}
		return "", false
	}
	if loc := r.RegexpMatch.FindStringIndex(raw); loc != nil {
		return raw[loc[0]:loc[1]], true
	}
	return "", false
}

// normalizeCommand trims surrounding whitespace and collapses runs of
// internal whitespace (spaces, tabs, newlines) into single spaces.
// Used for ExactMatch normalization.
func normalizeCommand(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	var b strings.Builder
	b.Grow(len(s))
	inSpace := false
	for _, r := range s {
		if r == ' ' || r == '\t' || r == '\n' || r == '\r' {
			if !inSpace {
				b.WriteByte(' ')
				inSpace = true
			}
			continue
		}
		b.WriteRune(r)
		inSpace = false
	}
	return b.String()
}

func panicMsg(i int, r Rule, hasExact, hasRegex bool) string {
	switch {
	case !hasExact && !hasRegex:
		return "classify: rule " + r.ID + " at index " + itoa(i) + " sets neither ExactMatch nor RegexpMatch"
	default: // both
		return "classify: rule " + r.ID + " at index " + itoa(i) + " sets both ExactMatch and RegexpMatch; set exactly one"
	}
}

func itoa(i int) string {
	// avoid pulling strconv into a panic path
	if i == 0 {
		return "0"
	}
	var buf [20]byte
	pos := len(buf)
	neg := i < 0
	if neg {
		i = -i
	}
	for i > 0 {
		pos--
		buf[pos] = byte('0' + i%10)
		i /= 10
	}
	if neg {
		pos--
		buf[pos] = '-'
	}
	return string(buf[pos:])
}
