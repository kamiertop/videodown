package apppaths

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/kamiertop/videodown/utils"
)

func TestDirectory(t *testing.T) {
	original := *customDir
	t.Cleanup(func() {
		*customDir = original
	})

	t.Run("自定义绝对路径", func(t *testing.T) {
		dir := t.TempDir()
		*customDir = dir

		got, err := Directory()
		utils.NoError(t, err)
		utils.Equal(t, dir, got)
	})

	t.Run("自定义相对路径", func(t *testing.T) {
		*customDir = "./testdata"

		want, err := filepath.Abs(*customDir)
		utils.NoError(t, err)

		got, err := Directory()
		utils.NoError(t, err)

		utils.Equal(t, want, got)
	})

	t.Run("默认目录", func(t *testing.T) {
		*customDir = ""

		executable, err := os.Executable()
		utils.NoError(t, err)

		executable, err = filepath.EvalSymlinks(executable)
		utils.NoError(t, err)

		want := filepath.Dir(executable)

		got, err := Directory()
		utils.NoError(t, err)

		utils.Equal(t, want, got)
	})
}

func TestInsideRejectsEscapesAndSymlinks(t *testing.T) {
	root := t.TempDir()
	outside := t.TempDir()
	for _, path := range []string{"../outside", outside} {
		if _, err := Inside(root, path); err == nil {
			t.Fatalf("accepted outside path %q", path)
		}
	}
	// Inside canonicalizes the root. On macOS the temporary directory may
	// itself be reached through a symlink, so compare canonical paths.
	canonicalRoot, err := filepath.EvalSymlinks(root)
	utils.NoError(t, err)
	if got, err := Inside(root, "download/new"); err != nil || got != filepath.Join(canonicalRoot, "download/new") {
		t.Fatalf("inside path: %s %v", got, err)
	}
	if err := os.Symlink(outside, filepath.Join(root, "link")); err != nil {
		t.Skip(err)
	}
	if _, err := Inside(root, "link/new/file"); err == nil {
		t.Fatal("accepted external symlink")
	}
}

func TestInsideWithSymlinkedRoot(t *testing.T) {
	root := t.TempDir()
	alias := filepath.Join(t.TempDir(), "root-link")
	if err := os.Symlink(root, alias); err != nil {
		t.Skip(err)
	}
	canonicalRoot, err := filepath.EvalSymlinks(root)
	utils.NoError(t, err)
	got, err := Inside(alias, "download/new")
	utils.NoError(t, err)
	utils.Equal(t, filepath.Join(canonicalRoot, "download/new"), got)
	outside := t.TempDir()
	if err := os.Symlink(outside, filepath.Join(root, "escape")); err != nil {
		t.Fatal(err)
	}
	if _, err := Inside(alias, "escape/new/file"); err == nil {
		t.Fatal("symlinked root allowed an escape through another symlink")
	}
	if _, err := Inside(alias, "../outside"); err == nil {
		t.Fatal("symlinked root allowed a parent directory escape")
	}
}
