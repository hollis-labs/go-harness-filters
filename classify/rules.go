package classify

import "regexp"

// NaniteDeployRule matches the worked example from the CLI runner /
// wrapper architecture note: an agent invoking `go build` against the
// Nanite cmd path is almost certainly trying to deploy a new Nanite
// binary by hand, which would produce a stale artifact divergent from
// the Cerberus-managed service.
//
// The recommended commands deploy through Cerberus instead. Note that
// Reversible is false — the original (a local build) and the recommended
// (a production deploy) do materially different things; the policy
// layer should NUDGE so the agent sees the recommendation but keeps
// owning the choice. REWRITE would silently convert a debug build into
// a production deploy.
//
// How the policy layer consumes this Match:
//
//   - Match.RuleID → [github.com/hollis-labs/go-agent-wrapper/policy.Decision.RuleID]
//   - Match.Recommended → joined into [Decision.Message] (for nudge)
//     or [Decision.Replacement] (for rewrite, when operator-allowed)
//   - Match.Reversible → policy default mode: false → nudge, true → rewrite
//
// The bridge from classify.Match to policy.Decision lives in the wrapper
// (it depends on session/turn context the filter library is unaware
// of); this rule provides the upstream signal.
var NaniteDeployRule = Rule{
	ID:     "hollis.deploy.nanite.cerberus-required",
	Intent: "deploy.nanite",
	// Matches `go build ... cmd/nanite ...` with any flag ordering.
	// `\b` word boundaries anchor "go build" and "cmd/nanite" so we
	// don't catch `go test ./cmd/nanite` or substrings like `algo` /
	// `gocmd`, but DO catch the command embedded in JSON-encoded
	// tool_use payloads (`{"command":"go build -o nanite ./cmd/nanite"}`)
	// since `"` is a non-word char and `\b` accepts that boundary.
	RegexpMatch: regexp.MustCompile(`\bgo\s+build\b[^\n]*\bcmd/nanite\b`),
	Confidence:  ConfidenceExact,
	Recommended: []string{
		"cerberus_resource_deploy nanite-api-service",
		"cerberus_resource_reload nanite-api-service",
	},
	Reversible: false,
}
