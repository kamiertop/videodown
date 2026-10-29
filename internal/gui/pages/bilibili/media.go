package bilibili

import (
	"fmt"
	"strings"
	"time"

	"github.com/ELY-Open-Source/MujicaUI/core"
	"github.com/egoist/mygo/ui"
	"github.com/kamiertop/videodown/internal/gui/navigation"
	"github.com/kamiertop/videodown/internal/gui/style"
)

func (s *State) selection(key string) *bool {
	if s.selected[key] == nil {
		s.selected[key] = new(bool)
	}
	return s.selected[key]
}
func matches(title, query string) bool {
	return strings.Contains(strings.ToLower(title), strings.ToLower(strings.TrimSpace(query)))
}
func (s *State) columns(c *ui.Context, sidebar bool) int {
	width, _ := c.Size()
	width -= 140
	if sidebar {
		width -= 176
	}
	return max(1, min(5, int(width/230)))
}

// Callers retain the typed API records; these arguments only describe the card.
func (s *State) mediaCard(c *ui.Context, bvid, title, cover, author, length string, play, danmaku, published int) {
	s.mediaCardWithAuthor(c, bvid, title, cover, author, length, play, danmaku, published, 0, nil)
}

// The card selects a video, while the author name is a separate navigation target.
func (s *State) mediaCardWithAuthor(c *ui.Context, bvid, title, cover, author, length string, play, danmaku, published int, authorID uint64, navigate func(string)) {
	s.mediaCardWithSelection(c, bvid, title, cover, author, length, play, danmaku, published, authorID, navigate, s.selection(bvid))
}

func (s *State) mediaCardWithSelection(c *ui.Context, bvid, title, cover, author, length string, play, danmaku, published int, authorID uint64, navigate func(string), selection *bool) {
	k := core.Tokens(c)
	valid := strings.HasPrefix(bvid, "BV")
	linkedAuthor := authorID > 0 && strings.TrimSpace(author) != "" && navigate != nil
	selected := valid && *selection
	frameColor := k.Border.Alpha(.7)
	card := ui.ButtonBase(c).Label("选择视频：" + title).Column().AlignSelf(ui.Stretch).MinWidth(0).
		Clip().Radius(style.RadiusMedia).
		Background(k.Surface).TextColor(k.Text).Disabled(!valid && !linkedAuthor)
	if valid {
		card.Cursor(ui.CursorPointer)
	}
	authorHovered := false
	card.Children(func() {
		s.mediaThumbnail(c, cover, length, play, danmaku)
		ui.Column(c).FillWidth().MinWidth(0).Padding(style.SpaceSmall).Gap(style.SpaceSmall).Children(func() {
			ui.Text(c, title).FillWidth().MinHeight(style.MediaTitleHeight).
				FontWeight(style.WeightSemibold).MaxLines(2)
			ui.Row(c).FillWidth().MinWidth(0).Gap(style.SpaceSmall).Children(func() {
				if linkedAuthor {
					ui.Row(c).Grow(1).MinWidth(0).Children(func() {
						link := ui.ButtonBase(c.Key("author-link")).Label("进入 UP 主主页："+author).
							MinWidth(0).MaxWidthPercent(100).Padding(0, 0, style.MediaAuthorUnderlineGap, 0).
							BorderWidth(0, 0, style.MediaAuthorUnderlineWidth, 0).BorderColor(ui.Transparent).
							Background(ui.Transparent).TextColor(k.TextMuted).Cursor(ui.CursorPointer)
						authorHovered = link.Hovered() || link.FocusVisible()
						if authorHovered {
							link.TextColor(k.AccentText).BorderColor(k.Accent)
						}
						link.Children(func() { ui.Text(c, author).SingleLine() }).
							OnClick(func() { navigate(navigation.BilibiliAuthorPath(authorID)) })
					})
				} else {
					ui.Text(c, author).Grow(1).MinWidth(0).SingleLine().TextColor(k.TextMuted)
				}
				if published > 0 {
					ui.Text(c, time.Unix(int64(published), 0).Format("2006-01-02")).
						Shrink(0).SingleLine().TextColor(k.TextMuted)
				}
			})
		})
	})
	if selected {
		frameColor = k.Accent
		card.Background(k.Selection).
			Shadow(0, 2, 10, 0, k.Accent.Alpha(.2))
	} else if valid && !authorHovered && (card.Hovered() || card.FocusVisible()) {
		frameColor = k.Accent.Alpha(.7)
		card.Background(k.SurfaceHover).
			Shadow(0, 2, 10, 0, k.Accent.Alpha(.14))
	}
	mediaOutline(card, frameColor, style.RadiusMedia, style.BorderEmphasis).
		OnClick(func() {
			if valid {
				*selection = !*selection
			}
		})
}
func duration(seconds int) string { return fmt.Sprintf("%d:%02d", seconds/60, seconds%60) }
func (s *State) stagingNotice(c *ui.Context) {
	if s.notice != "" {
		caption(c, s.notice)
	}
}
