package settings

import (
	"github.com/kamiertop/videodown/internal/gui/style"
	"math"
	"slices"
	"strings"

	"github.com/ELY-Open-Source/MujicaUI/core"
	"github.com/ELY-Open-Source/MujicaUI/input"
	"github.com/ELY-Open-Source/MujicaUI/theme"
	"github.com/egoist/mygo/ui"
	"github.com/kamiertop/videodown/internal/gui/components"
)

// Occurrences keep repeated fields independently draggable in edited templates.
type filenamePart struct {
	Token      string
	Occurrence int
}

type filenameOrderState struct {
	container ui.Handle
	items     map[filenamePart]*ui.Handle
}

func filenameParts(template string) []filenamePart {
	counts := map[string]int{}
	var parts []filenamePart
	for _, token := range filenameTokens.FindAllString(template, -1) {
		parts = append(parts, filenamePart{Token: token, Occurrence: counts[token]})
		counts[token]++
	}
	return parts
}

func (s *State) filenameOrder(c *ui.Context, disabled bool) {
	k := core.Tokens(c)
	parts := filenameParts(s.Config.Filename)
	region := ui.Column(c.Key("order")).Label("文件名字段排序区").FillWidth().MinHeight(52).
		Padding(8).Border(style.BorderWidth, k.Border).Radius(style.RadiusMenu).Disabled(disabled)
	state := ui.Local(region, "handles", func() filenameOrderState { return filenameOrderState{items: make(map[filenamePart]*ui.Handle)} })
	region.Bind(&state.container)
	region.Children(func() {
		ui.Row(c.Key("fields")).Wrap().FillWidth().MinHeight(style.InputHeight).Gap(8).Children(func() {
			for _, part := range parts {
				// Position-only motion avoids keeping removed button paint callbacks
				// alive beyond their build pass, as an exit transition would do.
				ui.Column(c.Key(part)).Transition(ui.ElementTransition{Position: true, Duration: core.Motion(c, theme.StateDuration)}).Children(func() {
					handle := state.items[part]
					if handle == nil {
						handle = &ui.Handle{}
						state.items[part] = handle
					}
					chip := components.Button(c, part.Token, input.ButtonOptions{Variant: input.Secondary, Disabled: disabled}).
						Label("排序字段 " + part.Token).Radius(style.RadiusMenu).Cursor(ui.CursorGrab).Bind(handle).Drag(part)
					if chip.Dragging() {
						chip.Cursor(ui.CursorGrabbing)
					}
				})
			}
		})
	})
	// Bind the entire region, including gaps and trailing whitespace, as a target.
	source, dropped := ui.Drop[filenamePart](region)
	_, over := ui.DragOver[filenamePart](region)
	if disabled || (!over && !dropped) {
		return
	}
	root := state.container.Bounds(c)
	bounds := make([]ui.Rect, 0, len(parts))
	for _, part := range parts {
		bounds = append(bounds, state.items[part].Bounds(c))
	}
	x, y, _ := region.PointerPosition()
	insertion, marker := filenameInsertion(ui.Point{X: root.X + x, Y: root.Y + y}, bounds)
	if over {
		// Capture only geometry and colors; drawing never reads frame elements.
		marker.X -= root.X
		marker.Y -= root.Y
		region.DrawOver(func(p *ui.Painter, r ui.Rect) {
			line := marker
			line.X += r.X
			line.Y += r.Y
			p.Fill(line, k.Accent, 1)
		})
	}
	if dropped {
		s.moveFilenamePart(source, insertion)
		c.Invalidate()
	}
}

// filenameInsertion finds the nearest before/after slot in a wrapping row.
// Distance to the vertical marker makes blank space after a row an end slot.
func filenameInsertion(pointer ui.Point, bounds []ui.Rect) (int, ui.Rect) {
	best := math.Inf(1)
	index := 0
	marker := ui.Rect{X: pointer.X, Y: pointer.Y, W: 2, H: 36}
	for i, r := range bounds {
		for side := 0; side < 2; side++ {
			x := r.X - 4
			if side == 1 {
				x = r.X + r.W + 4
			}
			dy := max(r.Y-pointer.Y, float32(0), pointer.Y-r.Y-r.H)
			dx := pointer.X - x
			distance := float64(dx*dx + dy*dy)
			if distance < best {
				best, index = distance, i+side
				marker = ui.Rect{X: x - 1, Y: r.Y, W: 2, H: r.H}
			}
		}
	}
	return index, marker
}

func (s *State) moveFilenamePart(source filenamePart, insertion int) {
	parts := filenameParts(s.Config.Filename)
	from := slices.Index(parts, source)
	if from < 0 {
		return
	}
	insertion = max(0, min(insertion, len(parts)))
	if insertion > from {
		insertion--
	}
	if insertion == from {
		return
	}
	parts = slices.Delete(parts, from, from+1)
	parts = slices.Insert(parts, insertion, source)
	tokens := make([]string, len(parts))
	for i, part := range parts {
		tokens[i] = part.Token
	}
	s.Config.Filename = strings.Join(tokens, s.separator)
}
