package api

import (
	"path/filepath"
	"testing"

	"github.com/kamiertop/videodown/internal/storage"
)

func TestResolveTargetDirGroupsByAuthor(t *testing.T) {
	store, err := storage.OpenMemory()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		_ = store.Close()
	}()
	if err := store.SetSavePreference(true); err != nil {
		t.Fatal(err)
	}
	b := &BiliBili{store: store}
	root := t.TempDir()

	tests := []struct {
		name string
		task DashDownloadTask
		want string
	}{
		{
			name: "plain video",
			task: DashDownloadTask{UpperName: "UP 主", SourceKind: "关注动态", SourceName: "关注动态"},
			want: filepath.Join(root, "UP_主"),
		},
		{
			name: "favorite video",
			task: DashDownloadTask{UpperName: "UP 主", SourceKind: "收藏夹", SourceName: "我的收藏"},
			want: filepath.Join(root, "UP_主", "我的收藏"),
		},
		{
			name: "collection video",
			task: DashDownloadTask{UpperName: "UP 主", SourceKind: "合集", SourceName: "课程合集"},
			want: filepath.Join(root, "UP_主", "课程合集"),
		},
		{
			name: "series video",
			task: DashDownloadTask{UpperName: "UP 主", SourceKind: "系列", SourceName: "更新系列"},
			want: filepath.Join(root, "UP_主", "更新系列"),
		},
		{
			name: "part video",
			task: DashDownloadTask{UpperName: "UP 主", SourceKind: "分P", SourceName: "多 P 标题"},
			want: filepath.Join(root, "UP_主"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := b.resolveTargetDir(root, tt.task)
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Fatalf("target dir = %q, want %q", got, tt.want)
			}
		})
	}
}
