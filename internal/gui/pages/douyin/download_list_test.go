package douyin

import (
	"fmt"
	"testing"

	"github.com/ELY-Open-Source/MujicaUI/core"
	"github.com/egoist/mygo/ui"
	"github.com/kamiertop/videodown/douyin/model"
)

func TestDownloadListGroupAndStagedTasks(t *testing.T) {
	s := NewState()
	group := &downloadGroup{ID: 1, Title: "测试分组", Status: "stopped", Error: "批量任务出错"}
	for i := range 60 {
		group.Items = append(group.Items, &queuedVideo{Data: model.AwemeItem{
			AwemeId: fmt.Sprint(i), ItemTitle: fmt.Sprintf("批量作品 %d", i),
		}})
	}
	group.Items[0].Status = "done"
	s.groups = []*downloadGroup{group}
	s.stage(model.AwemeItem{AwemeId: "40", ItemTitle: "单独暂存作品"})
	test := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		s.downloadPage(c)
	}, 1200, 720)
	test.Frame()
	if test.HasText("批量作品 1") || !test.HasText("单独暂存作品") {
		t.Fatal("collapsed group should show no children and leave the staged task visible")
	}
	if err := test.Click("展开抖音批量：测试分组"); err != nil {
		t.Fatal(err)
	}
	test.Frame()
	if !group.Open || !test.HasText("批量作品 1") || test.HasText("批量作品 0") || !test.HasText(group.Error) {
		t.Fatal("expanded group should show its error and unfinished items")
	}
	s.downloadList.ScrollTo(41, ui.Start)
	test.Frame()
	first, _ := s.downloadList.Visible()
	if first != 41 || !test.HasText("批量作品 40") {
		t.Fatalf("first visible row = %d, want the batch task at 41", first)
	}
	group.Items[1].Status = "done"
	test.Frame()
	first, _ = s.downloadList.Visible()
	if first != 40 || !test.HasText("批量作品 40") {
		t.Fatalf("first visible row = %d, want the same task at 40", first)
	}
	s.downloadList.ScrollTo(0, ui.Start)
	test.Frame()
	if err := test.Click("收起抖音批量：测试分组"); err != nil {
		t.Fatal(err)
	}
	test.Frame()
	if group.Open || test.HasText(group.Error) || !test.HasText("单独暂存作品") {
		t.Fatal("collapse should hide group details and reveal the staged task")
	}
	if err := test.Click("移除"); err != nil {
		t.Fatal(err)
	}
	if len(s.staged) != 0 {
		t.Fatal("remove did not clear the staged task")
	}
}
