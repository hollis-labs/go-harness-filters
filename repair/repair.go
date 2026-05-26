package repair

// Repairer attempts one narrowly-scoped repair on input. Multiple
// repairers compose into a chain; the first one that produces a
// [Result] with Repaired=true short-circuits the chain.
type Repairer interface {
	Repair(in Input) Result
}

// Input is what the caller hands to [Repairer.Repair].
type Input struct {
	// Kind classifies the content ("envelope", "json", "mcp_call",
	// ...). Open string.
	Kind string

	// Content is the raw bytes the repairer should inspect.
	Content []byte

	// Metadata is optional channel-specific context (envelope type,
	// MCP tool name, schema hint, ...).
	Metadata map[string]string
}

// Result is what a [Repairer] returns. The zero value means "no
// repair was applied"; Content was correct (or the repairer did not
// recognize it) and the caller should pass Input.Content through
// unchanged.
type Result struct {
	// Repaired is true when the repairer made a change. False means
	// the input passed through untouched.
	Repaired bool

	// Replacement is the repaired bytes. Empty when Repaired is false.
	Replacement []byte

	// Original mirrors Input.Content for callers that record the
	// before/after pair. Populated whenever Repaired is true.
	Original []byte

	// RuleID names the rule that produced the repair, for audit and
	// stable disclosure.
	RuleID string

	// Reason is the human-readable explanation
	// ("missing closing brace at offset 124", "deprecated tool name
	// 'list_files' aliased to 'fs_list'").
	Reason string

	// SemanticChange is true when the repair changes the document's
	// meaning (not just its syntactic shape). Per the architecture
	// note, repairs with SemanticChange=true must NOT be auto-applied
	// — callers should warn and preserve the raw content.
	SemanticChange bool
}
