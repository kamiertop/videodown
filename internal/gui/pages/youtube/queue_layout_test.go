package youtube

import (
	"testing"
	"time"

	"github.com/ELY-Open-Source/MujicaUI/core"
	"github.com/egoist/mygo/ui"
)

func TestQueueCardsKeepControlsBesideCover(t *testing.T) {
	for _, width := range []int{760, 1200, 1800} {
		s, _, _ := newQueueTest()
		video := queueVideo("jNQXAC9IVRw")
		video.Title = "【AI动态】每日科技新闻 | GPT5数学突破陶哲轩 | Opera Neon浏览器问题 | Gemini将识别成猫 | 字树发布仿生机器人 | 这是一段用于验证下载卡片布局的长标题"
		video.Author = "最佳拍档"
		video.Duration = 10*time.Minute + 45*time.Second
		s.enqueue(video, 137, false)
		video.ID = "KWGhxyu5osA"
		video.Title = "第二个视频"
		s.enqueue(video, 136, false)
		scroll := &ui.ScrollState{}
		test := ui.NewTester(func(c *ui.Context) {
			core.Use(c, core.Settings{})
			s.Layout(c, "/youtube/download", func(string) {}, func() { s.Form(c, "/youtube/download", scroll) })
		}, width, 700)
		test.Frame()
		var previous ui.Rect
		for _, id := range []string{"1", "2"} {
			row, _ := test.Find("YouTube 队列任务：" + id)
			cover, _ := test.Find("任务封面：" + id)
			info, _ := test.Find("任务信息：" + id)
			title, _ := test.Find("任务标题：" + id)
			quality, _ := test.Find("任务画质：" + id)
			actions, _ := test.Find("任务操作：" + id)
			start, _ := test.Find("下载任务：" + id)
			if row.H > 112 || cover.W != 160 || info.X < cover.X+cover.W || quality.X+quality.W > actions.X || actions.X+actions.W > float32(width) || start.Y+start.H > cover.Y+cover.H+1 || quality.Y+quality.H > cover.Y+cover.H+1 {
				t.Fatalf("queue card overlaps or is too tall at %d: row=%+v cover=%+v quality=%+v actions=%+v", width, row, cover, quality, actions)
			}
			if id == "2" && (row.H != previous.H) {
				t.Fatal("title length changed card height")
			}
			if title.H > 24 || title.X+title.W < actions.X+actions.W-1 || title.X+title.W > row.X+row.W {
				t.Fatalf("title must use the full right area on one line at %d: title=%+v actions=%+v", width, title, actions)
			}
			previous = row
		}
		if width == 1200 {
			saveQueueScreenshot(t, test, "/tmp/youtube-queue-layout-wide.png")
		}
	}
}
