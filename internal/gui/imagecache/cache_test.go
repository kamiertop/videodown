package imagecache

import (
	"fmt"
	"image"
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

func TestThumbnailBoundsRetainedPixels(t *testing.T) {
	for _, size := range []image.Point{{3840, 2160}, {2160, 3840}, {8000, 1}, {120, 80}} {
		src := image.NewRGBA(image.Rect(0, 0, size.X, size.Y))
		got := resizeThumbnail(src)
		bounds := got.Bounds()
		if bounds.Dx() > ThumbnailMaxSide || bounds.Dy() > ThumbnailMaxSide || bounds.Empty() {
			t.Fatalf("source %v produced invalid thumbnail %v", size, bounds)
		}
		if size.X <= ThumbnailMaxSide && size.Y <= ThumbnailMaxSide && got != src {
			t.Fatal("small images should not be enlarged or copied")
		}
		if size.X == 3840 && bounds.Size() != image.Pt(512, 288) {
			t.Fatalf("landscape aspect ratio changed: %v", bounds)
		}
		if size.Y == 3840 && bounds.Size() != image.Pt(288, 512) {
			t.Fatalf("portrait aspect ratio changed: %v", bounds)
		}
	}
}
