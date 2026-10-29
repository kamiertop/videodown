package imagecache

import (
	"fmt"
	"testing"
)

func TestCacheEvictsOldestAndKeepsLoadingNewImages(t *testing.T) {
	cache := Cache{}
	for i := range Capacity {
		cache.Add(fmt.Sprint(i))
	}
	kept := cache.Get("0")
	for i := Capacity; i < Capacity*2; i++ {
		key := fmt.Sprint(i)
		entry := cache.Add(key)
		if cache.Get(key) != entry || len(cache) != Capacity {
			t.Fatal("new image rejected at capacity")
		}
		if cache.Get("0") != kept {
			t.Fatal("recently used image evicted")
		}
	}
	if cache.Get("1") != nil {
		t.Fatal("old image retained")
	}
}
