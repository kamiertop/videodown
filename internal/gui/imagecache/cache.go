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
	"github.com/imroc/req/v3"
	"github.com/kamiertop/videodown/internal/netclient"
	"golang.org/x/image/draw"
)

var imageClient = netclient.New("", 12*time.Second, nil)

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
	return load(ctx, imageClient, url, referer, false)
}

// ThumbnailMaxSide bounds retained RGBA pixels to 1 MiB per card image.
const ThumbnailMaxSide = 512

// LoadThumbnail keeps card caches independent of the source image resolution.
func LoadThumbnail(ctx context.Context, url, referer string) (*ui.Bitmap, error) {
	return load(ctx, imageClient, url, referer, true)
}

func load(ctx context.Context, client *req.Client, url, referer string, thumbnail bool) (*ui.Bitmap, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	resp, err := client.R().SetContext(ctx).SetHeader("Referer", referer).Get(url)
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
	if !thumbnail {
		return ui.DecodeBitmap(data)
	}
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	return ui.NewBitmap(orientThumbnail(resizeThumbnail(img), jpegOrientation(data))), nil
}

func resizeThumbnail(img image.Image) image.Image {
	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	longest := max(w, h)
	if longest <= ThumbnailMaxSide {
		return img
	}
	w, h = max(1, w*ThumbnailMaxSide/longest), max(1, h*ThumbnailMaxSide/longest)
	dst := image.NewRGBA(image.Rect(0, 0, w, h))
	draw.BiLinear.Scale(dst, dst.Bounds(), img, bounds, draw.Src, nil)
	return dst
}

// NearViewport includes a small look-ahead region without loading the whole grid.
func NearViewport(bounds ui.Rect, width, height float32) bool {
	const margin float32 = 512
	return bounds.W > 0 && bounds.H > 0 && bounds.X < width && bounds.X+bounds.W > 0 && bounds.Y < height+margin && bounds.Y+bounds.H > -margin
}

// LoadThumbnailWithClient lets platform pages use their configured proxy.
func LoadThumbnailWithClient(ctx context.Context, client *req.Client, url, referer string) (*ui.Bitmap, error) {
	return load(ctx, client, url, referer, true)
}
