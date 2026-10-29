package douyin

import (
	"strings"
	"time"

	"github.com/ELY-Open-Source/MujicaUI/core"
	"github.com/ELY-Open-Source/MujicaUI/icons"
	"github.com/egoist/mygo/ui"
	"github.com/kamiertop/videodown/internal/gui/imagecache"
	"github.com/kamiertop/videodown/internal/gui/style"
)

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
	frame := ui.Box(c).Shrink(0).Clip().Radius(radius).Background(core.Tokens(c).Background).Center()
	if width == 0 {
		frame.FillWidth().AspectRatio(aspect)
	} else {
		frame.Size(width, height)
	}
	// Only request images intersecting the window; long grids build offscreen cards too.
	bounds := frame.Bounds()
	windowWidth, windowHeight := c.Size()
	visible := bounds.W > 0 && bounds.H > 0 && bounds.X < windowWidth && bounds.Y < windowHeight && bounds.X+bounds.W > 0 && bounds.Y+bounds.H > 0
	if bounds.W == 0 || bounds.H == 0 {
		c.After(time.Millisecond)
	}
	entry := s.images.Get(url)
	if entry == nil && visible && url != "" && s.runtime.Image != nil && !s.stopped {
		entry = s.images.Add(url)
		ctx := s.ctx
		target := entry
		epoch := s.epoch
		var bitmap *ui.Bitmap
		s.runtime.Image(func() error {
			var err error
			bitmap, err = imagecache.Load(ctx, url, "https://www.douyin.com/")
			return err
		}, func(err error) {
			if s.stopped || epoch != s.epoch {
				return
			}
			if err == nil {
				target.Bitmap = bitmap
			}
		})
	}
	frame.Children(func() {
		if entry != nil && entry.Bitmap != nil {
			ui.Image(c, entry.Bitmap).Fill().Fit(ui.Cover)
		} else {
			name := "image"
			if avatar {
				name = "user"
			}
			ui.Icon(c, icons.Must(name)).Size(24, 24).TextColor(core.Tokens(c).TextMuted)
		}
	})
}
