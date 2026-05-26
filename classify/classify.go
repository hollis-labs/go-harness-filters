package classify

// Classifier produces a [Result] for one input. Implementations choose
// the cheapest reliable signal that produces a Match; the registry
// pattern composes multiple classifiers in priority order.
type Classifier interface {
	Classify(in Input) Result
}

// Input is what the caller hands to [Classifier.Classify]. Kind names
// the content type so multi-domain classifiers can dispatch.
type Input struct {
	// Kind classifies the content ("command", "tool_output",
	// "command_output", "envelope", ...). Open string.
	Kind string

	// Content is the raw bytes (command string, JSON document,
	// stdout chunk, ...).
	Content []byte

	// Metadata is optional channel-specific context (provider name,
	// tool name, cwd hint, ...).
	Metadata map[string]string
}

// Result is what a [Classifier] returns. Zero value (no Match) means
// "the classifier could not classify this input"; callers should treat
// that as "fall through" rather than "definitely benign".
type Result struct {
	// Match, when non-nil, describes the matched intent and the rule
	// that matched it.
	Match *Match
}

// Match describes one classification hit.
type Match struct {
	// Intent is the canonical name the classifier assigns
	// ("deploy.nanite", "filesystem.delete", "git.force-push", ...).
	Intent string

	// RuleID names the specific rule that matched. Stable across
	// releases so audit logs and policy decisions can reference it.
	RuleID string

	// Confidence is exact / derived / inferred. PTY-text heuristics
	// produce "inferred"; deterministic exact matches and AST-aware
	// classifiers produce "exact".
	Confidence Confidence

	// Source is the textual span inside Input.Content the rule
	// matched. May be empty when the rule matches the whole input.
	Source string

	// Recommended is an optional list of canonical commands the
	// agent should prefer (e.g., for a deploy intent: the right
	// Cerberus invocation).
	Recommended []string

	// Reversible indicates whether the proposed alternative is safe
	// to rewrite/block on without operator approval.
	Reversible bool
}

// Confidence mirrors github.com/hollis-labs/go-runtime-events.Confidence
// but is defined locally so classify does not pull in the runtime-event
// schema. The two are intentionally kept in sync — see the architecture
// note.
type Confidence string

const (
	ConfidenceExact    Confidence = "exact"
	ConfidenceDerived  Confidence = "derived"
	ConfidenceInferred Confidence = "inferred"
)
