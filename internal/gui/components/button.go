package components

import (
	"github.com/egoist/mygo/ui"
	newc "github.com/oneo-me/onevra/source/components"

	"github.com/kamiertop/videodown/internal/gui/style"
)

// ButtonVariant selects a button's emphasis. The zero value is primary.
type ButtonVariant int

const (
	ButtonPrimary ButtonVariant = iota
	ButtonSecondary
	ButtonOutline
	ButtonGhost
	ButtonDanger
)

// ButtonOptions configure a Button.
type ButtonOptions struct {
	Variant ButtonVariant
	// Icon goes before the label.
	Icon     *ui.SVG
	Loading  bool
	Disabled bool
	// Label names an icon-only button for assistive technology; required without a label.
	Label string
}

// buttonSpinnerSize is the leading icon slot, which the spinner takes while loading.
const buttonSpinnerSize float32 = 16

// Button is the shared app button: Onevra's face with the app's tokens and
// interaction options. Change RadiusControl in style/tokens.go to adjust
// every button using this component.
func Button(c *ui.Context, label string, opts ButtonOptions) ui.Element {
	if label == "" && opts.Label == "" {
		panic("components: Button without a label needs Label to name it")
	}
	var b ui.Element
	switch opts.Variant {
	case ButtonDanger:
		b = dangerFace(c)
	default:
		b = newc.Button(c, "", newVariant(opts.Variant))
	}
	if opts.Label != "" {
		b.Label(opts.Label)
	}
	// Onevra 的按钮默认单子元素，不设 Gap；图标与文字同放时需要间距。
	b.Gap(c.Theme().Space(2))
	b.MinHeight(style.InputHeight).MinWidth(style.MinHitSize).
		Radius(style.RadiusControl).
		FontSize(style.Font(c, style.FontButton)).
		FontWeight(style.WeightSemibold)
	if opts.Disabled || opts.Loading {
		b.Disabled(true)
	}
	b.Children(func() {
		if opts.Loading {
			newc.Spinner(c, newc.SpinnerRing).Size(buttonSpinnerSize, buttonSpinnerSize)
		} else if opts.Icon != nil {
			ui.Icon(c, opts.Icon).Size(buttonSpinnerSize, buttonSpinnerSize)
		}
		if label != "" {
			ui.Text(c, label).SingleLine()
		}
	})
	return b
}

// newVariant maps the app's variants onto Onevra's.
func newVariant(v ButtonVariant) newc.ButtonVariant {
	switch v {
	case ButtonSecondary:
		return newc.ButtonSecondary
	case ButtonOutline:
		return newc.ButtonOutline
	case ButtonGhost:
		return newc.ButtonGhost
	default:
		return newc.ButtonPrimary
	}
}

// dangerFace styles a bare button base as the danger action, following
// Onevra's button metrics with the theme's danger fill.
func dangerFace(c *ui.Context) ui.Element {
	t := c.Theme()
	b := ui.ButtonBase(c).Padding(t.Space(1.5), t.Space(3.5)).Radius(t.Radius)
	bg, fg := t.Danger, ui.Hex("#FFF8EE")
	if !b.IsDisabled() {
		black := ui.RGB(0, 0, 0)
		switch {
		case b.Pressed():
			bg = t.Danger.Mix(black, 0.25)
		case b.Hovered():
			bg = t.Danger.Mix(black, 0.12)
		}
	}
	return b.Background(bg).TextColor(fg).Border(1, bg)
}

// IconButtonOptions configure an IconButton.
type IconButtonOptions struct {
	// Tooltip shows on hover; empty uses the button's name.
	Tooltip  string
	Loading  bool
	Disabled bool
}

// IconButton is a square button showing only an icon; label names its action.
func IconButton(c *ui.Context, icon *ui.SVG, label string, opts IconButtonOptions) ui.Element {
	if icon == nil {
		panic("components: IconButton needs an icon")
	}
	if label == "" {
		panic("components: IconButton needs a label naming its action")
	}
	b := newc.IconButton(c, icon, label, newc.IconButtonPlain)
	if opts.Disabled || opts.Loading {
		b.Disabled(true)
	}
	tooltip := opts.Tooltip
	if tooltip == "" {
		tooltip = label
	}
	return b.Tooltip(tooltip)
}
