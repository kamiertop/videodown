package douyin

import (
	"fmt"
	"net/url"
	"slices"
	"strings"

	"github.com/ELY-Open-Source/MujicaUI/core"
	"github.com/ELY-Open-Source/MujicaUI/icons"
	"github.com/ELY-Open-Source/MujicaUI/input"
	"github.com/ELY-Open-Source/MujicaUI/layout"
	"github.com/ELY-Open-Source/MujicaUI/overlay"
	"github.com/egoist/mygo/ui"
	"github.com/kamiertop/videodown/internal/gui/components"
	"github.com/kamiertop/videodown/internal/gui/style"
)

func panel(c *ui.Context, label string, body func()) ui.Element {
	k := core.Tokens(c)
	return ui.Column(c).Label(label).FillWidth().MinWidth(0).Radius(style.RadiusPanel).Border(style.BorderWidth, k.Border).Background(k.Surface).Clip().Children(body)
}
func button(c *ui.Context, text string, opts input.ButtonOptions) ui.Element {
	return components.Button(c, text, opts)
}
func caption(c *ui.Context, text string) {
	ui.Text(c, text).FontSize(style.Font(c, style.FontMeta)).TextColor(core.Tokens(c).TextMuted).MaxLines(2)
}

// Use the MujicaUI editor in an app-sized rounded frame, including its clear action.
func field(c *ui.Context, value *string, label, placeholder string, submit func()) ui.Element {
	k := core.Tokens(c)
	frame := ui.Row(c).MinWidth(0).MinHeight(style.InputHeight).Padding(0, style.SpaceMedium).Gap(style.SpaceSmall).
		Radius(style.RadiusControl).Border(style.BorderWidth, k.ControlBorder).Background(k.Surface)
	var editor ui.Element
	frame.Children(func() {
		editor = input.TextInputEditor(c.Key("editor"), frame, value, false, ui.TextInputBase).
			Label(label).Placeholder(placeholder).Grow(1).MinWidth(0).FocusRing(false)
		if submit != nil {
			editor.OnSubmit(submit)
		}
		if *value != "" {
			button(c.Key("clear"), "", input.ButtonOptions{Variant: input.Ghost, Icon: icons.Must("x"), Label: "清除" + label}).Size(24, 24).MinHeight(24).MinWidth(24).Padding(0).OnClick(func() { *value = "" })
		}
	})
	if editor.Focused() {
		frame.Border(style.BorderWidth, k.Focus)
	}
	return frame
}
func (s *State) errorRow(c *ui.Context, message string, retry func()) {
	if message == "" {
		return
	}
	ui.Row(c).FillWidth().Gap(style.SpaceSmall).AlignItems(ui.Center).Children(func() {
		ui.Text(c, message).Grow(1).MinWidth(0).MaxLines(3).TextColor(core.Tokens(c).DangerFill)
		if retry != nil {
			button(c, "重试", input.ButtonOptions{Variant: input.Outline}).OnClick(retry)
		}
	})
}
func empty(c *ui.Context, message string) {
	ui.Column(c).FillWidth().Grow(1).MinHeight(200).Center().Gap(style.SpaceMedium).Children(func() {
		ui.Icon(c, icons.Must("play")).Size(28, 28).TextColor(core.Tokens(c).TextMuted)
		caption(c, message)
	})
}
func section(path string) string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) < 2 {
		return "download"
	}
	return parts[1]
}
func (s *State) Layout(c *ui.Context, path string, navigate func(string), content func()) {
	s.enter(path)
	k := core.Tokens(c)
	ui.Row(c.Key("douyin-layout")).Fill().MinHeight(0).AlignItems(ui.Stretch).Children(func() {
		ui.Column(c.Key("rail")).Label("抖音导航").Width(style.PlatformNavWidth).Shrink(0).FillHeight().Padding(8).Gap(6).Background(k.Surface).BorderColor(k.Border).BorderWidth(0, 1, 0, 0).Children(func() {
			ui.Text(c, "抖音").FillWidth().TextAlign(ui.Center).FontSize(style.Font(c, style.FontCompact)).TextColor(k.TextMuted)
			layout.ScrollArea(c.Key("menu"), s.scroll("menu"), layout.ScrollAreaOptions{}, func() {
				ui.Column(c).FillWidth().Gap(6).Children(func() {
					ids := []string{"download", "user", "dynamic", "favorite", "play-history", "history"}
					labels := []string{"下载", "用户", "动态", "收藏 / 合集", "播放历史", "下载历史"}
					names := []string{"download", "user", "play", "star", "history", "clock"}
					for i, id := range ids {
						b := button(c.Key(id), labels[i], input.ButtonOptions{Variant: input.Ghost, Icon: icons.Must(names[i]), Label: "抖音" + labels[i]}).FillWidth().Height(style.PlatformNavItemHeight).Column().Center().Gap(7).Padding(4).FontSize(style.Font(c, style.FontCompact)).Border(1, k.Border).Background(k.Background)
						if section(path) == id {
							b.Background(k.Selection).TextColor(k.AccentText).Border(1, k.AccentText)
						}
						b.OnClick(func() { navigate("/douyin/" + id) })
					}
				})
			}).Grow(1).MinHeight(0)
			avatar := first(s.profile.Data.AvatarThumb.UrlList)
			if avatar == "" {
				avatar = first(s.profile.Data.AvatarMedium.UrlList)
			}
			if avatar == "" {
				avatar = first(s.profile.Data.AvatarLarger.UrlList)
			}
			if s.loggedIn() && avatar != "" {
				ui.Column(c.Key("account-avatar")).Label("抖音账号头像").FillWidth().AlignItems(ui.Center).Children(func() { s.image(c, avatar, 32, 32, true) })
			}
			button(c.Key("account"), "账号", input.ButtonOptions{Variant: input.Ghost, Icon: icons.Must("user"), Label: "抖音账号"}).FillWidth().Height(style.PlatformNavItemHeight).Column().Center().Gap(7).Padding(4).FontSize(style.Font(c, style.FontCompact)).OnClick(func() { navigate("/douyin/profile") })
		})
		content()
	})
}
func (s *State) Form(c *ui.Context, path string, navigate func(string), back func()) {
	ui.Column(c.Key("content")).Label("抖音内容").Grow(1).MinWidth(0).MinHeight(0).Padding(style.PagePadding).Gap(style.SpaceMedium).FontSize(style.Body(c)).Children(func() {
		s.errorRow(c, s.actionError, nil)
		if s.notice != "" {
			caption(c, s.notice)
		}
		switch section(path) {
		case "profile":
			s.accountPage(c)
		case "download":
			s.downloadPage(c)
		case "history":
			s.historyPage(c)
		default:
			if !s.loggedIn() {
				ui.Column(c).FillWidth().Grow(1).MinHeight(0).Center().Children(func() {
					panel(c, "登录提示", func() {
						ui.Column(c).FillWidth().Padding(style.SpaceXLarge).Gap(style.SpaceMedium).Center().Children(func() {
							caption(c, "请先设置抖音 Cookie")
							button(c, "前往账号", input.ButtonOptions{Variant: input.Outline}).OnClick(func() { navigate("/douyin/profile") })
						})
					}).MaxWidth(style.ProfileLoginMaxWidth)
				})
				return
			}
			if section(path) == "user" && len(strings.Split(strings.Trim(path, "/"), "/")) == 2 {
				s.authorsPage(c, navigate)
				return
			}
			if section(path) == "user" {
				panel(c, "抖音作者资料", func() {
					ui.Row(c).FillWidth().Padding(style.SpaceSmall).Gap(style.SpaceSmall).AlignItems(ui.Center).Children(func() {
						s.image(c, first(s.author.Data.AvatarThumb.UrlList), 36, 36, true)
						ui.Text(c, s.author.Data.NickName).FontWeight(style.WeightSemibold)
						ui.Text(c, s.author.Data.Signature).Grow(1).MinWidth(0).MaxLines(1)
						button(c, "返回", input.ButtonOptions{Variant: input.Outline}).OnClick(back)
					})
				})
				s.errorRow(c, s.author.Error, nil)
			}
			s.libraryPage(c, path, navigate)
		}
	})
}
func (s *State) Render(c *ui.Context, id string, navigate func(string)) {
	s.Form(c, "/douyin/"+id, navigate, func() { navigate("/douyin/user") })
}
func (s *State) accountPage(c *ui.Context) {
	ui.Column(c).Fill().MinWidth(0).Center().Children(func() {
		panel(c, "抖音账号信息", func() {
			ui.Column(c).FillWidth().Padding(style.SpaceXLarge).Gap(style.SpaceLarge).Children(func() {
				if s.profile.Loading {
					caption(c, "正在读取账号")
				}
				s.errorRow(c, s.profile.Error, s.loadProfile)
				if s.loggedIn() {
					ui.Row(c).FillWidth().Gap(style.SpaceLarge).Children(func() {
						s.image(c, first(s.profile.Data.AvatarLarger.UrlList), 64, 64, true)
						ui.Column(c).Grow(1).MinWidth(0).Gap(6).Children(func() {
							caption(c, "当前登录账号")
							ui.Text(c, s.profile.Data.NickName).FontSize(style.Font(c, style.FontTitle)).FontWeight(style.WeightSemibold)
							caption(c, s.profile.Data.Signature)
							caption(c, "抖音号："+s.profile.Data.UniqueId)
						})
					})
					ui.Row(c).FillWidth().Gap(style.SpaceSmall).Children(func() {
						button(c, "刷新账号", input.ButtonOptions{Variant: input.Outline, Disabled: s.authBusy}).OnClick(s.loadProfile)
						button(c, "更新 Cookie", input.ButtonOptions{Variant: input.Outline, Disabled: s.PendingDownloadCount() > 0}).OnClick(func() { s.editCookie = !s.editCookie })
						button(c, "退出登录", input.ButtonOptions{Variant: input.Ghost, Disabled: s.authBusy || s.PendingDownloadCount() > 0}).OnClick(func() { s.saveCookie(true) })
					})
				}
				if !s.loggedIn() || s.editCookie {
					input.TextArea(c.Key("cookie"), &s.cookie, input.TextAreaOptions{Label: "抖音 Cookie", Placeholder: "粘贴抖音网页 Cookie", MinHeight: 110, MaxHeight: 160, Disabled: s.authBusy}).Element.FillWidth()
					ui.Row(c).FillWidth().Gap(style.SpaceSmall).Children(func() {
						button(c, "保存并验证", input.ButtonOptions{Variant: input.Primary, Loading: s.authBusy, Disabled: strings.TrimSpace(s.cookie) == "" || s.api == nil || s.PendingDownloadCount() > 0}).OnClick(func() { s.saveCookie(false) })
						services := c.Services()
						button(c, "打开抖音", input.ButtonOptions{Variant: input.Outline}).OnClick(func() { services.OpenURL("https://www.douyin.com/") })
					})
				}
			})
		}).MaxWidth(style.ProfileCardMaxWidth)
	})
}
func (s *State) libraryPage(c *ui.Context, path string, navigate func(string)) {
	authorMode := section(path) == "user"
	libraryMode := section(path) == "favorite"

	panel(c, "抖音视频工具栏", func() {
		ui.Row(c).FillWidth().Wrap().Padding(style.SpaceSmall).Gap(style.SpaceSmall).AlignItems(ui.Center).Children(func() {
			if authorMode || libraryMode {
				tabs := []struct{ id, label string }{{"favorite", "收藏视频"}, {"folder", "收藏夹"}, {"collection", "合集"}}
				if authorMode {
					tabs = []struct{ id, label string }{{"posts", "作品"}, {"series", "合集"}}
				}
				for _, tab := range tabs {
					active := s.libraryTab == tab.id
					if authorMode {
						active = s.authorTab == tab.id
					}
					variant := input.Outline
					if active {
						variant = input.Primary
					}
					button(c.Key(tab.id), tab.label, input.ButtonOptions{Variant: variant}).OnClick(func() {
						if authorMode {
							s.authorTab = tab.id
							if tab.id == "posts" {
								s.useSource(s.authorSource(s.author.Data.SecUid, s.author.Data.NickName))
							} else {
								s.useSource(source{})
								s.prepareFolders("series", s.author.Data.SecUid)
							}
						} else {
							target := "/douyin/favorite"
							if tab.id != "favorite" {
								target += "/" + tab.id
							}
							navigate(target)
						}
					})
				}
			}

			name := s.source.Title
			if name == "" {
				name = "请选择列表"
			}
			ui.Text(c, name).Grow(1).MinWidth(0).MaxLines(1).FontWeight(style.WeightSemibold)
			caption(c, fmt.Sprintf("已加载 %d", len(s.videos.Data.Items)))
			button(c, "刷新视频", input.ButtonOptions{Variant: input.Outline, Loading: s.videos.Loading, Disabled: s.source.Load == nil}).OnClick(func() { s.useSource(s.source) })
			count := 0
			for _, v := range s.videos.Data.Items {
				if s.selected[v.AwemeId] {
					count++
				}
			}
			all := count == len(s.videos.Data.Items) && count > 0
			label := "全选"
			if all {
				label = "取消全选"
			}
			button(c, label, input.ButtonOptions{Variant: input.Ghost, Disabled: len(s.videos.Data.Items) == 0}).OnClick(func() {
				for _, v := range s.videos.Data.Items {
					s.selected[v.AwemeId] = !all
				}
			})
			button(c, fmt.Sprintf("下载已选 (%d)", count), input.ButtonOptions{Variant: input.Primary, Disabled: count == 0}).OnClick(func() { s.stageSelected(navigate) })
			if s.source.Load != nil && section(path) != "dynamic" && section(path) != "play-history" {
				button(c, "一键下载全部", input.ButtonOptions{Variant: input.Ghost, Disabled: s.backend == nil}).OnClick(func() { s.requestAll(navigate) })
			}
		})
	})

	showFolders := libraryMode && s.libraryTab != "favorite" || authorMode && s.authorTab == "series"
	ui.Row(c).FillWidth().Grow(1).MinHeight(0).AlignItems(ui.Stretch).Gap(style.SpaceMedium).Children(func() {
		if showFolders {
			panel(c.Key("folders"), "抖音来源列表", func() {
				ui.Row(c).FillWidth().Padding(style.SpaceSmall).Gap(style.SpaceSmall).Children(func() {
					ui.Text(c, fmt.Sprintf("列表 · %d", len(s.folders.Data))).Grow(1)
					button(c, "刷新列表", input.ButtonOptions{Variant: input.Ghost, Loading: s.folders.Loading}).OnClick(func() { s.loadFolders(false) })
				})
				s.errorRow(c, s.folders.Error, func() { s.loadFolders(s.folders.Loaded) })
				layout.ScrollArea(c.Key("folders-scroll"), s.scroll("folders"), layout.ScrollAreaOptions{}, func() {
					ui.Column(c).FillWidth().Padding(style.SpaceSmall).Gap(4).Children(func() {
						for _, f := range s.folders.Data {
							variant := input.Ghost
							kind := "collection"
							if s.libraryTab == "folder" && !authorMode {
								kind = "folder"
							}
							if s.source.Key == kind+":"+f.Key {
								variant = input.Secondary
							}
							button(c.Key(f.Key), fmt.Sprintf("%s (%d)", f.Title, f.Count), input.ButtonOptions{Variant: variant, Icon: icons.Must("folder")}).FillWidth().MinWidth(0).Justify(ui.Start).OnClick(func() {
								if authorMode {
									s.useSource(s.folderSource(kind, f))
								} else {
									navigate("/douyin/favorite/" + kind + "/" + url.PathEscape(f.Key))
								}
							})
						}
						if s.folders.Loading {
							core.Spinner(c, core.SpinnerOptions{Size: 16, Label: "正在加载列表", ShowLabel: true})
						}
						if len(s.folders.Data) == 0 && !s.folders.Loading {
							caption(c, "暂无列表")
						}
					})
				}).Grow(1).MinHeight(0)
			}).Width(style.LibraryListWidth).Shrink(0).FillHeight()
			s.maybeMoreFolders()
		}
		ui.Column(c).Grow(1).MinWidth(0).MinHeight(0).Gap(style.SpaceSmall).Children(func() {
			s.errorRow(c, s.videos.Error, func() { s.loadVideos(false) })
			layout.ScrollArea(c.Key("videos"), s.scroll("videos"), layout.ScrollAreaOptions{}, func() {
				ui.Column(c).FillWidth().Padding(0, style.ScrollbarGutter, style.SpaceMedium, 0).Gap(style.SpaceSmall).Children(func() {
					width, _ := c.Size()
					available := width - style.PlatformNavWidth - 2*style.PagePadding - style.ScrollbarGutter
					if showFolders {
						available -= style.LibraryListWidth + style.SpaceMedium
					}
					columns := videoColumns(available)
					ui.Grid(c).Columns(columns).FillWidth().GapX(style.SpaceMedium).GapY(style.DouyinMediaRowGap).Children(func() {
						for _, v := range s.videos.Data.Items {
							s.videoCard(c.Key(v.AwemeId), v, navigate)
						}
					})
					if len(s.videos.Data.Items) == 0 && !s.videos.Loading {
						empty(c, "暂无作品")
					}
					if s.videos.Loading || s.more.Loading {
						caption(c, "正在加载作品")
					}
					s.errorRow(c, s.more.Error, func() { s.loadVideos(true) })
					if s.videos.Data.More && (s.runtime.After == nil || s.autoFill >= 6) {
						button(c, "加载更多作品", input.ButtonOptions{Variant: input.Outline, Loading: s.more.Loading}).OnClick(func() { s.autoFill = 0; s.loadVideos(true) })
					}
				})
			}).Grow(1).MinHeight(0)
			s.maybeMore()
		})
	})
}
func (s *State) downloadPage(c *ui.Context) {
	panel(c, "抖音视频输入栏", func() {
		ui.Row(c).FillWidth().Padding(style.SpaceSmall).Gap(style.SpaceSmall).Children(func() {
			field(c.Key("link"), &s.link, "抖音视频链接", "视频 ID、链接或分享文本", s.parse).Grow(1)
			button(c, "解析", input.ButtonOptions{Variant: input.Primary, Loading: s.parsing, Disabled: s.api == nil || strings.TrimSpace(s.link) == ""}).OnClick(s.parse)
			button(c, "清空", input.ButtonOptions{Variant: input.Outline}).OnClick(func() { s.link = ""; s.parseSeq++; s.parsing = false })
		})
	})
	if len(s.staged) > 0 {
		panel(c, "抖音下载操作栏", func() {
			ui.Row(c).FillWidth().Padding(style.SpaceSmall).Gap(style.SpaceSmall).Children(func() {
				ui.Text(c, fmt.Sprintf("待下载 · %d", len(s.staged))).Grow(1)
				button(c, "开始下载", input.ButtonOptions{Variant: input.Primary, Disabled: s.backend == nil}).OnClick(func() { s.request(s.staged) })
				button(c, "清空列表", input.ButtonOptions{Variant: input.Ghost, Disabled: s.PendingDownloadCount() > 0}).OnClick(func() { s.staged = nil })
			})
		})
	}
	layout.ScrollArea(c.Key("queue"), s.scroll("queue"), layout.ScrollAreaOptions{}, func() {
		ui.Column(c).FillWidth().Padding(0, style.ScrollbarGutter, style.SpaceMedium, 0).Gap(style.SpaceMedium).Children(func() {
			for _, g := range s.groups {
				if g.Title == "" {
					continue
				}
				panel(c.Key(g.ID), "抖音批量："+g.Title, func() {
					var parts ui.CollapsibleParts
					ui.Row(c).FillWidth().Padding(style.SpaceSmall).Gap(style.SpaceSmall).Children(func() {
						parts = ui.CollapsibleBase(c.Key("open"), &g.Open)
						parts.Trigger.Label("展开抖音批量：" + g.Title).Grow(1).MinWidth(0).Gap(style.SpaceSmall).Children(func() {
							ui.Icon(c, icons.Must("chevron-right")).Size(16, 16).Rotate(90 * parts.Progress())
							ui.Text(c, g.Title).Grow(1).MinWidth(0).MaxLines(1).FontWeight(style.WeightSemibold)
						})
						if g.Status == "active" || g.Status == "queued" {
							button(c, "停止", input.ButtonOptions{Variant: input.Outline, Disabled: g.Stop}).OnClick(func() { s.stopGroup(g) })
						} else {
							if g.Failed > 0 || g.Status == "stopped" {
								button(c, "继续 / 重试", input.ButtonOptions{Variant: input.Primary}).OnClick(func() { s.resume(g) })
							}
							button(c, "关闭", input.ButtonOptions{Variant: input.Ghost}).OnClick(func() { s.dismiss(g) })
						}
					})
					ui.Row(c).FillWidth().Padding(0, style.SpaceSmall, style.SpaceSmall).Children(func() {
						caption(c, fmt.Sprintf("%s · 运行 %d · 完成 %d · 失败 %d", groupStatus(g), g.InFlight, g.Done, g.Failed))
					})
					parts.Panel(func() {
						ui.Column(c).FillWidth().Padding(style.SpaceSmall).Gap(style.SpaceSmall).Children(func() {
							s.errorRow(c, g.Error, nil)
							for _, v := range g.Items {
								if v.Status != "done" {
									s.downloadCard(c.Key(v.Data.AwemeId), v, true)
								}
							}
						})
					})
				})
			}
			for _, item := range s.staged {
				s.downloadCard(c.Key(item.Data.AwemeId), item, false)
			}
			if len(s.staged) == 0 && len(s.groups) == 0 {
				empty(c, "下载列表为空")
			}
		})
	}).Grow(1).MinHeight(0)
}
func groupStatus(g *downloadGroup) string {
	if g.Stop && g.Status == "active" {
		return "停止中"
	}
	return map[string]string{"queued": "排队中", "active": "下载中", "done": "已完成", "stopped": "已停止"}[g.Status]
}
func (s *State) downloadCard(c *ui.Context, item *queuedVideo, batch bool) {
	queued := item.Status == "active" || item.Status == "queued"
	panel(c, "抖音下载作品："+title(item.Data), func() {
		ui.Row(c).FillWidth().Padding(style.DownloadCardPadding).Gap(style.SpaceMedium).AlignItems(ui.Center).Children(func() {
			s.image(c, cover(item.Data), style.BatchCoverWidth, style.BatchCoverHeight, false)
			ui.Column(c).Grow(1).MinWidth(0).Gap(style.DownloadCardGap).Children(func() {
				ui.Row(c).FillWidth().Wrap().Gap(style.SpaceSmall).AlignItems(ui.Center).Children(func() {
					ui.Text(c, title(item.Data)).Grow(1).MinWidth(style.AuthorCardMinWidth).MaxLines(2).FontWeight(style.WeightSemibold)
					options := []input.SelectOption[int]{{Value: -1, Label: "默认画质"}}
					for i, track := range item.Data.Video.BitRate {
						if first(track.PlayAddr.UrlList) != "" {
							label := track.GearName
							if label == "" {
								label = fmt.Sprintf("%dP · %d kbps", track.PlayAddr.Height, track.BitRate/1000)
							}
							options = append(options, input.SelectOption[int]{Value: i, Label: label})
						}
					}
					if len(item.Data.Images) == 0 {
						input.Select(c.Key("quality"), &item.Track, options, input.SelectOptions{Label: "抖音下载画质", Disabled: queued || batch}).Element.Width(style.DouyinDownloadQualityWidth).Height(style.DownloadControlHeight).MinHeight(style.DownloadControlHeight).MinWidth(0).Radius(style.RadiusControl)
					} else {
						caption(c, fmt.Sprintf("图集 · %d 张", len(item.Data.Images)))
					}
					if item.Progress.Phase != "" {
						ui.Column(c.Key("progress")).Label("抖音下载进度").Width(style.DouyinDownloadProgressWidth).MinWidth(0).Children(func() { progressView(c, item.Progress) })
					}
					if !batch {
						label := "下载"
						if item.Status == "queued" {
							label = "已排队"
						} else if item.Status == "active" {
							label = "下载中"
						}
						button(c.Key("download"), label, input.ButtonOptions{Variant: input.Primary, Disabled: queued || item.CoverBusy || s.backend == nil}).Height(style.DownloadControlHeight).MinHeight(style.DownloadControlHeight).OnClick(func() { s.request([]*queuedVideo{item}) })
						button(c.Key("cover"), "封面", input.ButtonOptions{Variant: input.Ghost, Disabled: s.backend == nil || queued || item.CoverBusy, Loading: item.CoverBusy}).Height(style.DownloadControlHeight).MinHeight(style.DownloadControlHeight).OnClick(func() { s.saveCover(item) })
						button(c.Key("remove"), "移除", input.ButtonOptions{Variant: input.Ghost, Disabled: queued || item.CoverBusy}).Height(style.DownloadControlHeight).MinHeight(style.DownloadControlHeight).OnClick(func() { s.staged = slices.DeleteFunc(s.staged, func(v *queuedVideo) bool { return v == item }) })
					}
				})
				caption(c, item.Data.Author.Nickname)
				s.errorRow(c, item.Error, nil)
			})
		})
	})
}
func (s *State) historyPage(c *ui.Context) {
	if s.historySearch != s.historyQuery {
		s.loadHistory(false)
	}
	panel(c, "抖音历史工具栏", func() {
		ui.Row(c).FillWidth().Padding(style.SpaceSmall).Gap(style.SpaceSmall).Children(func() {
			ui.Text(c, fmt.Sprintf("下载历史 · %d", s.history.Data.Total)).Shrink(0).FontWeight(style.WeightSemibold)
			field(c.Key("search"), &s.historySearch, "搜索抖音下载历史", "搜索标题或作者", func() { s.loadHistory(false) }).Grow(1)
			button(c, "刷新历史", input.ButtonOptions{Variant: input.Outline, Loading: s.history.Loading, Disabled: s.backend == nil}).OnClick(func() { s.loadHistory(false) })
			button(c, "清空历史", input.ButtonOptions{Variant: input.Ghost, Disabled: s.history.Data.Total == 0 || s.historyBusy}).OnClick(func() { s.clearOpen = true })
		})
	})
	s.errorRow(c, s.history.Error, func() { s.loadHistory(false) })
	layout.ScrollArea(c.Key("history"), s.scroll("history"), layout.ScrollAreaOptions{}, func() {
		ui.Column(c).FillWidth().Gap(style.SpaceSmall).Children(func() {
			for _, v := range s.history.Data.Items {
				panel(c.Key(v.AwemeID), "抖音下载历史："+v.Title, func() {
					ui.Row(c).FillWidth().Padding(style.DownloadCardPadding).Gap(style.SpaceMedium).Children(func() {
						s.image(c, v.Cover, style.DownloadCoverWidth, style.DownloadCoverHeight, false)
						ui.Column(c).Grow(1).MinWidth(0).Gap(style.SpaceSmall).Children(func() {
							ui.Row(c).FillWidth().Gap(style.SpaceSmall).Children(func() {
								ui.Text(c, v.Title).Grow(1).MinWidth(0).MaxLines(2).FontWeight(style.WeightSemibold)
								button(c, "本地打开", input.ButtonOptions{Variant: input.Primary, Disabled: s.runtime.OpenFile == nil}).OnClick(func() { s.openFile(v.Path, false) })
								button(c, "打开目录", input.ButtonOptions{Variant: input.Outline, Disabled: s.runtime.OpenLocation == nil}).OnClick(func() { s.openFile(v.Path, true) })
								button(c, "删除记录", input.ButtonOptions{Variant: input.Ghost, Disabled: s.historyBusy}).OnClick(func() { s.changeHistory(func() error { return s.backend.DeleteDownloadHistory(v.AwemeID) }) })
							})
							caption(c, v.AuthorName+" · "+v.Downloaded)
							caption(c, v.Path)
						})
					})
				})
			}
			if len(s.history.Data.Items) == 0 && !s.history.Loading {
				empty(c, "暂无下载历史")
			}
			s.errorRow(c, s.historyMore.Error, func() { s.loadHistory(true) })
			if s.history.Data.HasMore {
				button(c, "加载更多历史", input.ButtonOptions{Variant: input.Outline, Loading: s.historyMore.Loading}).OnClick(func() { s.loadHistory(true) })
			}
		})
	}).Grow(1).MinHeight(0)
	overlay.Dialog(c.Key("clear"), &s.clearOpen, overlay.DialogOptions{Title: "清空抖音下载历史", Description: "清空下载记录？本地文件会保留。", HideClose: s.historyBusy, KeepOnEscape: s.historyBusy}, func() {
		ui.Row(c).Gap(style.SpaceSmall).Children(func() {
			button(c, "取消", input.ButtonOptions{Variant: input.Outline, Disabled: s.historyBusy}).OnClick(func() { s.clearOpen = false })
			button(c, "确认清空", input.ButtonOptions{Variant: input.Danger, Loading: s.historyBusy}).OnClick(func() { s.changeHistory(s.backend.ClearDownloadHistory) })
		})
	})
}
