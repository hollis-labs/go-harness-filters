# go-harness-filters

Fast, auditable filters for harness IO — directive parsing, classification,
slug normalization and deterministic JSON repair — usable inside
`go-agent-wrapper` or directly by any host. Every subpackage is a pure
transform over text or structured values: nothing here executes a command,
enforces a policy, or reaches the network.

## Start Here

- `README.md` states what each of the five subpackages does today.
- `ROADMAP.md` records deferred scope and the open policy-boundary question.
- `directive/directive.go` parses `@namespace:action key=value`.
- `classify/ruleset.go` drives rule matching; `classify/rules.go` holds the
  shipped rules.
- `normalize/normalize.go` owns slug canonicalization.
- `repair/repair.go` owns JSON delimiter repair and `Chain`.
- `event/event.go` is the event schema, which has no emitter helper yet.

## Commands

```bash
go vet ./...
go test -race -count=1 ./...
```

CI runs both.

## Boundaries

Repair is syntactic only and must stay that way: it appends missing `}` or `]`
and keeps the result only when the repaired document parses as JSON. It never
inserts commas, quotes, keys or values, because inventing content would turn a
truncated payload into a plausible wrong one.
`TestMissingClosingDelimiterJSONIgnoresSemanticRepairs` is the guard.

Classification recommends; it does not decide. Rules carry `Reversible`, and a
non-reversible match is what makes the policy layer default to a nudge rather
than a block. Enforcement lives in the consumer — see the open policy-boundary
question in `ROADMAP.md` before adding a `policy/` package here, since
`go-agent-wrapper/policy/` already owns that surface.

Rule matching is first-match-wins over an ordered set, and both the rule list
and the recommendation slice are returned as copies so a caller cannot mutate a
shared `RuleSet` (`TestRuleSetFirstMatchWins`, `TestRuleSetRulesReturnsCopy`,
`TestRuleSetClassifyCopiesRecommendedSlice`). An invalid matcher panics at
construction, naming the rule, rather than silently never matching.

`Directive.Raw` preserves the original text. Provenance is the reason these
filters are auditable; dropping it to save an allocation removes the ability to
explain a decision after the fact.
