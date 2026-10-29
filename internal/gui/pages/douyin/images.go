package douyin

import (
	"bytes"
	"errors"
	"image"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/ELY-Open-Source/MujicaUI/core"
	"github.com/ELY-Open-Source/MujicaUI/icons"
	"github.com/egoist/mygo/ui"
	"github.com/kamiertop/videodown/internal/gui/style"
)

var imageClient = &http.Client{Timeout: 12 * time.Second}

func (s *State) image(c *ui.Context, url string, width, height float32, avatar bool) {
	radius := style.RadiusMedia
	if avatar {
		radius = width / 2
	}
	s.imageWithRadius(c, url, width, height, avatar, radius)
}

func (s *State) imageWithRadius(c *ui.Context, url string, width, height float32, avatar bool, radius float32) {
	s.imageWithAspect(c, url, width, height, avatar, radius, style.MediaAspectRatio)
}
func (s *State) imageWithAspect(c *ui.Context, url string, width, height float32, avatar bool, radius, aspect float32) {
	if strings.HasPrefix(url, "//") {
		url = "https:" + url
	}
	entry := s.images[url]
	if entry == nil && url != "" && s.runtime.Image != nil && !s.stopped && len(s.images) < 256 {
		entry = &imageEntry{loading: true}
		s.images[url] = entry
		ctx := s.ctx
		target := entry
		epoch := s.epoch
		var bitmap *ui.Bitmap
		s.runtime.Image(func() error {
			req, err := http.NewRequest(http.MethodGet, url, nil)
			if err != nil {
				return err
			}
			if ctx != nil {
				req = req.WithContext(ctx)
			}
			req.Header.Set("Referer", "https://www.douyin.com/")
			resp, err := imageClient.Do(req)
			if err != nil {
				return err
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				return errors.New("图片加载失败")
			}
			data, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
			if err != nil {
				return err
			}
			config, _, err := image.DecodeConfig(bytes.NewReader(data))
			if err != nil {
				return err
			}
			if config.Width <= 0 || config.Height <= 0 || config.Width > 8192 || config.Height > 8192 || int64(config.Width)*int64(config.Height) > 16<<20 {
				return errors.New("图片尺寸过大")
			}
			bitmap, err = ui.DecodeBitmap(data)
			return err
		}, func(err error) {
			if s.stopped || epoch != s.epoch {
				return
			}
			target.loading = false
			if err == nil {
				target.bitmap = bitmap
			}
		})
	}
	frame := ui.Box(c).Shrink(0).Clip().Radius(radius).Background(core.Tokens(c).Background).Center()
	if width == 0 {
		frame.FillWidth().AspectRatio(aspect)
	} else {
		frame.Size(width, height)
	}
	frame.Children(func() {
		if entry != nil && entry.bitmap != nil {
			ui.Image(c, entry.bitmap).Fill().Fit(ui.Cover)
		} else {
			name := "image"
			if avatar {
				name = "user"
			}
			ui.Icon(c, icons.Must(name)).Size(24, 24).TextColor(core.Tokens(c).TextMuted)
		}
	})
}
