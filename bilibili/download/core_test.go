package download

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/kamiertop/videodown/internal/constant"
)

// 增量开关控制已下载跳过：开启（含未设置时的默认开启）时缓存命中直接跳过、不发起网络请求；
// 关闭后即使缓存命中也继续走下载流程，这里以“视频流地址为空”报错证明没有命中跳过分支。
func TestDownloadDashTaskRespectsIncrementalSwitch(t *testing.T) {
	s := newTestService(t)

	// 造一条“已下载且文件仍在”的缓存记录。
	file := filepath.Join(t.TempDir(), "v.mp4")
	if err := os.WriteFile(file, []byte("x"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	task := Task{Bvid: "BV1xx411c7mD", Cid: 1001, Title: "已下载视频", VideoURL: ""}
	s.markDownloaded(task, file, kindVideo)

	// 开关未设置时默认开启：VideoURL 为空也能命中跳过（未发起网络请求）。
	path, skipped, err := s.downloadDashTask(task)
	if err != nil || !skipped || path != file {
		t.Fatalf("incremental on: path=%q skipped=%v err=%v, want (%q, true, nil)", path, skipped, err, file)
	}

	// 关闭增量下载：不再跳过，老老实实重新下载。
	if err = s.store.Set(constant.SkipDownloadedKey, "false"); err != nil {
		t.Fatalf("set skipDownloaded: %v", err)
	}
	_, skipped, err = s.downloadDashTask(task)
	if err == nil || skipped {
		t.Fatalf("incremental off: skipped=%v err=%v, want download attempt (err != nil, skipped=false)", skipped, err)
	}
}
