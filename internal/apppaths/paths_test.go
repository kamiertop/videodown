package apppaths

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInsideRejectsEscapesAndSymlinks(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	for _, path := range []string{"../outside", outside} {
		if _, err := Inside(root, path); err == nil {
			t.Fatalf("accepted outside path %q", path)
		}
	}
	if got, err := Inside(root, "download/new"); err != nil || got != filepath.Join(root, "download/new") {
		t.Fatalf("inside path: %s %v", got, err)
	}
	if err := os.Symlink(outside, filepath.Join(root, "link")); err != nil {
		t.Skip(err)
	}
	if _, err := Inside(root, "link/new/file"); err == nil {
		t.Fatal("accepted external symlink")
	}
}
