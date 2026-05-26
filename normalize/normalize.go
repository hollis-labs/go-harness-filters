package normalize

// Normalizer canonicalizes one input value into a stable form, returning
// the normalized output plus provenance so callers can store both
// alongside each other.
type Normalizer interface {
	Normalize(input string) Result
}

// Result is the output of one normalization pass.
type Result struct {
	// Original is the unmodified input the caller passed to
	// [Normalizer.Normalize]. Always retained — apps store both
	// alongside each other so original intent is never lost.
	Original string

	// Normalized is the canonical form. Empty when [Result.Rejected]
	// is true.
	Normalized string

	// AliasOf, when non-empty, names the canonical term this input
	// was mapped to via an alias rule (e.g., "front end" → AliasOf =
	// "frontend").
	AliasOf string

	// Rejected is true when the normalizer refused the input (reserved
	// word, invalid characters that can't be folded, ...). Reason
	// carries the rule ID; Normalized is empty.
	Rejected bool

	// Reason is the human-readable explanation when Rejected is true
	// or when a non-obvious transform was applied.
	Reason string
}

// Config tunes a normalizer's behavior. Concrete normalizers wrap this
// with their own per-domain rule sets (singular noun preference,
// path-style, ...).
type Config struct {
	// Aliases maps source forms (already lowercased and trimmed) to
	// canonical normalized forms. Useful for "front end" → "frontend",
	// "knowledge bases" → "knowledge-base".
	Aliases map[string]string

	// Reserved lists terms the normalizer must refuse rather than
	// silently transform. Useful for keywords downstream systems
	// already reserved (e.g., "system", "admin", "root").
	Reserved []string
}
