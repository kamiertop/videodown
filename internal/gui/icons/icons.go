// Package icons provides the app's named icon catalog on top of the Onevra
// Lucide icon set. Names keep their upstream kebab-case form.
package icons

import (
	"fmt"

	"github.com/egoist/mygo/ui"
	"github.com/oneo-me/onevra/source/icons/lucide"
)

// catalog maps an icon name to its Lucide accessor. Icons that upstream
// Lucide renamed ("trash-2" is today's "trash") use the current name.
var catalog = map[string]func() *ui.SVG{
	"check":         lucide.Check,
	"chevron-down":  lucide.ChevronDown,
	"chevron-left":  lucide.ChevronLeft,
	"chevron-right": lucide.ChevronRight,
	"chevron-up":    lucide.ChevronUp,
	"circle-alert":  lucide.CircleAlert,
	"clock":         lucide.Clock,
	"copy":          lucide.Copy,
	"download":      lucide.Download,
	"file-text":     lucide.FileText,
	"folder":        lucide.Folder,
	"history":       lucide.Rewind, // 播放历史；当前 Lucide 集没有 history，用后退指针表达
	"image":         lucide.Image,
	"info":          lucide.Info,
	"list":          lucide.List,
	"loader-circle": lucide.LoaderCircle,
	"maximize":      lucide.Maximize,
	"minus":         lucide.Minus,
	"moon":          lucide.Moon,
	"palette":       lucide.Palette,
	"play":          lucide.Play,
	"plus":          lucide.Plus,
	"refresh-cw":    lucide.RefreshCw,
	"settings":      lucide.Settings,
	"square":        lucide.Square,
	"star":          lucide.Star,
	"trash-2":       lucide.Trash,
	"user":          lucide.User,
	"video":         lucide.Video,
	"x":             lucide.X,
	"zap":           lucide.Zap,
}

// Get returns the named icon and whether it exists.
func Get(name string) (*ui.SVG, bool) {
	accessor, ok := catalog[name]
	if !ok {
		return nil, false
	}
	return accessor(), true
}

// Must returns the named icon, panicking when the catalog lacks it.
func Must(name string) *ui.SVG {
	icon, ok := Get(name)
	if !ok {
		panic(fmt.Sprintf("icons: unknown icon %q", name))
	}
	return icon
}
