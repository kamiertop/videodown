package gui

import (
	"context"
	"fmt"
	"os"

	"github.com/ELY-Open-Source/MujicaUI/core"
	"github.com/ELY-Open-Source/MujicaUI/feedback"
	"github.com/ELY-Open-Source/MujicaUI/input"
	"github.com/ELY-Open-Source/MujicaUI/layout"
	"github.com/ELY-Open-Source/MujicaUI/overlay"
	"github.com/egoist/mygo/ui"
	"github.com/kamiertop/videodown/internal/gui/components"
	"github.com/kamiertop/videodown/internal/gui/style"
	"github.com/kamiertop/videodown/internal/updater"
	"golang.org/x/mod/semver"
)

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

// installUpdate downloads the new release, then hands over to the standalone
// updater, which replaces the binary and relaunches the app once we exit.
func (s *Shell) installUpdate() {
	if s.updateChecker == nil || s.updateInstalling || s.updateResult.DownloadURL == "" {
		return
	}
	if s.updateResult.DownloadURL == updater.UnSupportedOS {
		s.updateError = updater.UnSupportedOS
		return
	}
	s.updateInstalling = true
	s.updateDownloaded = false
	s.updateError = ""
	s.updateDone, s.updateTotal = 0, 0
	ctx, cancel := context.WithCancel(s.jobs.ctx)
	s.cancelUpdate = cancel
	checker, downloadURL := s.updateChecker, s.updateResult.DownloadURL
	var path string
	s.runUpdateJob(func() error {
		var err error
		path, err = checker.DownloadUpdate(ctx, downloadURL)
		return err
	}, func(err error) {
		cancel()
		s.cancelUpdate = nil
		if ctx.Err() != nil {
			// Cancelled or shutting down. A finished download skips its own
			// cleanup, so remove the package here; path is empty on error.
			_ = os.Remove(path)
			return
		}
		if err != nil {
			s.updateInstalling = false
			s.updateError = "下载更新失败：" + err.Error()
			return
		}
		s.updateDownloaded = true
		if err := checker.InstallUpdate(path); err != nil {
			// The package will not be used; do not leave it on disk.
			_ = os.Remove(path)
			s.updateInstalling = false
			s.updateDownloaded = false
			s.updateError = err.Error()
			return
		}
		// The helper now waits for this process to exit; leave without the
		// active-download confirmation.
		s.confirmedQuit = true
	})
}

// stopUpdateDownload cancels a running update download and restores the dialog.
func (s *Shell) stopUpdateDownload() {
	if s.cancelUpdate != nil {
		s.cancelUpdate()
		s.cancelUpdate = nil
	}
	s.updateInstalling = false
	s.updateDownloaded = false
	s.updateDone, s.updateTotal = 0, 0
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
			if s.updateError != "" {
				components.Muted(c, s.updateError)
			}
			if s.updateInstalling {
				if s.updateDownloaded {
					core.Spinner(c.Key("update-installing"), core.SpinnerOptions{Size: 14, Label: "正在安装更新，应用即将重启…", ShowLabel: true})
					return
				}
				if s.updateTotal > 0 {
					fraction := float64(s.updateDone) / float64(s.updateTotal)
					feedback.Progress(c.Key("update-progress"), max(0, min(1, fraction)), feedback.ProgressOptions{})
					components.Muted(c, fmt.Sprintf("正在下载更新 %.1f / %.1f MB", mb(s.updateDone), mb(s.updateTotal)))
				} else {
					// The server sent no Content-Length; only downloaded bytes are known.
					components.Muted(c, fmt.Sprintf("正在下载更新 %.1f MB", mb(s.updateDone)))
				}
				ui.Row(c).FillWidth().Justify(ui.End).Gap(style.SpaceSmall).Children(func() {
					components.Button(c, "取消", input.ButtonOptions{Variant: input.Outline}).OnClick(s.stopUpdateDownload)
				})
				return
			}
			ui.Row(c).FillWidth().Justify(ui.End).Gap(style.SpaceSmall).Children(func() {
				components.Button(c, "稍后再说", input.ButtonOptions{Variant: input.Outline}).OnClick(func() { s.updateOpen = false })
				components.Button(c, "立即更新", input.ButtonOptions{Variant: input.Primary}).OnClick(s.installUpdate)
			})
		})
	})
}

func mb(n int64) float64 { return float64(n) / (1 << 20) }
