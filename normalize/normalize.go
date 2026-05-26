package normalize

import (
	"strings"
	"unicode"
)

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

// SlugNormalizer canonicalizes free-form names/tags into lowercase
// hyphen-separated slugs. It is intentionally conservative: unknown
// punctuation collapses to hyphens, reserved terms are rejected, and
// aliases are applied after the initial slug pass.
type SlugNormalizer struct {
	Config Config

	// Singularize strips a trailing "s" from simple plural-looking words.
	// It is deliberately basic; domain-specific taxonomies should layer
	// richer morphology outside this generic normalizer.
	Singularize bool
}

// NewSlugNormalizer returns a conservative slug normalizer.
func NewSlugNormalizer(cfg Config) SlugNormalizer {
	return SlugNormalizer{Config: cfg}
}

// Normalize implements [Normalizer].
func (n SlugNormalizer) Normalize(input string) Result {
	original := input
	slug := slugify(input)
	if n.Singularize {
		slug = singularizeSlug(slug)
	}
	if slug == "" {
		return Result{Original: original, Rejected: true, Reason: "normalize.empty"}
	}

	if alias, ok := lookupAlias(n.Config.Aliases, slug); ok {
		slug = alias
		if n.Singularize {
			slug = singularizeSlug(slug)
		}
		if isReserved(slug, n.Config.Reserved) {
			return Result{Original: original, Rejected: true, Reason: "normalize.reserved"}
		}
		return Result{
			Original:   original,
			Normalized: slug,
			AliasOf:    slug,
			Reason:     "normalize.alias",
		}
	}

	if isReserved(slug, n.Config.Reserved) {
		return Result{Original: original, Rejected: true, Reason: "normalize.reserved"}
	}
	return Result{Original: original, Normalized: slug}
}

func slugify(input string) string {
	var b strings.Builder
	lastHyphen := false
	for _, r := range strings.TrimSpace(strings.ToLower(input)) {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			b.WriteRune(r)
			lastHyphen = false
		case unicode.IsSpace(r) || r == '-' || r == '_' || r == '/' || r == '.':
			if b.Len() > 0 && !lastHyphen {
				b.WriteByte('-')
				lastHyphen = true
			}
		default:
			if b.Len() > 0 && !lastHyphen {
				b.WriteByte('-')
				lastHyphen = true
			}
		}
	}
	return strings.Trim(b.String(), "-")
}

func singularizeSlug(slug string) string {
	parts := strings.Split(slug, "-")
	for i, p := range parts {
		if len(p) > 3 && strings.HasSuffix(p, "s") && !strings.HasSuffix(p, "ss") {
			parts[i] = strings.TrimSuffix(p, "s")
		}
	}
	return strings.Join(parts, "-")
}

func lookupAlias(aliases map[string]string, slug string) (string, bool) {
	if len(aliases) == 0 {
		return "", false
	}
	if alias, ok := aliases[slug]; ok {
		return slugify(alias), true
	}
	for k, v := range aliases {
		if slugify(k) == slug {
			return slugify(v), true
		}
	}
	return "", false
}

func isReserved(slug string, reserved []string) bool {
	for _, r := range reserved {
		if slugify(r) == slug {
			return true
		}
	}
	return false
}
