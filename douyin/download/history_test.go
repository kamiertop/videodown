package download

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"

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

// 分页 + 关键字过滤 + 缓存失效：翻页切片正确、搜索缩小 Total、删除后重建快照。
func TestDownloadHistoryPage(t *testing.T) {
	s := newTestService(t)
	dir := t.TempDir()

	for i := range 5 {
		title := "普通视频"
		if i%2 == 0 {
			title = "猫咪视频"
		}
		path := filepath.Join(dir, fmt.Sprintf("v%d.mp4", i))
		if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
			t.Fatalf("write video: %v", err)
		}
		s.markDownloaded(Task{AwemeID: fmt.Sprintf("id%d", i), Title: title, AuthorName: "作者甲"}, path, false, 0, kindVideo)
		// 拉开时间戳，保证倒序稳定等于写入逆序。
		time.Sleep(2 * time.Millisecond)
	}

	page1, err := s.DownloadHistoryPage(0, 2, "")
	if err != nil {
		t.Fatalf("page1: %v", err)
	}
	if page1.Total != 5 || len(page1.Items) != 2 || !page1.HasMore {
		t.Fatalf("page1 = total %d, %d items, hasMore %v; want 5, 2, true", page1.Total, len(page1.Items), page1.HasMore)
	}
	if page1.Items[0].AwemeID != "id4" {
		t.Fatalf("page1 first = %s, want newest id4", page1.Items[0].AwemeID)
	}

	page3, err := s.DownloadHistoryPage(4, 2, "")
	if err != nil {
		t.Fatalf("page3: %v", err)
	}
	if len(page3.Items) != 1 || page3.HasMore {
		t.Fatalf("page3 = %d items, hasMore %v; want 1, false", len(page3.Items), page3.HasMore)
	}

	search, err := s.DownloadHistoryPage(0, 50, "猫咪")
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if search.Total != 3 {
		t.Fatalf("search total = %d, want 3 (id0/id2/id4)", search.Total)
	}

	// 删除一条后缓存必须失效，Total 反映最新记录数。
	if err = s.DeleteDownloadHistory("id4"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	after, err := s.DownloadHistoryPage(0, 50, "")
	if err != nil {
		t.Fatalf("after delete: %v", err)
	}
	if after.Total != 4 {
		t.Fatalf("after delete total = %d, want 4", after.Total)
	}
}
