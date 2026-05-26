package repair

import "encoding/json"

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

// Chain applies repairers in order and returns the first repair that
// changes input. Use this to compose narrow deterministic rules without
// giving later rules a chance to rewrite already-repaired content.
type Chain []Repairer

// Repair implements [Repairer].
func (c Chain) Repair(in Input) Result {
	for _, r := range c {
		if r == nil {
			continue
		}
		out := r.Repair(in)
		if out.Repaired {
			return out
		}
	}
	return Result{}
}

// MissingClosingDelimiterJSON repairs JSON that is otherwise complete
// but ended before one or more closing '}' / ']' delimiters. It never
// inserts commas, quotes, keys, or values, so it is syntactic-only.
type MissingClosingDelimiterJSON struct{}

// Repair implements [Repairer].
func (MissingClosingDelimiterJSON) Repair(in Input) Result {
	if in.Kind != "" && in.Kind != "json" && in.Kind != "envelope" {
		return Result{}
	}
	content := trimJSONSpace(in.Content)
	if len(content) == 0 || json.Valid(content) {
		return Result{}
	}

	closers, ok := missingJSONClosers(content)
	if !ok || len(closers) == 0 {
		return Result{}
	}
	replacement := append(append([]byte(nil), content...), closers...)
	if !json.Valid(replacement) {
		return Result{}
	}
	return Result{
		Repaired:       true,
		Replacement:    replacement,
		Original:       append([]byte(nil), in.Content...),
		RuleID:         "json.missing-closing-delimiter",
		Reason:         "appended missing JSON closing delimiter(s)",
		SemanticChange: false,
	}
}

func trimJSONSpace(b []byte) []byte {
	start, end := 0, len(b)
	for start < end && isJSONSpace(b[start]) {
		start++
	}
	for end > start && isJSONSpace(b[end-1]) {
		end--
	}
	return b[start:end]
}

func isJSONSpace(b byte) bool {
	return b == ' ' || b == '\n' || b == '\r' || b == '\t'
}

func missingJSONClosers(b []byte) ([]byte, bool) {
	stack := make([]byte, 0, 4)
	inString := false
	escaped := false

	for _, c := range b {
		if inString {
			if escaped {
				escaped = false
				continue
			}
			switch c {
			case '\\':
				escaped = true
			case '"':
				inString = false
			}
			continue
		}

		switch c {
		case '"':
			inString = true
		case '{':
			stack = append(stack, '}')
		case '[':
			stack = append(stack, ']')
		case '}', ']':
			if len(stack) == 0 || stack[len(stack)-1] != c {
				return nil, false
			}
			stack = stack[:len(stack)-1]
		}
	}
	if inString || escaped {
		return nil, false
	}

	closers := make([]byte, len(stack))
	for i := range stack {
		closers[i] = stack[len(stack)-1-i]
	}
	return closers, true
}
