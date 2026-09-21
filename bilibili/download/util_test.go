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
	return NewService(logger.New(), store, nil, func() (string, error) { return "", nil })
}

func TestDownloadedVideoKeys(t *testing.T) {
	s := newTestService(t)

	dir := t.TempDir()
	existing := filepath.Join(dir, "exists.mp4")
	if err := os.WriteFile(existing, []byte("x"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	// BV1：单 P 已下载且文件在；BV2：历史有但文件被删；BV3：封面记录不算视频；BV4：多 P 只下过 P1。
	s.markDownloaded(Task{Bvid: "BV1", Cid: 100, Title: "a"}, existing, kindVideo)
	s.markDownloaded(Task{Bvid: "BV2", Cid: 200, Title: "b"}, filepath.Join(dir, "deleted.mp4"), kindVideo)
	s.markDownloaded(Task{Bvid: "BV3", Cid: -300, Title: "c"}, existing, kindCover)
	s.markDownloaded(Task{Bvid: "BV4", Cid: 400, Title: "d"}, existing, kindVideo)

	keys := []DownloadKey{
		{Bvid: "BV1", Cid: 100}, // 精确 cid 命中
		{Bvid: "BV1", Cid: 0},   // 仅 bvid，任意分 P 命中
		{Bvid: " bv1 ", Cid: 0}, // 大小写与空白归一化后命中
		{Bvid: "BV2", Cid: 0},   // 文件被删，不算已下载
		{Bvid: "BV3", Cid: 0},   // 封面记录不算
		{Bvid: "BV4", Cid: 400}, // 已下载的分 P
		{Bvid: "BV4", Cid: 401}, // 未下载的分 P，不能因同 bvid 其他 P 已下载而误判
		{Bvid: "BV5", Cid: 0},   // 无历史
		{Bvid: "", Cid: 0},      // 空 bvid
	}

	got, err := s.DownloadedVideoKeys(keys)
	if err != nil {
		t.Fatalf("DownloadedVideoKeys: %v", err)
	}
	if want := []int{0, 1, 2, 5}; !slices.Equal(got, want) {
		t.Fatalf("matched indices = %v, want %v", got, want)
	}
}

// 全部带 cid 的查询走 isDownloaded 点查路径，不扫描下载历史。
func TestDownloadedVideoKeysAllWithCid(t *testing.T) {
	s := newTestService(t)

	dir := t.TempDir()
	existing := filepath.Join(dir, "exists.mp4")
	if err := os.WriteFile(existing, []byte("x"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	s.markDownloaded(Task{Bvid: "BV1", Cid: 100, Title: "a"}, existing, kindVideo)
	s.markDownloaded(Task{Bvid: "BV2", Cid: 200, Title: "b"}, filepath.Join(dir, "deleted.mp4"), kindVideo)

	got, err := s.DownloadedVideoKeys([]DownloadKey{
		{Bvid: "BV1", Cid: 100},
		{Bvid: "BV2", Cid: 200},
		{Bvid: "BV3", Cid: 300},
	})
	if err != nil {
		t.Fatalf("DownloadedVideoKeys: %v", err)
	}
	if want := []int{0}; !slices.Equal(got, want) {
		t.Fatalf("matched indices = %v, want %v", got, want)
	}
}

func TestDownloadedVideoKeysEmpty(t *testing.T) {
	s := newTestService(t)
	got, err := s.DownloadedVideoKeys(nil)
	if err != nil {
		t.Fatalf("DownloadedVideoKeys(nil): %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("expected no matches, got %v", got)
	}
}
