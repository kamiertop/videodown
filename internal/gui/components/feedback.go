// 进度指示：基于 Onevra Progress，支持标签、百分比与不定进度。
package components

import (
	"fmt"

	"github.com/egoist/mygo/ui"
	newc "github.com/oneo-me/onevra/source/components"
)

// ProgressOptions configure a Progress.
type ProgressOptions struct {
	// Indeterminate shows work of unknown length; the value is ignored.
	Indeterminate bool
	// Label names the indicator and shows beside it.
	Label     string
	ShowValue bool
}

// Progress shows work done, value 0 to 1; indeterminate ones hold still
// under reduced motion.
func Progress(c *ui.Context, value float64, opts ProgressOptions) ui.Element {
	v := value
	if opts.Indeterminate {
		v = -1
	}
	if opts.Label == "" && !opts.ShowValue {
		return newc.Progress(c, v)
	}
	row := ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(8)
	row.Children(func() {
		newc.Progress(c, v).Grow(1).MinWidth(0)
		text := opts.Label
		if opts.ShowValue && !opts.Indeterminate {
			if text != "" {
				text += " "
			}
			text += fmt.Sprintf("%.0f%%", value*100)
		}
		if text != "" {
			ui.Text(c, text).SingleLine().Shrink(0).FontSize(12).TextColor(c.Theme().TextMuted)
		}
	})
	return row
}
