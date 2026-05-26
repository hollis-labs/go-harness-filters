// Package event defines the normalized event types the filter pipeline
// emits for downstream consumers.
//
// Filter events are intentionally separate from the runtime activity
// events in github.com/hollis-labs/go-runtime-events. The runtime
// schema describes "what the wrapped process did"; this package
// describes "what the filter pipeline observed, classified, or
// repaired about that activity". Apps that want to surface both into
// a single UI can map filter events to runtime events at the boundary.
//
// Concrete emission helpers and sink wiring land in follow-up files;
// this package owns the schema.
package event
