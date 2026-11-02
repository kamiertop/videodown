package settings

import (
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/ELY-Open-Source/MujicaUI/core"
	"github.com/ELY-Open-Source/MujicaUI/input"
	"github.com/egoist/mygo/ui"
	"github.com/kamiertop/videodown/internal/gui/components"
	"github.com/kamiertop/videodown/internal/gui/preferences"
	"github.com/kamiertop/videodown/internal/gui/style"
	"github.com/kamiertop/videodown/utils"
)

var filenameTokens = regexp.MustCompile(`\{(?:title|source|author|folder|collection|publish_date|date|time|id|author_id)\}`)
var filenameFields = []struct{ token, label string }{
	{"{title}", "视频标题"}, {"{author}", "作者名称"}, {"{folder}", "收藏夹名称"}, {"{collection}", "合集名称"}, {"{id}", "视频 ID"}, {"{author_id}", "作者 ID"}, {"{publish_date}", "发布时间"}, {"{date}", "下载日期"}, {"{time}", "下载时间"},
}

func (s *State) filenameSettings(c *ui.Context, w Services) {
	settingCard(c, "", "file-text", func() {
		ui.Text(c, "选择文件名内容").FontWeight(style.WeightSemibold)
		ui.Grid(c).FillWidth().Columns(2).Gap(8).Children(func() {
			for _, field := range filenameFields {
				selected := strings.Contains(s.Config.Filename, field.token)
				variant := input.Outline
				marker := "○ "
				if selected {
					variant = input.Secondary
					marker = "✓ "
				}
				components.Button(c.Key(field.token), marker+field.label+"  "+field.token, input.ButtonOptions{Variant: variant, Disabled: w.SettingsBusy}).Radius(style.RadiusMenu).Label(field.token).MinWidth(0).Clip().Justify(ui.Start).OnClick(func() { s.toggleFilenameField(field.token) })
			}
		})
		ui.Text(c, "当前顺序（拖动调整）").FontWeight(style.WeightSemibold)
		s.filenameOrder(c, w.SettingsBusy)
		ui.Text(c, "字段之间的分隔符").FontWeight(style.WeightSemibold)
		ui.Row(c).FillWidth().Wrap().Gap(8).Children(func() {
			for _, item := range []struct{ value, label string }{{" - ", "短横线"}, {"_", "下划线"}, {" ", "空格"}, {" | ", "竖线"}} {
				variant := input.Outline
				if s.separator == item.value {
					variant = input.Secondary
				}
				components.Button(c.Key(item.label), item.label, input.ButtonOptions{Variant: variant, Disabled: w.SettingsBusy}).Radius(style.RadiusMenu).OnClick(func() {
					s.separator = item.value
					s.Config.Filename = strings.Join(filenameTokens.FindAllString(s.Config.Filename, -1), item.value)
				})
			}
		})
		previewBox(c, "文件名预览", filenamePreview(s.Config.Filename))
		components.Muted(c, "超长文件名处理：超过 200 字节时自动改用视频 ID 命名；抖音标题尾部的话题标签会先去除。")
		ui.Row(c).FillWidth().Gap(12).AlignItems(ui.Center).Children(func() {
			ui.Text(c, "直接编辑模板").Shrink(0).FontWeight(style.WeightSemibold)
			field := input.TextInput(c.Key("filename"), &s.Config.Filename, input.TextInputOptions{Label: "文件名模板", Placeholder: "{title}", Disabled: w.SettingsBusy})
			field.Element.Grow(1).MinWidth(0)
			field.Input.OnSubmit(func() { s.saveFilename(w) })
		})
		ui.Column(c).FillWidth().Padding(16).Gap(16).Radius(style.RadiusPanel).Border(style.BorderWidth, core.Tokens(c).Border).Children(func() {
			insetToggle(c, "目录分组规则", "选择下载文件的目录组织方式", &s.Config.Grouping, w.SettingsBusy, func() {})
			ui.Row(c).FillWidth().Wrap().Gap(8).Children(func() {
				for _, rule := range []struct{ id, label string }{{"none", "不分组"}, {"author", "按作者"}, {"author_source", "作者+来源"}, {"source", "来源"}, {"date", "下载日期"}} {
					variant := input.Outline
					if s.Config.GroupingRule == rule.id {
						variant = input.Secondary
					}
					components.Button(c.Key(rule.id), rule.label, input.ButtonOptions{Variant: variant, Disabled: !s.Config.Grouping || w.SettingsBusy}).Radius(style.RadiusMenu).OnClick(func() { s.Config.GroupingRule = rule.id })
				}
			})
			previewBox(c, "示例路径", s.groupedPreview())
		})
		components.Button(c.Key("save-template"), "保存模板", input.ButtonOptions{Loading: w.SettingsBusy}).Radius(style.RadiusMenu).FillWidth().OnClick(func() { s.saveFilename(w) })
	})
}
func (s *State) saveFilename(w Services) {
	if w.SaveSetting != nil {
		w.SaveSetting(preferences.Changes{Filename: preferences.Value(s.Config.Filename), Grouping: preferences.Value(s.Config.Grouping), GroupingRule: preferences.Value(s.Config.GroupingRule)})
	}
}
func (s *State) toggleFilenameField(token string) {
	parts := filenameTokens.FindAllString(s.Config.Filename, -1)
	if index := slices.Index(parts, token); index >= 0 {
		parts = slices.Delete(parts, index, index+1)
	} else {
		parts = append(parts, token)
	}
	s.Config.Filename = strings.Join(parts, s.separator)
}

func previewBox(c *ui.Context, label, value string) {
	ui.Column(c).FillWidth().Padding(16).Gap(8).Radius(style.RadiusPanel).Background(core.Tokens(c).Background).Border(style.BorderWidth, core.Tokens(c).Border).Children(func() { components.Muted(c, label); ui.Text(c, value).FillWidth().TextColor(core.Tokens(c).AccentText) })
}
func filenamePreview(template string) string {
	return utils.ApplyFilenameTemplate(template, map[string]string{"title": "示例视频", "author": "作者", "id": "BV1xx", "folder": "收藏夹", "collection": "合集", "author_id": "123456", "publish_date": "2026-10-09", "date": "2026-10-25", "time": "14-35-22"}) + ".mp4"
}
func (s *State) groupedPreview() string {
	group := ""
	if s.Config.Grouping {
		switch s.Config.GroupingRule {
		case "author":
			group = "作者名"
		case "source":
			group = "来源名"
		case "author_source":
			group = filepath.Join("作者名", "来源名")
		case "date":
			group = "下载日期"
		}
	}
	return filepath.Join(s.Config.Directory, group, filenamePreview(s.Config.Filename))
}
