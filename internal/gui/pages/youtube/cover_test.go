package youtube

import (
	"context"
	"errors"
	"image"
	"testing"

	"github.com/egoist/mygo/ui"
	"github.com/kamiertop/videodown/internal/gui/guiruntime"
)

func TestCoverFallbackAndDownloadDoNotDiscardImage(t *testing.T) {
	bitmap, err := loadCoverWith(context.Background(), []string{"missing", "image"}, func(ctx context.Context, url, referer string) (*ui.Bitmap, error) {
		if url == "missing" {
			return nil, errors.New("404")
		}
		return ui.NewBitmap(image.NewRGBA(image.Rect(0, 0, 16, 9))), nil
	})
	if err != nil || bitmap == nil {
		t.Fatalf("cover fallback: %v", err)
	}
	s := NewState()
	s.Configure(&fakeAPI{}, guiruntime.Runtime{Background: func(work func() error, done func(error)) { done(work()) }}, func(fn func()) { fn() })
	s.resolve()
	generation := s.coverGeneration
	s.download()
	s.finishCover(generation, bitmap, nil)
	if s.cover == nil {
		t.Fatal("starting download discarded pending cover")
	}
}
