package utils

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestFFmpegUsesCustomExecutableForDetectionAndDownload(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("uses a shell fixture")
	}
	dir := t.TempDir()
	custom := filepath.Join(dir, "custom-ffmpeg")
	output := filepath.Join(dir, "result.m4a")
	script := "#!/bin/sh\nif [ \"$1\" = '-version' ]; then echo 'ffmpeg version fixture'; exit 0; fi\nfor arg do last=\"$arg\"; done\nprintf fixture > \"$last\"\n"
	if err := os.WriteFile(custom, []byte(script), 0700); err != nil {
		t.Fatal(err)
	}
	path, version, err := CheckFFmpeg(custom)
	if err != nil || path != custom || version != "ffmpeg version fixture" {
		t.Fatalf("detection: %s %s %v", path, version, err)
	}
	if err := NewFFmpeg(custom).Remux("input", output); err != nil {
		t.Fatal(err)
	}
	if content, err := os.ReadFile(output); err != nil || string(content) != "fixture" {
		t.Fatal("custom executable was not used")
	}
	if _, _, err := CheckFFmpeg(filepath.Join(dir, "missing")); err == nil {
		t.Fatal("invalid custom path fell back to system executable")
	}
}
