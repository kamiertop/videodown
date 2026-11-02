package douyin

import (
	"image/png"
	"os"
	"testing"
	"time"

	"github.com/ELY-Open-Source/MujicaUI/core"
	"github.com/egoist/mygo/ui"
)

func TestPortraitCardsAndUnifiedLibraryToolbar(t *testing.T) {
	s, _, _, r := setup(t)
	path := "/douyin/favorite/folder/7"
	navigate := func(p string) { path = p }
	test := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		ui.Column(c).Fill().Children(func() { s.Layout(c, path, navigate, func() { s.Form(c, path, navigate, func() {}) }) })
	}, 1200, 800)
	test.Frame()
	r.flush(t)
	test.Frame()
	test.Advance(time.Second)
	cover, ok := test.Find("抖音作品封面")
	if !ok || cover.H <= cover.W || cover.W/cover.H < .59 || cover.W/cover.H > .61 {
		t.Fatalf("cover should be 3:5: %+v", cover)
	}
	toolbar, ok := test.Find("抖音视频工具栏")
	if !ok || toolbar.H > 60 {
		t.Fatalf("toolbar should fit one compact row: %+v", toolbar)
	}
	var rowY float32
	for i, label := range []string{"收藏视频", "收藏夹", "合集", "刷新视频", "全选", "下载已选 (0)", "一键下载全部"} {
		bounds, ok := test.Find(label)
		if !ok || bounds.Y < toolbar.Y || bounds.Y+bounds.H > toolbar.Y+toolbar.H {
			t.Fatalf("control outside shared toolbar: %s %+v", label, bounds)
		}
		if i == 0 {
			rowY = bounds.Y
		} else if bounds.Y != rowY {
			t.Fatalf("control on extra row: %s %+v", label, bounds)
		}
	}
	if _, ok := test.Find("抖音内容分类"); ok {
		t.Fatal("obsolete category panel remains")
	}
	test.SetSize(760, 600)
	test.Frame()
	test.Advance(time.Second)
	for _, label := range []string{"收藏视频", "收藏夹", "合集", "下载已选 (0)", "一键下载全部"} {
		bounds, ok := test.Find(label)
		if !ok || bounds.X+bounds.W > 760 {
			t.Fatalf("narrow toolbar clipped: %s %+v", label, bounds)
		}
	}
}

func TestCompactCardFitsTwoLineTitleAndSVGPhotoBadge(t *testing.T) {
	s, _, _, _ := setup(t)
	v := fixture("live-photo")
	v.Desc = "这是一个足够长的抖音视频标题，用来检查两行标题是否完整展示"
	v.IsLivePhoto = 1
	v.CreateTime = 1760000000
	test := ui.NewTester(func(c *ui.Context) {
		core.Use(c, core.Settings{})
		ui.Column(c).Width(170).Children(func() { s.videoCard(c, v, func(string) {}) })
	}, 220, 500)
	test.Frame()
	badge, ok := test.Find("动图")
	if !ok || badge.W != 24 || badge.H != 24 {
		t.Fatalf("missing compact SVG media badge: %+v", badge)
	}
	f, err := os.Create("/tmp/douyin-compact-live-photo.png")
	if err != nil {
		t.Fatal(err)
	}
	if err := png.Encode(f, test.Image()); err != nil {
		f.Close()
		t.Fatal(err)
	}
	f.Close()
	if test.HasText("Live Photo") {
		t.Fatal("text badge should be replaced by the SVG icon")
	}
	title, ok := test.Find(v.Desc)
	if !ok || title.H < 36 {
		t.Fatalf("two-line title clipped: %+v", title)
	}
	author, ok := test.Find("进入抖音作者主页：测试作者")
	if !ok {
		t.Fatal("missing author")
	}
	date, ok := test.Find("发布 " + time.Unix(int64(v.CreateTime), 0).Format("2006-01-02"))
	if !ok {
		t.Fatal("missing publish date")
	}
	if author.Y-(title.Y+title.H) > 4 || date.Y-(author.Y+author.H) > 3 {
		t.Fatalf("metadata spacing too large: title=%+v author=%+v date=%+v", title, author, date)
	}
}
