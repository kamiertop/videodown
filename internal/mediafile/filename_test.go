package mediafile

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestFileName(t *testing.T) {
	rawName := `
4K音乐 尽在/
BiliBili\ #4k #music
`
	fmt.Println(FileName(rawName))
}

func TestUniqueFilePath(t *testing.T) {
	dir := t.TempDir()
	first := filepath.Join(dir, "video.mp4")
	if got := UniqueFilePath(first); got != first {
		t.Fatalf("UniqueFilePath(%q) = %q, want unchanged", first, got)
	}
	if err := os.WriteFile(first, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(dir, "video(1).mp4")
	if got := UniqueFilePath(first); got != want {
		t.Fatalf("UniqueFilePath(%q) = %q, want %q", first, got, want)
	}
}

func TestUniqueFilePathTooLongNameReturns(t *testing.T) {
	// 300 字节的文件名超过 NAME_MAX，stat 返回 ENAMETOOLONG 而非“不存在”；
	// 修复前 UniqueFilePath 会在这里无限循环，把下载 worker 和前端一起卡死。
	dir := t.TempDir()
	base := filepath.Join(dir, strings.Repeat("a", 300))
	done := make(chan string, 1)
	go func() { done <- UniqueFilePath(base + ".mp4") }()
	select {
	case got := <-done:
		if want := base + "(1).mp4"; got != want {
			t.Fatalf("UniqueFilePath() = %q, want %q", got, want)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("UniqueFilePath did not return for over-long file name (infinite loop)")
	}
}

func TestHistoryPathExists(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "video.mp4")
	if got, err := HistoryPathExists(file); got || err != nil {
		t.Errorf("HistoryPathExists(missing file) = %v, %v; want false, nil", got, err)
	}
	if err := os.WriteFile(file, []byte("video"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got, err := HistoryPathExists(file); !got || err != nil {
		t.Errorf("HistoryPathExists(existing file) = %v, %v; want true, nil", got, err)
	}
	loop := filepath.Join(dir, "loop")
	if err := os.Symlink(loop, loop); err != nil {
		t.Skipf("cannot create symlink: %v", err)
	}
	var pathErr *os.PathError
	if got, err := HistoryPathExists(loop); got || !errors.As(err, &pathErr) {
		t.Errorf("HistoryPathExists(symlink loop) = %v, %v; want false and a wrapped path error", got, err)
	}
}
