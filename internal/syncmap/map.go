// Package syncmap provides a typed wrapper around sync.Map.
package syncmap

import "sync"

// Map is safe for concurrent use. Its zero value is ready to use.
// A Map must not be copied after first use.
type Map[K comparable, V any] struct {
	entries sync.Map
}

// Load returns the value for key, or the zero value if key is absent.
func (m *Map[K, V]) Load(key K) (V, bool) {
	entry, ok := m.entries.Load(key)
	if !ok {
		var zero V
		return zero, false
	}
	return *entry.(*V), true
}

// Store sets the value for key.
func (m *Map[K, V]) Store(key K, value V) {
	m.entries.Store(key, &value)
}

// Delete removes key.
func (m *Map[K, V]) Delete(key K) {
	m.entries.Delete(key)
}

// Update atomically replaces the value for key if next returns true.
// It returns the new value on success or the observed value on rejection.
// Under contention next can run more than once, so it must have no side effects.
func (m *Map[K, V]) Update(key K, next func(current V, loaded bool) (V, bool)) (V, bool) {
	for {
		entry, loaded := m.entries.Load(key)
		var current V
		if loaded {
			current = *entry.(*V)
		}
		value, replace := next(current, loaded)
		if !replace {
			return current, false
		}
		if loaded {
			if m.entries.CompareAndSwap(key, entry, &value) {
				return value, true
			}
		} else if _, loaded := m.entries.LoadOrStore(key, &value); !loaded {
			return value, true
		}
	}
}
