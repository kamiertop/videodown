package utils

import "testing"

func TestApplyFilenameTemplate(t *testing.T) {
	got := ApplyFilenameTemplate("{author} - {title} - {folder}", map[string]string{"author": "作者", "title": "标题", "folder": ""})
	if got != "作者 - 标题" {
		t.Fatalf("unexpected filename: %q", got)
	}
}

func TestSupportedFilenameTemplate(t *testing.T) {
	if !SupportedFilenameTemplate("{title}") || SupportedFilenameTemplate("plain name") {
		t.Fatal("template validation failed")
	}
}

func TestFilenameSafety(t *testing.T) {
	if got := ApplyFilenameTemplate("{title} - {folder}", map[string]string{"title": "CON", "folder": ""}); got != "_CON" {
		t.Fatalf("unexpected safe filename: %q", got)
	}
	if got := ApplyFilenameTemplate("{title} {unknown}", map[string]string{"title": "video"}); got != "video" {
		t.Fatalf("unexpected unknown token result: %q", got)
	}
}
