package bilibili

import (
	"testing"

	"github.com/egoist/mygo/ui"
	"github.com/kamiertop/videodown/bilibili/model"
	"github.com/kamiertop/videodown/internal/gui/style"
)

// TestFollowCardOpensAuthorDetail pins the regression where the fixed-toolbar
// branch intercepted /bilibili/up/{mid} and re-rendered the follows list:
// clicking a card must leave the list and show the author page.
func TestFollowCardOpensAuthorDetail(t *testing.T) {
	s := NewState()
	s.auth.loggedIn = true
	s.follows.list.Loaded = true
	s.follows.list.Data = model.FollowData{Total: 1, List: []model.FollowDataList{
		{Mid: 42, Uname: "测试UP主", Sign: "签名"},
	}}
	path := "/bilibili/up"
	render := func(c *ui.Context) {
		style.UseTheme(c)
		s.Layout(c, path, func(target string) { path = target }, func() {
			s.Form(c, path, func(target string) { path = target }, func() {})
		})
	}
	test := ui.NewTester(render, 1200, 880)
	test.Frame()
	if !test.HasText("我的关注") {
		t.Fatal("follows list missing")
	}
	if err := test.Click("进入测试UP主的主页"); err != nil {
		t.Fatal(err)
	}
	test.Frame()
	if path != "/bilibili/up/42" {
		t.Fatalf("card did not navigate, path=%q", path)
	}
	// 资料未加载时应显示 UP 主页的加载态，而不是回到关注列表。
	if !test.HasText("正在加载 UP 主视频") {
		t.Fatal("author detail route not rendered")
	}
	if test.HasText("我的关注 · 1") {
		t.Fatal("follows list re-rendered on the detail route")
	}
	// 资料加载完成后渲染详情页。
	s.author.info.Loaded = true
	s.author.info.Data = model.UserInfoData{Mid: 42, Name: "测试UP主"}
	test.Frame()
	if !test.HasText("返回") || !test.HasText("投稿") {
		t.Fatal("author detail page missing after profile load")
	}
}
