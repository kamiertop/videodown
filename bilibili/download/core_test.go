package download

import (
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/imroc/req/v3"
	"github.com/kamiertop/videodown/internal/storage"
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
	if err = s.store.UpdateConfig(func(c *storage.Config) error { c.Download.SkipDownloaded = false; return nil }); err != nil {
		t.Fatalf("set skipDownloaded: %v", err)
	}
	_, skipped, err = s.downloadDashTask(task)
	if err == nil || skipped {
		t.Fatalf("incremental off: skipped=%v err=%v, want download attempt (err != nil, skipped=false)", skipped, err)
	}
}

func TestNativeProgressListenerReceivesUpdates(t *testing.T) {
	s := newTestService(t)
	var received []Progress
	s.SetProgressListener(func(p Progress) { received = append(received, p) })
	s.emitProgress(Progress{Bvid: "BV1xx411c7mD", Cid: 42, Phase: "video", Percent: 35})
	s.emitProgress(Progress{Bvid: "BV1xx411c7mD", Cid: 42, Phase: "done", Percent: 95})
	if len(received) != 2 || received[0].Percent != 35 || received[1].Percent != 100 {
		t.Fatalf("progress not delivered to native listener: %+v", received)
	}
}

// A broken response body must remain an error even when both files close cleanly.
type failedDownloadReader struct{}

func (failedDownloadReader) Read([]byte) (int, error) { return 0, io.ErrUnexpectedEOF }

func TestDownloadToFilePreservesReadError(t *testing.T) {
	s := newTestService(t)
	s.client.GetTransport().WrapRoundTripFunc(func(http.RoundTripper) req.HttpRoundTripFunc {
		return func(r *http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: 200, Status: "200 OK", Header: make(http.Header), Request: r, ContentLength: 1024,
				Body: io.NopCloser(io.MultiReader(strings.NewReader("incomplete"), failedDownloadReader{}))}, nil
		}
	})
	err := s.downloadToFile("https://example.com/video", filepath.Join(t.TempDir(), "video.tmp"), "BV1xx411c7mD", "测试视频", 42, "video", "", 0, 100)
	if !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatalf("truncated download error = %v", err)
	}
}
