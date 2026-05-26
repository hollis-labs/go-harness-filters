// Package directive defines the directive syntax agents can emit when
// prose is the wrong interface but a full tool call is too heavy.
//
// Examples (from the harness filters architecture note):
//
//   @nanite:recover-session short_code=c287 history=last_50
//   @torque:block task=CW-20260525-0001 reason="missing context after reboot"
//   @fragment:capture kind=decision title="Use Cerberus for Nanite deploys"
//   @stack-explorer:index path=apps/nanite include=go,ts,tsx
//   @hadron:record-flow name="admin agent CRUD"
//
// Parsing is deterministic-first. LLM classification is reserved for
// ambiguous prose; the parser in this package only handles the
// structured @namespace:action key=value form.
//
// High-risk directives are NEVER silently executed. This package
// produces a [Directive] value with namespace, action, typed args, and
// provenance; downstream policy decides whether to run, gate on
// approval, or block.
package directive
