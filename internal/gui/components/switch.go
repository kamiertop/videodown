package components

import (
	"github.com/ELY-Open-Source/MujicaUI/core"
	"github.com/ELY-Open-Source/MujicaUI/input"
	"github.com/ELY-Open-Source/MujicaUI/theme"
	"github.com/egoist/mygo/ui"
)

// SettingsSwitch keeps MujicaUI's switch input, focus and accessibility behavior.
// The pinned component exposes no shape options, so its decoration is hidden
// and a pointer-transparent layer draws the rounded track and thumb instead.
func SettingsSwitch(c *ui.Context, value *bool, label string, opts input.SwitchOptions) ui.Element {
	k := core.Tokens(c)
	var sw ui.Element
	ui.Row(c).Width(44).Shrink(0).Children(func() {
		sw = input.Switch(c, value, "", opts).Label(label).Width(44).Opacity(0)
		target := float32(0)
		if *value {
			target = 1
		}
		position := sw.Animate("rounded-position", target, core.Motion(c, theme.StateDuration))
		focused := sw.FocusVisible()
		track, thumb := k.ControlBorder.Mix(k.Background, .55), k.Surface
		if *value {
			track, thumb = k.Accent, k.OnAccent
			if sw.Hovered() {
				track = k.AccentHover
			}
		}
		if opts.Disabled || opts.Loading {
			track = track.Mix(k.Background, .55)
			thumb = thumb.Mix(k.Background, .35)
		}
		ui.Box(c).Absolute().Fill().PassThrough().Draw(func(p *ui.Painter, r ui.Rect) {
			pill := ui.Rect{X: r.X, Y: r.Y + (r.H-24)/2, W: 44, H: 24}
			p.Fill(pill, track, 12)
			knob := ui.Rect{X: pill.X + 3 + position*20, Y: pill.Y + 3, W: 18, H: 18}
			p.Fill(knob, thumb, 9)
			if focused {
				p.Stroke(ui.Rect{X: pill.X - 3, Y: pill.Y - 3, W: pill.W + 6, H: pill.H + 6}, k.Focus, 15, 2)
			}
		})
	})

	return sw
}
