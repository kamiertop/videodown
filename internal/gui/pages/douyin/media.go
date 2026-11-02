package douyin

import (
	"net/url"
	"time"

	"github.com/ELY-Open-Source/MujicaUI/core"
	"github.com/egoist/mygo/ui"
	"github.com/kamiertop/videodown/douyin/model"
	"github.com/kamiertop/videodown/internal/gui/assets"
	"github.com/kamiertop/videodown/internal/gui/style"
)

func (s *State) videoCard(c *ui.Context, v model.AwemeItem, navigate func(string)) {
	k := core.Tokens(c)
	card := ui.ButtonBase(c).Label("选择抖音作品：" + title(v)).Column().AlignSelf(ui.Stretch).MinWidth(0).Clip().Radius(style.RadiusMedia).Background(k.Surface).TextColor(k.Text).FocusRing(true)
	authorHovered := false
	card.Children(func() {
		ui.Box(c).Label("抖音作品封面").FillWidth().AspectRatio(style.DouyinMediaAspectRatio).Shrink(0).Clip().Children(func() {
			s.imageWithAspect(c, cover(v), 0, 0, false, style.RadiusMedia, style.DouyinMediaAspectRatio)
			if v.IsTop != 0 {
				ui.Text(c, "置顶").Absolute().Start(style.SpaceSmall).Top(style.SpaceSmall).Padding(3, 6).Radius(style.RadiusChrome).Background(k.Selection).TextColor(k.AccentText).FontSize(style.Font(c, style.FontMeta))
			}
			livePhoto := v.IsLivePhoto != 0 || v.IsSliedes
			for _, image := range v.Images {
				if image.ClipType == 5 || image.LivePhotoType != 0 || first(image.Video.PlayAddr.UrlList) != "" {
					livePhoto = true
					break
				}
			}
			if livePhoto || len(v.Images) > 0 {
				icon, label := assets.DouyinImage, "图文"
				if livePhoto {
					icon, label = assets.DouyinLivePhoto, "动图"
				}
				ui.Box(c).Label(label).Absolute().End(style.SpaceSmall).Top(style.SpaceSmall).Size(style.DouyinMediaBadgeSize, style.DouyinMediaBadgeSize).Radius(style.RadiusControl).Background(ui.RGB(0, 0, 0).Alpha(.55)).Border(1, ui.RGB(255, 255, 255).Alpha(.2)).Center().Children(func() {
					ui.Icon(c, icon).Size(style.DouyinMediaBadgeIconSize, style.DouyinMediaBadgeIconSize).TextColor(ui.RGB(255, 255, 255))
				})
			} else {
				ui.Text(c, duration(videoDuration(v))).Absolute().End(style.SpaceSmall).Bottom(style.SpaceSmall).FontSize(style.Font(c, style.MediaDurationFontSize)).Padding(style.MediaDurationPaddingY, style.MediaDurationPaddingX).Radius(style.RadiusChrome).Background(ui.RGB(0, 0, 0).Alpha(style.MediaDurationOpacity)).TextColor(ui.RGB(255, 255, 255))
			}
		})
		ui.Column(c).FillWidth().MinWidth(0).Padding(style.SpaceSmall).Gap(style.DouyinMediaTextGap).Shrink(0).Children(func() {
			ui.Text(c, title(v)).FillWidth().MinHeight(style.DouyinMediaTitleHeight).Shrink(0).MaxLines(2).FontWeight(style.WeightSemibold).FontSize(style.Font(c, style.FontMeta))
			ui.Column(c).FillWidth().MinWidth(0).Gap(style.DouyinMediaMetaGap).Children(func() {
				ui.Row(c).FillWidth().MinWidth(0).Children(func() {
					if v.Author.SecUid != "" && navigate != nil {
						link := ui.ButtonBase(c.Key("author")).Label("进入抖音作者主页："+v.Author.Nickname).MinWidth(0).MaxWidthPercent(100).Padding(0, 0, style.MediaAuthorUnderlineGap, 0).BorderWidth(0, 0, style.MediaAuthorUnderlineWidth, 0).BorderColor(ui.Transparent).Background(ui.Transparent).TextColor(k.TextMuted).Cursor(ui.CursorPointer)
						authorHovered = link.Hovered() || link.FocusVisible()
						if authorHovered {
							link.TextColor(k.AccentText).BorderColor(k.Accent)
						}
						link.Children(func() { ui.Text(c, "@"+v.Author.Nickname).SingleLine().FontSize(style.Font(c, style.FontMeta)) }).OnClick(func() { navigate("/douyin/user/" + url.PathEscape(v.Author.SecUid)) })
					} else {
						caption(c, v.Author.Nickname)
					}
				})
				if v.CreateTime > 0 {
					ui.Text(c, "发布 "+time.Unix(int64(v.CreateTime), 0).Format("2006-01-02")).FontSize(style.Font(c, style.FontMeta)).Shrink(0).SingleLine().TextColor(k.TextMuted)
				}
			})
		})
	})
	color := k.Border
	if s.selected[v.AwemeId] {
		color = k.Accent
		card.Background(k.Selection)
	} else if !authorHovered && (card.Hovered() || card.FocusVisible()) {
		color = k.Accent
		card.Background(k.SurfaceHover)
	}
	// Paint the outline without insetting the image or shifting layout on selection.
	card.DrawOver(func(p *ui.Painter, r ui.Rect) { p.Stroke(r, color, style.RadiusMedia, style.BorderEmphasis) }).OnClick(func() { s.selected[v.AwemeId] = !s.selected[v.AwemeId] })
}

// Keep portrait cards readable as the available grid width changes.
func videoColumns(width float32) int {
	for _, breakpoint := range []struct {
		width   float32
		columns int
	}{{1500, 8}, {1240, 7}, {1040, 6}, {840, 5}, {620, 4}, {420, 3}} {
		if width >= breakpoint.width {
			return breakpoint.columns
		}
	}
	return 2
}
