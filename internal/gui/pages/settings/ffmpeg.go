package settings

import (
	"github.com/ELY-Open-Source/MujicaUI/chat"
	"github.com/ELY-Open-Source/MujicaUI/core"
	"github.com/ELY-Open-Source/MujicaUI/input"
	"github.com/egoist/mygo/ui"
	"github.com/kamiertop/videodown/internal/gui/components"
	"github.com/kamiertop/videodown/internal/gui/style"
)

func (s *State) ffmpegSettings(c *ui.Context, w Services) {
	settingCard(c, "FFmpeg 状态", "video", func() {
		status := "尚未检测"
		description := "请安装或手动指定 FFmpeg 路径"
		if w.FFmpegChecking {
			status = "检测中…"
		} else if w.FFmpegChecked {
			if w.FFmpegError != "" {
				status = "未检测到"
			} else {
				status = "已就绪"
				description = "FFmpeg 可正常使用"
			}
		}
		ui.Row(c).FillWidth().Gap(12).Children(func() {
			ui.Text(c, status).FontWeight(style.WeightSemibold).Shrink(0)
			ui.Column(c).Grow(1).MinWidth(0).Children(func() { components.Muted(c, description) })
		})
		if w.FFmpegPath != "" {
			previewBox(c, "当前路径", w.FFmpegPath)
			components.Muted(c, w.FFmpegVersion)
		}
		if w.FFmpegError != "" {
			components.Muted(c, w.FFmpegError)
		}
		components.Button(c.Key("detect"), "重新检测", input.ButtonOptions{Variant: input.Outline, Loading: w.FFmpegChecking, Disabled: !w.SettingsAvailable}).OnClick(func() {
			if w.CheckFFmpeg != nil {
				w.CheckFFmpeg()
			}
		})
	})
	settingCard(c, "自定义路径", "folder", func() {
		components.Muted(c, "自动检测不到时，可手动指定 FFmpeg 完整路径；也可将 ffmpeg 放到应用同目录或 bin/ 子目录。")
		ui.Row(c).FillWidth().Gap(8).Children(func() {
			input.TextInput(c.Key("ffmpeg"), &s.Config.FFmpeg, input.TextInputOptions{Label: "FFmpeg 路径", Placeholder: "例如 /usr/bin/ffmpeg 或 C:\\ffmpeg\\bin\\ffmpeg.exe", ReadOnly: true}).Element.Grow(1).MinWidth(0)
			components.Button(c.Key("browse"), "浏览", input.ButtonOptions{Disabled: w.BrowsingSettings || !w.SettingsAvailable}).OnClick(func() {
				if w.BrowseFFmpeg != nil {
					w.BrowseFFmpeg()
				}
			})
		})
	})
	settingCard(c, "安装指南", "info", func() {
		ui.Text(c, "Windows").FontWeight(style.WeightSemibold)
		k := core.Tokens(c)
		ui.RichText(c).FillWidth().TextColor(k.TextMuted).Children(func() {
			ui.Text(c, "从 ")
			ui.Link(c, "ffmpeg.org ↗", "https://ffmpeg.org/download.html").Underline().TextColor(k.AccentText).Tooltip("打开 FFmpeg 官网下载页")
			ui.Text(c, " 下载，将 ffmpeg.exe 放到应用同目录或 bin/ 子目录，或添加到系统 PATH。")
		})
		ui.Text(c, "macOS").FontWeight(style.WeightSemibold)

		chat.CodeBlock(c.Key("macos-install"), "brew install ffmpeg", codeBockOption).Element.FillWidth().Label("macOS 安装命令")
		ui.Text(c, "Linux").FontWeight(style.WeightSemibold)
		components.Muted(c, "Debian / Ubuntu")
		chat.CodeBlock(c.Key("apt-install"), "sudo apt install ffmpeg", codeBockOption).Element.FillWidth().Label("Debian / Ubuntu 安装命令")
		components.Muted(c, "Fedora / Rocky / Centos")
		chat.CodeBlock(c.Key("dnf-install"), "sudo dnf install ffmpeg", codeBockOption).Element.FillWidth().Label("Fedora 安装命令")
	})
}

var codeBockOption = chat.CodeBlockOptions{
	Language:    "shell",
	LineNumbers: false,
	Wrap:        false,
	Collapsed:   false,
}
