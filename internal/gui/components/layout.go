// 滚动区域：基于 Onevra ScrollView，绑定应用管理的滚动状态。
package components

import (
	"github.com/egoist/mygo/ui"
	newc "github.com/oneo-me/onevra/source/components"
)

// ScrollAreaOptions configure a ScrollArea.
type ScrollAreaOptions struct {
	// Label names the region for assistive technology.
	Label string
}

// ScrollArea scrolls children; state keeps and sets the offset, nil keeps it by place.
func ScrollArea(c *ui.Context, state *ui.ScrollState, opts ScrollAreaOptions, children func()) ui.Element {
	if children == nil {
		panic("layout: ScrollArea needs children")
	}
	e := newc.ScrollView(c, newc.ScrollViewOptions{})
	if state == nil {
		state = ui.Local(e, "scroll_area.state", func() ui.ScrollState { return ui.ScrollState{} })
	}
	e.TrackScroll(state)
	if opts.Label != "" {
		e.Label(opts.Label)
	}
	return e.Children(children)
}
