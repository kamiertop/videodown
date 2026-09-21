package download

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/kamiertop/videodown/internal/storage"
	"github.com/kamiertop/videodown/logger"
)

func newTestService(t *testing.T) *Service {
	t.Helper()
	store, err := storage.OpenMemory()
	if err != nil {
		t.Fatalf("open memory store: %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })
	return New(logger.New(), store, nil, func() (map[string]string, error) { return nil, nil })
}

// 图文作品按目录判重、视频按文件判重、封面是独立记录：增量下载不会重复下载已存在的内容。
func TestDownloadedAwemeIDs(t *testing.T) {
	s := newTestService(t)

	dir := t.TempDir()
	// 图文：落盘为一个目录（kindAlbum，Path=dir）。
	albumDir := filepath.Join(dir, "album")
	if err := os.MkdirAll(albumDir, 0o755); err != nil {
		t.Fatalf("mkdir album: %v", err)
	}
	s.markDownloaded(Task{AwemeID: "a1", Title: "图文"}, albumDir, true, 3, kindAlbum)
	// 视频：落盘为文件。
	videoFile := filepath.Join(dir, "v.mp4")
	if err := os.WriteFile(videoFile, []byte("x"), 0o644); err != nil {
		t.Fatalf("write video: %v", err)
	}
	s.markDownloaded(Task{AwemeID: "v1", Title: "视频"}, videoFile, false, 0, kindVideo)
	// 封面：独立 :cover 键，不影响视频判重。
	s.markDownloaded(Task{AwemeID: "v2", Title: "封面"}, filepath.Join(dir, "c.jpg"), false, 0, kindCover)
	// 历史有但目录被删：视为未下载，会重新下载。
	s.markDownloaded(Task{AwemeID: "a2", Title: "被删的图文"}, filepath.Join(dir, "deleted"), true, 2, kindAlbum)

	got, err := s.DownloadedAwemeIDs([]string{"a1", "v1", "v2", "a2", "a1", "  ", ""})
	if err != nil {
		t.Fatalf("DownloadedAwemeIDs: %v", err)
	}
	if want := []string{"a1", "v1"}; !slices.Equal(got, want) {
		t.Fatalf("downloaded = %v, want %v", got, want)
	}
}
