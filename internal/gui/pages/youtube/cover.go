package youtube

import (
	"context"
	"errors"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/kamiertop/videodown/internal/gui/imagecache"
)

func (s *State) loadCover() {
	if s.video.ID == "" {
		return
	}
	entry := s.thumbnailEntry(s.video)
	if entry.Bitmap == nil && !entry.Loading {
		s.requestThumbnail(s.video, entry)
	}
	s.cover, s.coverLoading, s.coverError = entry.Bitmap, entry.Loading, entry.Error
}

func (s *State) finishCover(generation uint64, bitmap *ui.Bitmap, err error) {
	if s.stopped || s.coverGeneration != generation {
		return
	}
	s.coverLoading = false
	if err != nil {
		s.coverError = "封面加载失败，点击重试"
		return
	}
	s.cover = bitmap

}

func loadCover(parent context.Context, urls []string) (*ui.Bitmap, error) {
	return loadCoverWith(parent, urls, imagecache.Load)
}

func loadCoverWith(parent context.Context, urls []string, load func(context.Context, string, string) (*ui.Bitmap, error)) (*ui.Bitmap, error) {
	ctx, cancel := context.WithTimeout(parent, 12*time.Second)
	defer cancel()
	type result struct {
		image *ui.Bitmap
		err   error
	}
	results := make(chan result, len(urls))
	seen := map[string]bool{}
	count := 0
	for _, url := range urls {
		if url == "" || seen[url] {
			continue
		}
		seen[url] = true
		count++
		go func() {
			bitmap, err := load(ctx, url, "https://www.youtube.com/")
			results <- result{bitmap, err}
		}()
	}
	var last error = errors.New("没有封面地址")
	for range count {
		select {
		case result := <-results:
			if result.err == nil {
				return result.image, nil
			}
			last = result.err
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	return nil, last
}
