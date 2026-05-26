package event

import "time"

// Kind names a normalized filter event. Open string — new kinds may be
// added without a schema bump; consumers must tolerate unknown kinds.
type Kind string

const (
	// KindClassified is emitted when a classifier produces a Match
	// against an input.
	KindClassified Kind = "filter.classified"

	// KindRepaired is emitted when a repairer rewrites an input.
	// Original and Replacement live in the payload for audit.
	KindRepaired Kind = "filter.repaired"

	// KindNormalized is emitted when a normalizer canonicalizes a
	// value (tag, slug, path) into a different form than the
	// original.
	KindNormalized Kind = "filter.normalized"

	// KindDirective is emitted when the directive parser accepts a
	// well-formed directive from agent text or tool output.
	KindDirective Kind = "filter.directive"

	// KindWarning is emitted when the pipeline detects a problem it
	// declines to auto-fix (e.g., a semantic-changing repair refused
	// per policy).
	KindWarning Kind = "filter.warning"
)

// Event is the on-the-wire envelope for a filter-emitted observation.
//
// Filter events carry less context than runtime events because they
// describe pipeline output rather than process lifecycle. Apps that
// need to correlate filter events back to a wrapper session attach
// their own correlation IDs via [Event.Context].
type Event struct {
	// Kind names what the pipeline observed. Required.
	Kind Kind `json:"kind"`

	// Time is the wall-clock timestamp the event was produced.
	Time time.Time `json:"time"`

	// RuleID is the stable identifier of the rule that produced this
	// event. Required for KindClassified, KindRepaired,
	// KindNormalized; optional for KindWarning and KindDirective.
	RuleID string `json:"rule_id,omitempty"`

	// Original is the input the pipeline observed. Always retained
	// for audit.
	Original []byte `json:"original,omitempty"`

	// Result is the pipeline's output (rewritten content, normalized
	// value, classification match details). Encoding is kind-specific.
	Result []byte `json:"result,omitempty"`

	// Notes carries human-readable explanation
	// ("repaired malformed envelope", "aliased deprecated tool name").
	Notes []string `json:"notes,omitempty"`

	// Context is caller-supplied correlation metadata
	// ("wrapper_session_id", "turn_id", "tool_name", ...). Pass
	// through to downstream consumers verbatim.
	Context map[string]string `json:"context,omitempty"`
}
