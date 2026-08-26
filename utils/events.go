package utils

// EventEmitter sends backend events to the frontend. Runtime-specific
// implementations are supplied by the application layer.
type EventEmitter interface {
	EmitEvent(name string, data any) bool
}
