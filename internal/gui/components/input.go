package components

import (
	"github.com/egoist/mygo/ui"
	newc "github.com/oneo-me/onevra/source/components"

	"github.com/kamiertop/videodown/internal/gui/style"
)

// CheckboxOptions configure a Checkbox.
type CheckboxOptions struct {
	Disabled bool
}

// Checkbox toggles *value by a click on the box or its label, as does Space.
func Checkbox(c *ui.Context, value *bool, label string, opts CheckboxOptions) ui.Element {
	return newc.Checkbox(c, value, label).Disabled(opts.Disabled)
}

// ToggleOptions configure a Toggle.
type ToggleOptions struct {
	// Icon goes before the label.
	Icon     *ui.SVG
	Disabled bool
	// Label names an icon-only toggle for assistive technology.
	Label string
}

// Toggle is a button that stays pressed while *on.
func Toggle(c *ui.Context, on *bool, label string, opts ToggleOptions) ui.Element {
	b := newc.ToggleButton(c, on, label)
	if label == "" && opts.Label != "" {
		b.Label(opts.Label)
	}
	if opts.Icon != nil {
		b.Gap(c.Theme().Space(2))
		b.Children(func() {
			ui.Icon(c, opts.Icon).Size(buttonSpinnerSize, buttonSpinnerSize)
			if label != "" {
				ui.Text(c, label).SingleLine()
			}
		})
	}
	return b.Disabled(opts.Disabled)
}

// SliderOptions configure a Slider. Min and Max both zero mean 0 to 100.
type SliderOptions struct {
	Min, Max float64
	// Step snaps the value to Min plus multiples of it; zero is continuous.
	Step     float64
	Disabled bool
	// Label names the slider for assistive technology.
	Label string
}

// Slider drags *value between Min and Max, snapping to Step. The returned
// element reports Changed for the frame of a change.
func Slider(c *ui.Context, value *float64, opts SliderOptions) ui.Element {
	lo, hi := opts.Min, opts.Max
	if lo == 0 && hi == 0 {
		lo, hi = 0, 100
	}
	step := opts.Step
	if step == 0 {
		step = 1
	}
	s := newc.Slider(c, value, lo, hi, newc.SliderOptions{Step: step})
	if opts.Label != "" {
		s.Label(opts.Label)
	}
	if opts.Disabled {
		s.Element.Disabled(true)
	}
	return s.Element
}

// NumberInputOptions configure a NumberInput.
type NumberInputOptions struct {
	// Min and Max bound the value; equal bounds (the zero value) mean no bounds.
	Min, Max float64
	// Step is what Up, Down and the buttons add; zero means 1.
	Step float64
	// Label names the field for assistive technology.
	Label    string
	Disabled bool
}

// NumberInput edits *value with a stepper, clamped to its bounds. The
// returned element reports Changed for the frame of a change.
func NumberInput(c *ui.Context, value *float64, opts NumberInputOptions) ui.Element {
	lo, hi := opts.Min, opts.Max
	if lo == 0 && hi == 0 {
		lo, hi = -1e9, 1e9
	}
	step := opts.Step
	if step == 0 {
		step = 1
	}
	field := newc.NumberInput(c, value, lo, hi, step)
	if opts.Label != "" {
		field.Label(opts.Label)
	}
	if opts.Disabled {
		field.Element.Disabled(true)
	}
	return field.Element
}

// SelectOption is one choice of a Select.
type SelectOption[T comparable] struct {
	Value    T
	Label    string
	Disabled bool
}

// SelectOptions configure a Select.
type SelectOptions struct {
	Disabled bool
	// Label names the select for assistive technology.
	Label string
}

// SelectLabel returns the label options give v.
func SelectLabel[T comparable](options []SelectOption[T], v T) (string, bool) {
	for _, o := range options {
		if o.Value == v {
			return o.Label, true
		}
	}
	return "", false
}

// Select chooses *value from options in a popup, returning its trigger and
// whether the user changed the value this frame.
func Select[T comparable](c *ui.Context, value *T, options []SelectOption[T], opts SelectOptions) (ui.Element, bool) {
	old := *value
	t := c.Theme()
	sel := newc.Select(c, value)
	sel.Trigger.MinHeight(style.InputHeight).MinWidth(0)
	if opts.Label != "" {
		sel.Trigger.Label(opts.Label)
	}
	if opts.Disabled {
		sel.Trigger.Disabled(true)
	}
	label, has := SelectLabel(options, *value)
	sel.Trigger.Children(func() {
		text := ui.Text(c, label)
		if !has {
			text = ui.Text(c, "请选择")
			text.TextColor(t.TextMuted)
		}
		text.SingleLine().Grow(1).Shrink(1)
	})
	if !opts.Disabled {
		sel.Popup(func(panel ui.Element) {
			ui.Scroll(c).FillWidth().MaxHeight(280).Children(func() {
				if len(options) == 0 {
					ui.Text(c, "没有可选项").TextColor(t.TextMuted)
				}
				for _, o := range options {
					item := sel.Item(o.Value)
					if o.Disabled {
						item.Disabled(true)
					}
					item.Children(func() { ui.Text(c, o.Label).SingleLine() })
				}
			})
		})
	}
	return sel.Trigger, *value != old
}

// TextInputOptions configure a TextInput.
type TextInputOptions struct {
	Placeholder string
	// Label names the field for assistive technology.
	Label    string
	Disabled bool
}

// TextInput edits *value on a single line. The returned element carries
// the input's Changed and Submitted reports.
func TextInput(c *ui.Context, value *string, opts TextInputOptions) ui.Element {
	field := newc.TextBox(c, value, newc.TextBoxOptions{})
	if opts.Label != "" {
		field.Label(opts.Label)
	}
	if opts.Placeholder != "" {
		field.Placeholder(opts.Placeholder)
	}
	if opts.Disabled {
		field.Disabled(true)
	}
	return field
}

// TextAreaOptions configure a TextArea.
type TextAreaOptions struct {
	Placeholder string
	// Label names the field for assistive technology.
	Label string
	// MinHeight is the least height in DIPs.
	MinHeight float32
	// MaxHeight caps the height in DIPs, then the text scrolls; zero grows freely.
	MaxHeight float32
	Disabled  bool
}

// TextArea edits *value across several lines.
func TextArea(c *ui.Context, value *string, opts TextAreaOptions) ui.Element {
	field := newc.TextBox(c, value, newc.TextBoxOptions{Multiline: true})
	if opts.Label != "" {
		field.Label(opts.Label)
	}
	if opts.Placeholder != "" {
		field.Placeholder(opts.Placeholder)
	}
	if opts.MinHeight > 0 {
		field.MinHeight(opts.MinHeight)
	}
	if opts.MaxHeight > 0 {
		field.MaxHeight(opts.MaxHeight)
	}
	if opts.Disabled {
		field.Disabled(true)
	}
	return field
}

// TextInputEditor places the native editor for value inside a custom frame.
// While locked the editor edits a copy, keeping the real value untouched.
func TextInputEditor(c *ui.Context, frame ui.Element, value *string, locked bool, base func(*ui.Context, *string) ui.Element) ui.Element {
	if !locked {
		return base(c, value)
	}
	shown := ui.Local(frame, "locked", func() string { return *value })
	*shown = *value
	in := base(c, shown)
	*shown = *value
	return in
}
