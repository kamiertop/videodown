package gui

import (
	"context"
	"net/url"

	"github.com/ELY-Open-Source/MujicaUI/input"
	"github.com/ELY-Open-Source/MujicaUI/layout"
	"github.com/ELY-Open-Source/MujicaUI/overlay"
	"github.com/egoist/mygo/ui"
	"github.com/kamiertop/videodown/internal/gui/assets"
	"github.com/kamiertop/videodown/internal/gui/components"
	"github.com/kamiertop/videodown/internal/gui/style"
	"github.com/kamiertop/videodown/internal/updater"
	"golang.org/x/mod/semver"
)

type UpdateChecker interface {
	Check(context.Context) (updater.Result, error)
}

func (s *Shell) startUpdateCheck() {
	if s.updateStarted || s.updateChecker == nil || s.runUpdateJob == nil {
		return
	}
	s.updateStarted = true
	if !s.pages.Settings.Config.AutoUpdate {
		return
	}
	checker, ctx := s.updateChecker, s.jobs.ctx
	var result updater.Result
	s.runUpdateJob(func() (err error) { result, err = checker.Check(ctx); return err }, func(err error) {
		if ctx.Err() != nil || err != nil || !s.pages.Settings.Config.AutoUpdate || !result.Available || !semver.IsValid(result.LatestVersion) {
			return
		}
		s.updateResult = result
		s.updateOpen = true
	})
}
func (s *Shell) updateDialog(c *ui.Context) {
	if s.closing.Open || s.quitting {
		return
	}
	title := "发现新版本"
	if !semver.IsValid(s.updateResult.CurrentVersion) {
		title = "发现正式版本"
	}
	overlay.Dialog(c.Key("app-update"), &s.updateOpen, overlay.DialogOptions{Title: title, Description: "最新版本：" + s.updateResult.LatestVersion}, func() {
		ui.Column(c).FillWidth().Gap(style.SpaceMedium).Children(func() {
			current := s.updateResult.CurrentVersion
			if current == "" {
				current = "dev"
			}
			components.Muted(c, "当前版本："+current)
			if s.updateResult.ReleaseNotes != "" {
				layout.ScrollArea(c.Key("release-notes"), &s.updateNotes, layout.ScrollAreaOptions{}, func() { ui.Text(c, s.updateResult.ReleaseNotes).FillWidth() }).MaxHeight(180)
			}
			services := c.Services()
			ui.Row(c).FillWidth().Justify(ui.End).Gap(style.SpaceSmall).Children(func() {
				components.Button(c, "稍后再说", input.ButtonOptions{Variant: input.Outline}).OnClick(func() { s.updateOpen = false })
				components.Button(c, "查看更新", input.ButtonOptions{Variant: input.Primary}).OnClick(func() {
					services.OpenURL(assets.RepositoryURL + "/releases/tag/" + url.PathEscape(s.updateResult.LatestVersion))
					s.updateOpen = false
				})
			})
		})
	})
}
