package components

import (
	"testing"

	"github.com/egoist/mygo/ui"
	"github.com/oneo-me/onevra/source/icons/lucide"

	"github.com/kamiertop/videodown/internal/gui/style"
)

// TestButtonIconLabelGap verifies the icon and the label keep the theme's
// two-unit gap instead of touching. The button centers its content, so
// adding icon(16)+gap(8) shifts the centered text by half of it: 12 DIP.
// A missing gap would shift it by only 8 (half of the icon width).
func TestButtonIconLabelGap(t *testing.T) {
	measure := func(withIcon bool) float32 {
		test := ui.NewTester(func(c *ui.Context) {
			style.UseTheme(c)
			opts := ButtonOptions{Variant: ButtonOutline}
			if withIcon {
				opts.Icon = lucide.Download()
			}
			Button(c, "下载视频", opts)
		}, 300, 80)
		test.Frame()
		text, ok := test.Find("下载视频")
		if !ok {
			t.Fatal("button label not found")
		}
		return text.X
	}
	shift := measure(true) - measure(false)
	if diff := shift - 12; diff < -1 || diff > 1 {
		t.Fatalf("icon-label gap wrong: centered text shifted %.1f DIP, want ~12 (icon 16 + gap 8, halved by centering)", shift)
	}
}
