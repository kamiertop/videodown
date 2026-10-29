package gui

import (
	"image/png"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
)

func TestFilenameOrderInsertionTargets(t *testing.T) {
	for _, tc := range []struct {
		name   string
		source string
		target func(*ui.Tester) (float32, float32)
		want   string
	}{
		{"right half", "{title}", func(tt *ui.Tester) (float32, float32) {
			r, _ := tt.Find("排序字段 {author}")
			return r.X + r.W - 3, r.Y + r.H/2
		}, "{author} - {title} - {id}"},
		{"left half", "{id}", func(tt *ui.Tester) (float32, float32) {
			r, _ := tt.Find("排序字段 {author}")
			return r.X + 3, r.Y + r.H/2
		}, "{title} - {id} - {author}"},
		{"gap", "{title}", func(tt *ui.Tester) (float32, float32) {
			r, _ := tt.Find("排序字段 {author}")
			return r.X + r.W + 4, r.Y + r.H/2
		}, "{author} - {title} - {id}"},
		{"trailing whitespace", "{title}", func(tt *ui.Tester) (float32, float32) {
			r, _ := tt.Find("排序字段 {id}")
			return r.X + r.W + 60, r.Y + r.H/2
		}, "{author} - {id} - {title}"},
		{"leading padding", "{id}", func(tt *ui.Tester) (float32, float32) {
			r, _ := tt.Find("文件名字段排序区")
			return r.X + 3, r.Y + r.H/2
		}, "{id} - {title} - {author}"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := NewShell()
			s.pages.Settings.Config.Filename = "{title} - {author} - {id}"
			s.router.Push("/settings/filename")
			tt := ui.NewTester(s.view, 1200, 1100)
			tt.Advance(300 * time.Millisecond)
			source, ok := tt.Find("排序字段 " + tc.source)
			if !ok {
				t.Fatal("missing drag source")
			}
			x, y := tc.target(tt)
			tt.Press(source.X+source.W/2, source.Y+source.H/2)
			tt.Move(source.X+source.W/2+8, source.Y+source.H/2)
			tt.Move(x, y)
			tt.Release(x, y)
			tt.Advance(300 * time.Millisecond)
			if got := s.pages.Settings.Config.Filename; got != tc.want {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
			// Removing a keyed field after reordering must not retain stale drawing callbacks.
			if err := tt.Click("{author}"); err != nil {
				t.Fatal(err)
			}
			tt.Advance(300 * time.Millisecond)
		})
	}
}

func TestFilenameOrderCancelAndBusy(t *testing.T) {
	s := NewShell()
	original := "{title} - {author}"
	s.pages.Settings.Config.Filename = original
	s.router.Push("/settings/filename")
	tt := ui.NewTester(s.view, 1200, 1100)
	tt.Advance(300 * time.Millisecond)
	source, _ := tt.Find("排序字段 {title}")
	target, _ := tt.Find("排序字段 {author}")
	tt.Press(source.X+10, source.Y+10)
	tt.Move(target.X+target.W+40, target.Y+10)
	tt.Key(0, ui.KeyEscape)
	tt.Release(target.X+target.W+40, target.Y+10)
	tt.Advance(300 * time.Millisecond)
	if s.pages.Settings.Config.Filename != original {
		t.Fatal("Escape did not cancel sorting")
	}
	s.settingsBusy = true
	tt.Frame()
	tt.Press(source.X+10, source.Y+10)
	tt.Move(target.X+target.W+40, target.Y+10)
	tt.Release(target.X+target.W+40, target.Y+10)
	tt.Advance(300 * time.Millisecond)
	if s.pages.Settings.Config.Filename != original {
		t.Fatal("busy form accepted a drag")
	}
}

func TestFilenameOrderWrappedRowEnd(t *testing.T) {
	s := NewShell()
	tokens := []string{"{title}", "{source}", "{author}", "{folder}", "{collection}", "{publish_date}", "{date}", "{time}", "{id}", "{author_id}"}
	s.pages.Settings.Config.Filename = strings.Join(tokens, " - ")
	s.router.Push("/settings/filename")
	tt := ui.NewTester(s.view, 760, 1100)
	tt.Advance(300 * time.Millisecond)
	first, _ := tt.Find("排序字段 " + tokens[0])
	end := 0
	for i, token := range tokens {
		r, ok := tt.Find("排序字段 " + token)
		if !ok {
			t.Fatalf("missing field %s", token)
		}
		if r.Y == first.Y {
			end = i
		}
	}
	if end == len(tokens)-1 {
		t.Fatal("test did not wrap the fields")
	}
	source, _ := tt.Find("排序字段 " + tokens[len(tokens)-1])
	region, _ := tt.Find("文件名字段排序区")
	x, y := region.X+region.W-3, first.Y+first.H/2
	tt.Press(source.X+10, source.Y+10)
	tt.Move(source.X+20, source.Y+10)
	tt.Move(x, y)
	// Keep a snapshot of the insertion marker before dropping.
	file, err := os.Create("/tmp/filename-order-drag.png")
	if err != nil {
		t.Fatal(err)
	}
	err = png.Encode(file, tt.Image())
	file.Close()
	if err != nil {
		t.Fatal(err)
	}
	tt.Release(x, y)
	tt.Advance(300 * time.Millisecond)
	expected := append([]string(nil), tokens[:end+1]...)
	expected = append(expected, tokens[len(tokens)-1])
	expected = append(expected, tokens[end+1:len(tokens)-1]...)
	if got := s.pages.Settings.Config.Filename; got != strings.Join(expected, " - ") {
		t.Fatalf("wrapped drop got %q, want %q", got, strings.Join(expected, " - "))
	}
}
