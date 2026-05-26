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

The first concrete normalize/repair rules have landed:

- **`normalize/`** — `SlugNormalizer` handles conservative slugify,
  aliases, reserved guard, provenance, and optional simple singularize.
  Tesseract-specific taxonomy/collision behavior should layer on top.
- **`repair/`** — `MissingClosingDelimiterJSON` fixes obvious missing
  terminal `}` / `]` cases only when the repaired document validates.
  It is syntactic-only and does not add commas, quotes, keys, or values.

Still useful follow-ups:

- Wrong fenced-block language repair.
- Deprecated MCP tool-name aliasing.
- Missing-comma JSON repair, if we can keep it deterministic and
  semantics-preserving.
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

`go-agent-wrapper/wrapper.Config.Filters` is now invoked by the wrapper.
The wrapper also ships a `filters.RepairPipeline` adapter that composes
`go-harness-filters/repair` repairers. A richer implementation could:

- Wrap one or more `Repairer`s for envelope/JSON content.
- Wrap one or more `Normalizer`s for tag/slug content.
- Wrap a `Classifier` (often the same one the wrapper passes to
  `Policy` via `classifybridge`).
- Emit per-event metadata back to the wrapper.

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
