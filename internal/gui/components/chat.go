package components

import (
	"fmt"
	"strings"
	"time"

	"github.com/egoist/mygo/ui"
	newc "github.com/oneo-me/onevra/source/components"
	"github.com/oneo-me/onevra/source/icons/lucide"

	"github.com/kamiertop/videodown/internal/gui/style"
)

// codeBlockCopiedFor is how long the copy button shows its check.
const codeBlockCopiedFor = 2 * time.Second

// CodeBlockOptions configure a CodeBlock.
type CodeBlockOptions struct {
	// Language labels the block; it does not pick highlighting.
	Language    string
	LineNumbers bool
	// Wrap and Collapsed set the toggles' first state.
	Wrap, Collapsed bool
}

// codeBlockState holds the user's toggles and the copy confirmation.
type codeBlockState struct {
	wrap, collapsed bool
	copiedAt        time.Time
}

// CodeBlock shows code with a language label, copy, collapse and wrap toggles.
func CodeBlock(c *ui.Context, code string, opts CodeBlockOptions) ui.Element {
	t := c.Theme()
	root := ui.Column(c).Radius(style.RadiusPanel).Border(style.BorderWidth, t.Border).Background(t.Surface).Clip().MinWidth(0)
	st := ui.Local(root, "code", func() codeBlockState { return codeBlockState{wrap: opts.Wrap, collapsed: opts.Collapsed} })
	lines := strings.Split(strings.TrimSuffix(code, "\n"), "\n")
	root.Children(func() {
		ui.Row(c).Padding(0, 8, 0, 12).Gap(8).AlignItems(ui.Center).Background(t.SurfaceHover).Children(func() {
			lang := opts.Language
			if lang == "" {
				lang = "文本"
			}
			ui.Text(c, lang).FontSize(12).TextColor(t.TextMuted).SingleLine().Grow(1)
			wrapIcon, wrapName := lucide.TextAlignJustify(), "自动换行"
			if st.wrap {
				wrapName = "取消自动换行"
			}
			if newc.IconButton(c, wrapIcon, wrapName, newc.IconButtonCompact).Clicked() {
				st.wrap = !st.wrap
			}
			icon, name := lucide.Copy(), "复制代码"
			if left := codeBlockCopiedFor - c.Now().Sub(st.copiedAt); left > 0 {
				icon, name = lucide.Check(), "已复制"
				c.After(left)
			}
			if newc.IconButton(c, icon, name, newc.IconButtonCompact).Clicked() {
				c.WriteClipboard(code)
				st.copiedAt = c.Now()
			}
			fold, foldName := lucide.ChevronUp(), "折叠"
			if st.collapsed {
				fold, foldName = lucide.ChevronDown(), fmt.Sprintf("展开 %d 行", len(lines))
			}
			if newc.IconButton(c, fold, foldName, newc.IconButtonCompact).Clicked() {
				st.collapsed = !st.collapsed
			}
		})
		if st.collapsed {
			return
		}
		var body ui.Element
		if st.wrap {
			body = ui.Column(c).Padding(8, 12).MinWidth(0)
		} else {
			body = ui.ScrollHorizontal(c).Padding(8, 12).AlignItems(ui.Start)
		}
		body.Font(style.MonoFont).FontSize(13).Children(func() {
			for _, line := range lines {
				ui.Text(c, line).MinWidth(0)
			}
		})
	})
	return root
}
