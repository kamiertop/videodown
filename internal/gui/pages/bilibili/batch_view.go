package bilibili

import (
	"fmt"
	"github.com/kamiertop/videodown/internal/gui/components"

	"github.com/egoist/mygo/ui"
	bilidownload "github.com/kamiertop/videodown/bilibili/download"
	"github.com/kamiertop/videodown/internal/gui/icons"
	"github.com/kamiertop/videodown/internal/gui/style"
)

func batchStatus(b *batchSession) string {
	if b.StopRequested && b.Active {
		return "停止中，等待当前任务结束"
	}
	switch b.Status {
	case "queued":
		return "排队中"
	case "loading":
		return "加载下一页"
	case "downloading":
		sleeping := 0
		var only *batchItem
		for _, item := range b.Items {
			if item.Status == "downloading" {
				only = item
				if item.Progress.Phase == "sleep" {
					sleeping++
				}
			}
		}
		if b.InFlight == 1 && only != nil {
			if only.Progress.Phase == "sleep" {
				return fmt.Sprintf("休眠中 · %d 秒", max(0, only.Progress.SleepRemaining))
			}
			if only.Progress.Phase == "" {
				return "解析地址"
			}
		}
		return fmt.Sprintf("运行中 %d · 休眠 %d", b.InFlight, sleeping)
	case "stopped":
		return "已停止"
	case "done":
		if b.Failed > 0 {
			return "部分失败"
		}
		return "已完成"
	}
	return "等待开始"
}

func (s *State) batchDownloadCard(c *ui.Context, b *batchSession) {
	k := c.Theme()
	panel(c, fmt.Sprintf("批量下载：%d", b.ID), func() {
		var disclosure ui.CollapsibleParts
		ui.Row(c.Key("header")).FillWidth().Padding(style.SpaceSmall).Gap(style.SpaceSmall).AlignItems(ui.Center).Children(func() {
			disclosure = ui.CollapsibleBase(c.Key("disclosure"), &b.Expanded)
			header := disclosure.Trigger.Label(fmt.Sprintf("展开或折叠批量：%d", b.ID)).Grow(1).MinWidth(0).MinHeight(style.InputHeight).Padding(4).Gap(style.SpaceSmall).AlignItems(ui.Center).Radius(style.RadiusControl).FocusRing(true)
			if header.Hovered() {
				header.Background(k.SurfaceHover)
			}
			header.Children(func() {
				ui.Icon(c, icons.Must("chevron-right")).Size(16, 16).Rotate(90 * disclosure.Progress()).TextColor(k.TextMuted).Shrink(0)
				ui.Text(c, b.Source.Title).Grow(1).MinWidth(0).MaxLines(1).FontWeight(style.WeightSemibold)
			})
			if b.Active {
				downloadCardButton(c.Key("stop"), "停止", components.ButtonOptions{Variant: components.ButtonOutline, Label: fmt.Sprintf("停止批量：%d", b.ID), Disabled: b.StopRequested}).OnClick(func() { s.stopBatch(b) })
			} else {
				if b.Status == "stopped" || b.Failed > 0 {
					downloadCardButton(c.Key("resume"), "继续 / 重试", components.ButtonOptions{Variant: components.ButtonPrimary, Label: fmt.Sprintf("继续批量：%d", b.ID), Disabled: !s.loggedIn}).OnClick(func() { s.resumeBatch(b) })
				}
				downloadCardButton(c.Key("dismiss"), "关闭", components.ButtonOptions{Variant: components.ButtonGhost, Label: fmt.Sprintf("关闭批量：%d", b.ID)}).OnClick(func() { s.dismissBatch(b) })
			}
		})
		ui.Row(c.Key("summary")).FillWidth().Padding(0, style.SpaceMedium, style.SpaceSmall).Gap(style.SpaceSmall).Wrap().AlignItems(ui.Center).Children(func() {
			if b.Active && b.Status != "queued" {
				components.Spinner(c, components.SpinnerOptions{Size: 14, Label: batchStatus(b)})
			}
			ui.Text(c, batchStatus(b)).FontSize(style.Font(c, style.FontMeta)).TextColor(k.Accent)
			ui.Text(c, fmt.Sprintf("已加载 %d/%d · 已完成 %d · 跳过 %d · 失败 %d", len(b.Items), max(b.Total, len(b.Items)), b.Done, b.Skipped, b.Failed)).FontSize(style.Font(c, style.FontMeta)).TextColor(k.TextMuted)
		})
		disclosure.Panel(func() {
			ui.Column(c.Key("items")).FillWidth().Padding(0, style.SpaceSmall, style.SpaceSmall).Gap(style.SpaceSmall).Children(func() {
				if b.Error != "" {
					s.errorRow(c, b.Error, nil)
				}
				pending := 0
				for index, item := range b.Items {
					if item.Status == "done" || item.Status == "skipped" {
						continue
					}
					pending++
					ui.Row(c.Key(index)).Label(fmt.Sprintf("批量视频：%d:%s", b.ID, item.Task.Bvid)).FillWidth().Padding(style.SpaceSmall).Gap(style.SpaceSmall).Radius(style.RadiusControl).Background(k.Background).AlignItems(ui.Start).Children(func() {
						s.image(c, item.Task.Cover, style.BatchItemCoverWidth, style.BatchItemCoverHeight, false)
						ui.Column(c.Key("info")).Grow(1).MinWidth(0).Gap(4).Children(func() {
							ui.Text(c, item.Task.Title).FillWidth().MaxLines(2).FontWeight(style.WeightSemibold)
							if item.Error != "" {
								ui.Text(c, item.Error).FillWidth().MaxLines(2).TextColor(k.Danger).FontSize(style.Font(c, style.FontMeta))
							} else if item.Status == "downloading" && item.Progress.Phase != "" {
								downloadProgress(c.Key("progress"), item.Progress)
							} else if item.Status == "downloading" {
								components.Spinner(c, components.SpinnerOptions{Size: 14, Label: "正在解析地址", ShowLabel: true})
							} else {
								ui.Text(c, "等待解析与下载").FontSize(style.Font(c, style.FontMeta)).TextColor(k.TextMuted)
							}
						})
					})
				}
				if pending == 0 && !b.Active && b.Status == "done" {
					caption(c, "本批任务已完成")
				}
			})
		})
	})
}

// phaseLabel names a download phase for progress displays.
func phaseLabel(phase string) string {
	if label := map[string]string{"video": "下载视频", "audio": "下载音频", "merge": "合并文件", "done": "已完成", "error": "下载失败"}[phase]; label != "" {
		return label
	}
	return "下载中"
}

func downloadProgress(c *ui.Context, p bilidownload.Progress) {
	if p.Phase == "sleep" {
		// 进度条保持在顶部位置不变，休眠指示放在它下面，
		// 阶段切换时不会把进度条顶下去。
		fraction := 0.0
		if p.SleepTotal > 0 {
			fraction = float64(p.SleepTotal-p.SleepRemaining) / float64(p.SleepTotal)
		}
		components.Progress(c.Key("sleep-progress"), max(0, min(1, fraction)), components.ProgressOptions{}).Label("休眠进度")
		label := fmt.Sprintf("休眠中 · %d 秒", max(0, p.SleepRemaining))
		components.Spinner(c.Key("sleep"), components.SpinnerOptions{Size: 14, Label: label, ShowLabel: true})
		return
	}
	percent := max(0, min(100, p.Percent))
	components.Progress(c, percent/100, components.ProgressOptions{Label: phaseLabel(p.Phase), ShowValue: true, Indeterminate: p.Phase == "merge" || p.Percent < 0})
}
