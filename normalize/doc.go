// Package normalize canonicalizes tags, slugs, names, and paths to a
// deterministic form suitable for stable IDs, taxonomy keys, and
// cross-app metadata.
//
// Per the harness-filters architecture note, normalization is
// deterministic-first — LLM classification belongs in classify, not
// here. This package owns:
//
//   - Unicode trim/fold/normalize
//   - Slugify for IDs and URLs
//   - Singularize / pluralize for domains that want canonical noun forms
//   - Alias map for domain-specific terms (e.g., "front end" → "frontend")
//   - Reserved-word guard
//   - Collision handling
//   - Provenance (original + normalized kept together)
//
// The current pass exposes the contract (Normalizer interface,
// Result type, alias-map and reserved-word configuration). Concrete
// rule sets land per-domain (Tesseract taxonomy first).
package normalize
