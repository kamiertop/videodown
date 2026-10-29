package gui

import (
	"github.com/egoist/mygo/ui"
	"github.com/kamiertop/videodown/internal/gui/navigation"
	"github.com/kamiertop/videodown/internal/gui/pages/home"
)

func (s *Shell) content(c *ui.Context) {
	if s.win != nil && !s.win.IsDestroyed() {
		s.activatePlatform()
	}
	if current, ok := navigation.ByPath(s.router.Path()); !ok || current.ID != "bilibili" {
		s.pages.Bilibili.Leave()
	}
	s.router.View(c, func(r *ui.Route) {
		p, ok := navigation.ByPath(r.Path())
		if !ok {
			r.Title("未找到")
			ui.Text(c, "页面不存在："+r.Path())
			return
		}
		r.Title(p.Label)
		if p.ID == "home" {
			home.Render(c, s.router.Push)
			return
		}
		if r.Match("/youtube/{section...}") {
			s.pages.YouTube.Layout(c, s.router.Path(), s.router.Push, func() {
				transition := s.router.Transition
				s.router.Transition = ui.TransitionFade
				defer func() { s.router.Transition = transition }()
				r.View(c, func(part *ui.Route) {
					path := "/youtube"
					if part.Path() != "/" {
						path += part.Path()
					}
					s.pages.YouTube.Form(c, path, s.scroll(path))
				})
			})
			return
		}
		if r.Match("/settings/{section...}") {
			s.pages.Settings.Layout(c, p, s.router.Path(), s.router.Push, func() {
				// Sections are peers, including /settings itself (the general page).
				// Apply one fade to this nested view instead of inferring depth.
				transition := s.router.Transition
				s.router.Transition = ui.TransitionFade
				defer func() { s.router.Transition = transition }()
				r.View(c, func(section *ui.Route) {
					path := "/settings"
					if section.Path() != "/" {
						path += section.Path()
					}
					if p.CurrentSection(path) == "ffmpeg" && !s.ffmpegChecked && !s.ffmpegChecking {
						s.checkFFmpeg()
					}
					s.pages.Settings.Form(c, p, path, s.settingsServices(), s.scroll(path))
				})
			})
			return
		}
		if r.Match("/bilibili/{section...}") {
			s.pages.Bilibili.Layout(c, s.router.Path(), s.router.Push, func() {
				transition := s.router.Transition
				s.router.Transition = ui.TransitionFade
				defer func() { s.router.Transition = transition }()
				r.View(c, func(section *ui.Route) {
					path := "/bilibili"
					if section.Path() != "/" {
						path += section.Path()
					}
					s.pages.Bilibili.Form(c, path, s.router.Push, func() {
						if s.router.CanGoBack() {
							s.router.Back()
						} else {
							s.router.Push("/bilibili/up")
						}
					})
				})
			})
			return
		}
		if r.Match("/douyin/{section...}") {
			s.pages.Douyin.Layout(c, s.router.Path(), s.router.Push, func() {
				transition := s.router.Transition
				s.router.Transition = ui.TransitionFade
				defer func() { s.router.Transition = transition }()
				r.View(c, func(part *ui.Route) {
					path := "/douyin"
					if part.Path() != "/" {
						path += part.Path()
					}
					s.pages.Douyin.Form(c, path, s.router.Push, func() {
						if s.router.CanGoBack() {
							s.router.Back()
						} else {
							s.router.Push("/douyin/user")
						}
					})
				})
			})
			return
		}

		ui.Text(c, "页面不存在："+r.Path())
	})
}

// Activate only the current route, never pages retained for a transition.
// Each platform guards its own initialization so returning preserves its state.
func (s *Shell) activatePlatform() {
	page, ok := navigation.ByPath(s.router.Path())
	if !ok {
		return
	}
	switch page.ID {
	case "bilibili":
		s.pages.Bilibili.Start()
	case "douyin":
		s.pages.Douyin.Start()
	case "youtube":
		s.pages.YouTube.Start()
	}
}
