// Package repair defines the contract for narrowly-scoped, deterministic
// repairs of malformed envelopes, MCP tool calls, and JSON documents
// that agents commonly mis-emit.
//
// Per the harness-filters architecture note:
//
//   - Safe deterministic repairs may be applied automatically.
//   - The response/event MUST include Repaired=true, the original text,
//     and the repair reason.
//   - If a repair would change semantics, do NOT auto-repair — emit a
//     dev warning and preserve the raw content.
//   - NEVER auto-repair destructive commands.
//
// Concrete repair rules (envelope-language repair, missing-comma JSON
// repair, deprecated MCP tool-name aliasing) land per-domain in follow-up
// files; this package owns the contract.
package repair
