# Changelog

All notable changes to go-harness-filters are documented here. The format
follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/) and the
project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## v0.1.0 — 2026-05-26

Initial cut. 27 tests across 5 packages, all `-race` clean. Two
subpackages callable today; remaining three ship contracts only.

### Added

- **`directive/` — full `@namespace:action key=value` parser.**
  - `Parse(string) (Directive, error)` with strict validation: leading
    `@` required, lowercased namespace + action with a-z/0-9/- charset,
    keys allow underscores too (for MCP-style conventions like
    `short_code`), quoted values with `\"` and `\\` escapes only,
    trailing-junk rejection.
  - `Directive.Raw` preserves the original text exactly as the agent
    emitted it, for audit and disclosure.
  - `Directive.Get(key)` convenience for first-value lookup; callers
    that need duplicates iterate `Directive.Args`.
  - `ErrMissingAtPrefix` sentinel.

- **`classify/` — rule-driven classifier.**
  - `Classifier` interface (`Classify(Input) Result`).
  - `Input` (Kind / Content / Metadata), `Result` (optional `Match`),
    `Match` (Intent / RuleID / Confidence / Source / Recommended /
    Reversible).
  - `Confidence` constants (exact / derived / inferred) — kept in sync
    with `go-runtime-events.Confidence` but defined locally so classify
    doesn't depend on the runtime-event schema.
  - `Rule` type with mutually-exclusive `ExactMatch` / `RegexpMatch`.
  - `NewRuleSet(...)` ordered rule set, first-match-wins; panics on
    misconfigured rules at construction time (both or neither matcher
    set) so a broken rule never silently fails to match in production.
  - `RuleSet.Classify` normalizes whitespace for exact-match rules;
    regex rules see raw content.
  - Recommended slice is deep-copied per Match so callers mutating
    results can't poison subsequent classifications.
  - `NaniteDeployRule` — worked example matching `go build -o nanite
    ./cmd/nanite` and variants (including JSON-encoded tool_use
    payloads), recommending the Cerberus deploy/reload pair,
    `Reversible=false` so the policy layer defaults to nudge.

- **`normalize/` — canonicalization contract.**
  - `Normalizer` interface, `Result` (Original / Normalized / AliasOf /
    Rejected / Reason), `Config` (Aliases / Reserved). Concrete domain
    rules deferred.

- **`repair/` — deterministic-repair contract.**
  - `Repairer` interface, `Input` (Kind / Content / Metadata), `Result`
    (Repaired / Replacement / Original / RuleID / Reason /
    SemanticChange). Architecture-doc guardrails encoded in field
    semantics: SemanticChange repairs MUST NOT auto-apply.

- **`event/` — normalized filter-event schema.**
  - `Kind` constants for filter.classified, filter.repaired,
    filter.normalized, filter.directive, filter.warning.
  - `Event` envelope (Kind / Time / RuleID / Original / Result / Notes
    / Context). Caller-supplied correlation metadata flows through
    `Context` verbatim.

- **Initial module scaffold from folio's `go-lib` preset** — CI
  workflow, MIT license.

### Notes

- `policy/` subpackage from the architecture doc is intentionally NOT
  in this module — it overlaps with `go-agent-wrapper/policy/` and the
  factoring (extract to a neutral module? duplicate? cross-import?)
  needs a real consumer driving the decision. See README and
  [ROADMAP.md](./ROADMAP.md).
- See [ROADMAP.md](./ROADMAP.md) for what's deferred from each
  subpackage.
