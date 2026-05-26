# go-harness-filters Roadmap

Status as of v0.1.0 (2026-05-26). See
[CHANGELOG.md](./CHANGELOG.md) for what landed.

## Publish blockers

- None internal to this module — `go-harness-filters` has no Hollis
  Labs dependencies. Tag after `go-runtime-events` and before
  `go-agent-wrapper`.
- Standard pre-tag polish: `examples/` is empty.

## Deferred this pass

### Concrete rule sets

Each subpackage with "contract only" status needs real implementations
to be useful end-to-end. The contracts are stable; the rule sets are
what consumers will reach for.

- **`normalize/`** — Tesseract taxonomy normalizer is the first
  expected consumer. Likely shape: a `Normalizer` impl that applies
  the layers from the architecture note (Unicode fold → slugify →
  singularize via `gobuffalo/flect` or similar → alias map → reserved
  guard → collision handling).
- **`repair/`** — Envelope repairs are the highest-value first cut
  (malformed JSON, wrong fenced-block language, deprecated MCP tool
  names). The architecture doc lists the targets; deterministic ones
  ship first, anything that changes semantics stays off-by-default.
- **`event/`** — Schema only today. A reference emitter helper +
  fan-out sink (analogous to `runtimeevents.MultiSink`) would let
  filter consumers wire up easily. Defer until at least two consumers
  need it.

### More classify rules

Only `NaniteDeployRule` ships today. Other architecture-doc-aligned
candidates:

- `git.force-push.main` — block force-push to main/master without
  operator opt-in.
- `filesystem.delete-tree` — flag `rm -rf` against project roots.
- `mcp.deprecated-tool` — alias old MCP tool names to current ones
  (probably belongs in `repair/` rather than `classify/`).

Drop these as rules accumulate — keep the rule registry organized by
domain (`hollis.deploy.*`, `git.*`, `filesystem.*`, etc.).

### Filter pipeline integration

`go-agent-wrapper/wrapper.Config.Filters` is a `filters.Pipeline` field
the wrapper exposes but doesn't invoke yet. A real `Pipeline`
implementation would:

- Wrap one or more `Repairer`s for envelope/JSON content.
- Wrap one or more `Normalizer`s for tag/slug content.
- Wrap a `Classifier` (often the same one the wrapper passes to
  `Policy` via `classifybridge`).
- Emit per-event metadata back to the wrapper.

That integration belongs in the wrapper when there's a real consumer.

## Open design questions

1. **Policy boundary.** The companion architecture note lists `policy/`
   as a subpackage here too, overlapping with
   `go-agent-wrapper/policy/`. Filter consumers (Nanite, Torque,
   Tether) may want filter-driven policy without pulling in the
   wrapper. Three options:
   - Extract `Mode`/`Decision`/`Rule` to a third-party module
     (`go-policy-decisions`?).
   - Duplicate the types here with explicit "kept in sync" docstrings.
   - Accept the wrapper dependency on filter consumers that need
     decisions.

   Revisit once a non-wrapper consumer actually needs policy decisions.

2. **Classifier rule-loading source.** Rules ship as Go values
   (`NaniteDeployRule`). Should we accept rule definitions from YAML/JSON
   config too? The architecture doc hints at it ("Should policy rules
   be YAML/JSON config, database-backed, or both?"). Defer until a
   consumer actually wants dynamic rule loading.

3. **Directive registry / dispatch.** The parser produces a structured
   `Directive` but does not execute it. The architecture note explicitly
   forbids silent execution of high-risk directives. A future
   `directive/` extension could ship a `Registry` mapping
   `(namespace, action)` to handlers with policy gating. Skipped until
   the directive feature has a concrete first consumer.

## Related docs

- Original architecture: `chrispian/inbox/harness-filters-directives-normalization-2026-05-26.md`
- Companion wrapper doc: `chrispian/inbox/cli-runner-wrapper-architecture-2026-05-26.md`
- Next-session handoff: `chrispian/inbox/cli-wrapper-implementation-followups-2026-05-26.md`
