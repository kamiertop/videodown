package imagecache

import (
	"bytes"
	"context"
	"errors"
	"image"
	"io"
	"net/http"
	"time"

	"github.com/egoist/mygo/ui"
)

var imageClient = &http.Client{Timeout: 12 * time.Second}

// Cache is owned by the UI thread. Workers only return decoded bitmaps.
type Cache map[string]*Entry
type Entry struct {
	Bitmap *ui.Bitmap
	used   uint64
}

const Capacity = 256

func (c Cache) Get(url string) *Entry {
	entry := c[url]
	if entry != nil {
		var newest uint64
		for _, item := range c {
			newest = max(newest, item.used)
		}
		entry.used = newest + 1
	}
	return entry
}
func (c Cache) Add(url string) *Entry {
	var newest uint64
	var oldest string
	var oldestUse uint64 = ^uint64(0)
	for key, item := range c {
		newest = max(newest, item.used)
		if item.used < oldestUse {
			oldest, oldestUse = key, item.used
		}
	}
	if len(c) >= Capacity {
		delete(c, oldest)
	}
	entry := &Entry{used: newest + 1}
	c[url] = entry
	return entry
}

func Load(ctx context.Context, url, referer string) (*ui.Bitmap, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	if ctx != nil {
		req = req.WithContext(ctx)
	}
	req.Header.Set("Referer", referer)
	resp, err := imageClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("图片加载失败")
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return nil, err
	}
	config, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	if config.Width <= 0 || config.Height <= 0 || config.Width > 8192 || config.Height > 8192 || int64(config.Width)*int64(config.Height) > 16<<20 {
		return nil, errors.New("图片尺寸过大")
	}
	return ui.DecodeBitmap(data)
}
