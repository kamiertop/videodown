package api

import (
	"path/filepath"
	"testing"

	"github.com/kamiertop/videodown/logger"
	"github.com/kamiertop/videodown/utils"
)

func TestResolveDownloadDirGroupsByAuthor(t *testing.T) {
	settings := utils.NewSettingsWithMemory(logger.New())
	defer func() {
		_ = settings.Close()
	}()
	d := &Douyin{settings: settings}
	root := t.TempDir()

	tests := []struct {
		name string
		task DouyinDownloadTask
		want string
	}{
		{
			name: "manual parse",
			task: DouyinDownloadTask{AuthorName: "作者 A", SourceKind: "解析结果", SourceName: ""},
			want: filepath.Join(root, "作者_A"),
		},
		{
			name: "favorite video",
			task: DouyinDownloadTask{AuthorName: "作者 A", SourceKind: "收藏视频", SourceName: "收藏视频"},
			want: filepath.Join(root, "作者_A"),
		},
		{
			name: "favorite collection",
			task: DouyinDownloadTask{AuthorName: "作者 A", SourceKind: "收藏合集", SourceName: "喜欢的合集"},
			want: filepath.Join(root, "作者_A", "喜欢的合集"),
		},
		{
			name: "user collection",
			task: DouyinDownloadTask{AuthorName: "作者 A", SourceKind: "用户合集", SourceName: "作品合集"},
			want: filepath.Join(root, "作者_A", "作品合集"),
		},
		{
			name: "user works",
			task: DouyinDownloadTask{AuthorName: "作者 A", SourceKind: "用户作品", SourceName: "作者 A"},
			want: filepath.Join(root, "作者_A"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := d.resolveDownloadDir(root, tt.task)
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Fatalf("target dir = %q, want %q", got, tt.want)
			}
		})
	}
}
