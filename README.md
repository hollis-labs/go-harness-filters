# go-harness-filters

## Moved to substrate

This standalone repository is deprecated. New development lives in the
[`github.com/hollis-labs/substrate/harness`](https://github.com/hollis-labs/substrate/tree/harness/v0.3.0/harness)
module, released as **`harness/v0.3.0`**.

```sh
go get github.com/hollis-labs/substrate/harness@v0.3.0
```

Follow the [package and API migration guide](https://github.com/hollis-labs/substrate/blob/harness/v0.3.0/harness/docs/units/go-harness-filters/MIGRATION.md) when updating imports;
the consolidation can include API changes. Existing standalone tags and history
are preserved. The documentation below describes the standalone releases and
is retained for historical reference. Applications migrate separately; this
redirect does not deploy or update any consumer.

Fast, auditable classification, normalization, repair, and directive
pipeline for harness IO — commands, tool output, envelopes, tags, agent
text. Usable inside `go-agent-wrapper` and directly from Nanite, Torque,
Tether, Hadron, Tesseract, and Stack Explorer.

## Status (v0.1.0, 2026-05-26)

Four subpackages are callable today:

- **`directive/`** — full `@namespace:action key=value` parser with
  strict validation, quoted values, escape handling, and provenance
  (`Directive.Raw` preserves the original).
- **`classify/`** — `Classifier` interface + `RuleSet` driver. Ships
  the worked `NaniteDeployRule` from the architecture doc (matches
  bare commands AND JSON-encoded tool_use payloads via word
  boundaries; recommends the Cerberus deploy/reload pair; marked
  `Reversible=false` so the policy layer defaults to nudge).
- **`normalize/`** — `SlugNormalizer` for conservative lowercase,
  hyphen-separated tag/name/path canonicalization with aliases,
  reserved-term rejection, provenance, and optional simple singularizing.
- **`repair/`** — deterministic `MissingClosingDelimiterJSON` repair
  plus `Chain` composition. It only appends missing `}` / `]` delimiters
  when the repaired document validates as JSON; it does not insert
  commas, quotes, keys, or values.

The remaining `event/` subpackage ships its schema but no reference
emitter helper yet.

27 tests across 5 packages, all `-race` clean.

See [ROADMAP.md](./ROADMAP.md) for deferred scope and the policy-boundary
decision.

Module path: `github.com/hollis-labs/go-harness-filters`

## Quickstart

### Parse a directive

```go
import "github.com/hollis-labs/go-harness-filters/directive"

d, err := directive.Parse(`@fragment:capture kind=decision title="Use Cerberus for Nanite deploys"`)
if err != nil { /* ... */ }
// d.Namespace == "fragment"
// d.Action    == "capture"
// d.Get("kind") == "decision"
// d.Get("title") == "Use Cerberus for Nanite deploys"
```

### Classify a command against rules

```go
import "github.com/hollis-labs/go-harness-filters/classify"

rules := classify.NewRuleSet(classify.NaniteDeployRule)
result := rules.Classify(classify.Input{
    Kind:    "command",
    Content: []byte("go build -o nanite ./cmd/nanite"),
})
// result.Match.RuleID == "hollis.deploy.nanite.cerberus-required"
// result.Match.Recommended == [cerberus_resource_deploy nanite-api-service, ...]
// result.Match.Reversible == false
```

### Drive wrapper policy decisions

The classifier output feeds wrapper policy via
`go-agent-wrapper/classifybridge`:

```go
import (
    "github.com/hollis-labs/go-agent-wrapper/classifybridge"
    "github.com/hollis-labs/go-harness-filters/classify"
)

rules  := classify.NewRuleSet(classify.NaniteDeployRule)
engine := &classifybridge.Engine{Classifier: rules}

// Hand engine to wrapper.Config.Policy — see go-agent-wrapper README.
```

## Subpackages

| Path | What it owns | Status |
|---|---|---|
| `directive/` | Parser for the `@namespace:action key=value` directive syntax. Strict by design — silent malformed-directive handling would let agents submit near-misses that don't do what they look like. | **Callable** |
| `classify/` | `Classifier` interface, `RuleSet` driver, `Rule` type (exact / regex matchers), and the worked `NaniteDeployRule` from the architecture doc. | **Callable** |
| `normalize/` | Canonicalization contract plus `SlugNormalizer` for conservative tags/slugs/names/paths. | **Callable** |
| `repair/` | Narrow deterministic repair contract plus `Chain` and `MissingClosingDelimiterJSON`. Never auto-repairs destructive commands; never auto-applies semantic-changing repairs. | **Callable** |
| `event/` | Normalized filter-event schema for downstream consumers. Separate from `go-runtime-events` because filter events describe pipeline output, not process activity. | Schema only |

## Architecture notes

- `chrispian/inbox/harness-filters-directives-normalization-2026-05-26.md`
- `chrispian/inbox/cli-runner-wrapper-architecture-2026-05-26.md`
- `chrispian/inbox/cli-wrapper-implementation-followups-2026-05-26.md`
  (handoff for the next session)

## Development

```sh
go test -race ./...   # tests
go vet ./...          # vet
gofmt -l .            # formatting check (no output = clean)
golangci-lint run     # lint
govulncheck ./...     # vulnerability scan
```

CI (`.github/workflows/check.yml`) runs the same checks on push and pull
request to `main`.

## License

MIT — see [LICENSE](./LICENSE).
