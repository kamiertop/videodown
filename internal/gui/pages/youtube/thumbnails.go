package youtube

import (
	"time"

	"github.com/ELY-Open-Source/MujicaUI/core"
	"github.com/ELY-Open-Source/MujicaUI/icons"
	"github.com/ELY-Open-Source/MujicaUI/input"
	"github.com/egoist/mygo/ui"
	"github.com/kamiertop/videodown/internal/gui/components"
	"github.com/kamiertop/videodown/internal/gui/style"
	yt "github.com/kamiertop/videodown/youtube"
)

type thumbnail struct {
	Bitmap           *ui.Bitmap
	Loading          bool
	Error            string
	Generation, Used uint64
}

func (s *State) thumbnailEntry(video yt.Video) *thumbnail {
	if s.thumbnails == nil {
		s.thumbnails = map[string]*thumbnail{}
	}
	entry := s.thumbnails[video.ID]
	if entry == nil {
		if len(s.thumbnails) >= 256 {
			oldest, used := "", ^uint64(0)
			for id, item := range s.thumbnails {
				if item.Used < used {
					oldest, used = id, item.Used
				}
			}
			delete(s.thumbnails, oldest)
		}
		entry = &thumbnail{}
		s.thumbnails[video.ID] = entry
	}
	s.thumbnailUse++
	entry.Used = s.thumbnailUse
	return entry
}

func (s *State) requestThumbnail(video yt.Video, entry *thumbnail) {
	if s.stopped || s.runtime.Image == nil || entry.Loading || video.ID == "" {
		return
	}
	entry.Loading, entry.Error = true, ""
	entry.Generation++
	generation, ctx := entry.Generation, s.ctx
	loader := s.thumbnailLoader
	if loader == nil {
		loader = loadCover
	}
	urls := []string{video.Cover, "https://i.ytimg.com/vi/" + video.ID + "/hqdefault.jpg", "https://img.youtube.com/vi/" + video.ID + "/hqdefault.jpg"}
	var bitmap *ui.Bitmap
	s.runtime.Image(func() error { var err error; bitmap, err = loader(ctx, urls); return err }, func(err error) {
		if s.stopped || entry.Generation != generation {
			return
		}
		entry.Loading = false
		if err != nil {
			entry.Error = "封面加载失败，点击重试"
		} else {
			entry.Bitmap = bitmap
		}
		if s.video.ID == video.ID {
			s.finishCover(s.coverGeneration, bitmap, err)
		}
	})
}

// Source rows and queue groups share one bounded cache. Offscreen rows and
// collapsed group children do not request images.
func (s *State) thumbnailFrame(c *ui.Context, video yt.Video, width, height float32, label string) {
	k := core.Tokens(c)
	frame := ui.Box(c).Label(label).Size(width, height).Shrink(0).Clip().Radius(style.RadiusMedia).Background(k.Background).Center()
	entry := s.thumbnails[video.ID]
	if frame.Bounds().W == 0 {
		c.After(time.Millisecond)
	}
	if frame.InView() {
		entry = s.thumbnailEntry(video)
		if entry.Bitmap == nil && entry.Error == "" && !entry.Loading {
			s.requestThumbnail(video, entry)
		}
	}
	frame.Children(func() {
		switch {
		case entry != nil && entry.Bitmap != nil:
			ui.Image(c, entry.Bitmap).Fill().Fit(ui.Cover)
		case entry != nil && entry.Loading:
			core.Spinner(c, core.SpinnerOptions{Size: 16, Label: "正在加载封面"})
		case entry != nil && entry.Error != "":
			components.Button(c.Key("retry"), "", input.ButtonOptions{Variant: input.Ghost, Icon: icons.Must("refresh-cw"), Label: "重试封面：" + video.ID}).Size(28, 28).MinWidth(28).Padding(0).Tooltip(entry.Error).OnClick(func() { s.requestThumbnail(video, entry) })
		default:
			ui.Icon(c, icons.Must("image")).Size(20, 20).TextColor(k.TextMuted)
		}
	})
}
