package components

import (
	"github.com/egoist/mygo/ui"
	newc "github.com/oneo-me/onevra/source/components"

	"github.com/kamiertop/videodown/internal/gui/style"
)

// MenuButton shares stable geometry and selected decoration across page menus.
func MenuButton(c *ui.Context, label string, icon *ui.SVG, active bool) ui.Element {
	t := c.Theme()
	variant := ButtonGhost
	if active {
		variant = ButtonSecondary
	}
	button := Button(c, label, ButtonOptions{Variant: variant, Icon: icon}).
		Label(label).
		FontSize(style.Body(c)).
		Radius(style.RadiusMenu).
		Border(style.BorderWidth, ui.Transparent).Shrink(0)
	if active {
		button.Border(style.BorderWidth, t.Accent).Background(t.Selection).TextColor(t.Accent)
	}

	return button
}

// SwitchOptions configure a SettingsSwitch.
type SwitchOptions struct {
	Disabled bool
	// Description is read out after the label.
	Description string
}

// SettingsSwitch is the settings row's switch: mygo's SwitchBase owns
// input, focus and accessibility while a Onevra-themed track and thumb
// paint on top, so the returned element reports its own clicks.
func SettingsSwitch(c *ui.Context, value *bool, label string, opts SwitchOptions) ui.Element {
	t := c.Theme()
	sw := ui.SwitchBase(c, value).Size(44, 24).Shrink(0)
	if label != "" {
		sw.Label(label)
	}
	if opts.Description != "" {
		sw.Description(opts.Description)
	}
	target := float32(0)
	if *value {
		target = 1
	}
	position := sw.Animate("rounded-position", target, style.Motion(c, style.StateDuration))
	focused := sw.FocusVisible()
	track, thumb := t.Border.Mix(t.Background, .55), t.Surface
	if *value {
		track, thumb = t.Accent, t.AccentText
		if sw.Hovered() {
			track = t.AccentHover
		}
	}
	if opts.Disabled {
		track = track.Mix(t.Background, .55)
		thumb = thumb.Mix(t.Background, .35)
	}
	sw.Background(ui.Transparent).Draw(func(p *ui.Painter, r ui.Rect) {
		pill := ui.Rect{X: r.X, Y: r.Y + (r.H-24)/2, W: 44, H: 24}
		p.Fill(pill, track, 12)
		knob := ui.Rect{X: pill.X + 3 + position*20, Y: pill.Y + 3, W: 18, H: 18}
		p.Fill(knob, thumb, 9)
		if focused {
			p.Stroke(ui.Rect{X: pill.X - 3, Y: pill.Y - 3, W: pill.W + 6, H: pill.H + 6}, t.Focus, 15, 2)
		}
	})
	return sw.Disabled(opts.Disabled)
}

// SpinnerOptions configure a Spinner. Size is 16 DIPs when zero.
type SpinnerOptions struct {
	Size float32
	// Label is read out and shown beside the mark.
	Label     string
	ShowLabel bool
}

// Spinner turns a ring while work is pending.
func Spinner(c *ui.Context, opts SpinnerOptions) ui.Element {
	size := opts.Size
	if size == 0 {
		size = 16
	}
	label := opts.Label
	if label == "" {
		label = "加载中"
	}
	row := ui.Row(c).Gap(8).AlignItems(ui.Center)
	row.Children(func() {
		newc.Spinner(c, newc.SpinnerRing).Size(size, size).Shrink(0).Label(label)
		if opts.ShowLabel {
			ui.Text(c, label).TextColor(c.Theme().TextMuted).SingleLine()
		}
	})
	return row
}
