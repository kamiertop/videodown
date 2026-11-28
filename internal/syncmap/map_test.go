package syncmap

import (
	"sync"
	"testing"
)

func TestMapUpdateConcurrent(t *testing.T) {
	var counts Map[string, int]
	const workers = 8
	const updates = 100
	var wg sync.WaitGroup
	for range workers {
		wg.Go(func() {
			for range updates {
				counts.Update("total", func(current int, _ bool) (int, bool) {
					return current + 1, true
				})
			}
		})
	}
	wg.Wait()
	if got, ok := counts.Load("total"); !ok || got != workers*updates {
		t.Fatalf("Load(total) = %d, %t, want %d, true", got, ok, workers*updates)
	}
	if got, updated := counts.Update("total", func(int, bool) (int, bool) {
		return 0, false
	}); updated || got != workers*updates {
		t.Errorf("rejected Update(total) = %d, %t, want %d, false", got, updated, workers*updates)
	}
}

func TestMapStoresNilValues(t *testing.T) {
	var entries Map[string, any]
	entries.Store("nil", nil)
	if got, ok := entries.Load("nil"); got != nil || !ok {
		t.Errorf("Load after Store = %v, %t, want nil, true", got, ok)
	}
	entries.Delete("nil")
	if got, ok := entries.Load("nil"); got != nil || ok {
		t.Errorf("Load after Delete = %v, %t, want nil, false", got, ok)
	}
}
