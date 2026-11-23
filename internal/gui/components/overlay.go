// 模态对话框：基于 Onevra Dialog，带标题、描述与关闭按钮。
package components

import (
	"github.com/egoist/mygo/ui"
	newc "github.com/oneo-me/onevra/source/components"
	"github.com/oneo-me/onevra/source/icons/lucide"

	"github.com/kamiertop/videodown/internal/gui/style"
)

// DialogOptions configure a Dialog.
type DialogOptions struct {
	Title       string
	Description string
	// KeepOnBackdrop ignores clicks on the backdrop.
	KeepOnBackdrop bool
	// KeepOnEscape ignores Escape.
	KeepOnEscape bool
	// HideClose removes the close button from the header.
	HideClose bool
}

// Dialog shows a modal dialog while *open is true, trapping and then
// restoring the focus.
func Dialog(c *ui.Context, open *bool, opts DialogOptions, content func()) ui.Element {
	options := newc.DialogOptions{
		DisableBackdropClose: opts.KeepOnBackdrop,
		DisableEscapeClose:   opts.KeepOnEscape,
		Label:                opts.Title,
	}
	var panel ui.Element
	newc.Dialog(c, open, options, func(p ui.Element) {
		panel = p
		ui.Column(c).FillWidth().MinWidth(280).Gap(style.SpaceMedium).Children(func() {
			ui.Row(c).FillWidth().AlignItems(ui.Center).Gap(style.SpaceSmall).Children(func() {
				ui.Text(c, opts.Title).Grow(1).MinWidth(0).SingleLine().FontWeight(style.WeightSemibold).FontSize(style.Font(c, style.FontHeading))
				if !opts.HideClose {
					if newc.IconButton(c, lucide.X(), "关闭对话框", newc.IconButtonPlain).Clicked() {
						*open = false
					}
				}
			})
			if opts.Description != "" {
				ui.Text(c, opts.Description).FillWidth().TextColor(c.Theme().TextMuted)
			}
			if content != nil {
				content()
			}
		})
	})
	return panel
}
