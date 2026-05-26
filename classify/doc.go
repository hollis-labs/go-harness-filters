// Package classify exposes the harness command and output classifier.
//
// Classifier output drives policy decisions, repair attempts, and
// runtime-event tagging. Per the harness-filters architecture note,
// the classifier uses the cheapest reliable signal:
//
//  1. Exact command match for known patterns.
//  2. Regex / token classifier for command families.
//  3. AST / schema-aware classifier for JSON, YAML, envelopes, MCP args.
//  4. Path-aware classifier for repo / app names.
//  5. Small local model or LLM only when deterministic rules fall short.
//
// This package owns the contract (Classifier interface, Result and
// Match types). Concrete classifier registries (deterministic command
// table, regex rule set, JSON-schema-aware) land in follow-up passes.
package classify
