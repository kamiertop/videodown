package bilibili

import (
	"fmt"

	"github.com/egoist/mygo/ui"
	"github.com/kamiertop/videodown/internal/gui/style"
)

// Native eye and chat icon paths.
var (
	mediaEyeIcon = ui.MustParseSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor">
		<path d="M10 12a2 2 0 100-4 2 2 0 000 4z"/>
		<path fill-rule="evenodd" clip-rule="evenodd" d="M.458 10C1.732 5.943 5.522 3 10 3s8.268 2.943 9.542 7c-1.274 4.057-5.064 7-9.542 7S1.732 14.057.458 10zM14 10a4 4 0 11-8 0 4 4 0 018 0z"/>
	</svg>`))
	mediaChatIcon = ui.MustParseSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor">
		<path d="M2 5a2 2 0 012-2h7a2 2 0 012 2v4a2 2 0 01-2 2H9l-3 3v-3H4a2 2 0 01-2-2V5z"/>
		<path d="M15 7v2a4 4 0 01-4 4H9.828l-1.766 1.767c.28.149.599.233.938.233h2l3 3v-3h2a2 2 0 002-2V9a2 2 0 00-2-2h-1z"/>
	</svg>`))
)

// A painted outline leaves the cover flush with the card edge. A normal
// Element.Border would inset the cover by its width and reveal a thin gap.
func mediaOutline(e ui.Element, color ui.Color, radius, width float32) ui.Element {
	return e.DrawOver(func(p *ui.Painter, r ui.Rect) {
		p.Stroke(r, color, radius, width)
	})
}

func (s *State) mediaThumbnail(c *ui.Context, cover, length string, play, danmaku int) {
	ui.Box(c).FillWidth().AspectRatio(style.MediaAspectRatio).Clip().Children(func() {
		// The image and outer card share a radius, so dark covers cannot leave
		// square pixels at the upper corners.
		s.imageWithRadius(c, cover, 0, 0, false, style.RadiusMedia)
		ui.Row(c).Absolute().Start(0).End(0).Bottom(0).
			Padding(style.MediaOverlayPaddingY, style.MediaOverlayPaddingX).
			Gap(style.SpaceSmall).AlignItems(ui.Center).Children(func() {
			if play > 0 || danmaku > 0 {
				ui.Row(c).Grow(1).Gap(style.SpaceSmall).AlignItems(ui.Center).Children(func() {
					mediaStat(c, mediaEyeIcon, play)
					mediaStat(c, mediaChatIcon, danmaku)
				})
			} else {
				ui.Box(c).Grow(1)
			}
			if length != "" {
				ui.Text(c, length).Shrink(0).SingleLine().
					FontSize(style.Font(c, style.MediaDurationFontSize)).
					Padding(style.MediaDurationPaddingY, style.MediaDurationPaddingX).
					Radius(style.RadiusChrome).
					Background(ui.RGB(0, 0, 0).Alpha(style.MediaDurationOpacity)).
					TextColor(ui.RGB(255, 255, 255))
			}
		})
	})
}

func mediaStat(c *ui.Context, icon *ui.SVG, count int) {
	white := ui.RGB(255, 255, 255)
	ui.Row(c).Gap(style.MediaStatGap).AlignItems(ui.Center).Children(func() {
		ui.Icon(c, icon).Size(style.MediaStatIconSize, style.MediaStatIconSize).Shrink(0).TextColor(white)
		ui.Text(c, mediaCount(count)).SingleLine().FontSize(style.Font(c, style.MediaStatFontSize)).TextColor(white)
	})
}

func mediaCount(n int) string {
	if n >= 10000 {
		return fmt.Sprintf("%.1f万", float64(n)/10000)
	}
	return fmt.Sprint(n)
}
