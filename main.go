package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/egoist/mygo"
	biliapi "github.com/kamiertop/videodown/bilibili/api"
	bilidownload "github.com/kamiertop/videodown/bilibili/download"
	dyapi "github.com/kamiertop/videodown/douyin/api"
	dydownload "github.com/kamiertop/videodown/douyin/download"
	"github.com/kamiertop/videodown/internal/gui"
	"github.com/kamiertop/videodown/internal/gui/assets"
	"github.com/kamiertop/videodown/internal/storage"
	"github.com/kamiertop/videodown/internal/updater"
	"github.com/kamiertop/videodown/logger"
)

func main() {
	dataDir := flag.String("data-dir", "", "应用数据目录（包含 config.toml 和 history.sqlite）")
	flag.Parse()
	if err := Run(*dataDir); err != nil {
		log.Print(err)
		os.Exit(1)
	}
}
func Run(dataDir string) (err error) {
	if dataDir == "" {
		executable, e := os.Executable()
		if e != nil {
			return e
		}
		cwd, e := os.Getwd()
		if e != nil {
			return e
		}
		dataDir = defaultDataDir(executable, cwd, os.TempDir())
	}
	dataDir, err = filepath.Abs(dataDir)
	if err != nil {
		return err
	}
	store, err := openAppData(dataDir)
	if err != nil {
		return fmt.Errorf("打开应用配置 %s: %w", dataDir, err)
	}
	defer func() { err = errors.Join(err, store.Close()) }()
	vdlog := logger.New()
	service := biliapi.New(vdlog, store)
	downloader := bilidownload.New(vdlog, store, service.CookieFunc())
	dyService := dyapi.New(vdlog, store)
	dyDownloader := dydownload.New(vdlog, store, dyService.PublicHeaders())
	shell := gui.NewShellWithOptions(gui.Options{UpdateChecker: updater.New(vdlog, store), Preferences: store, Settings: store, Bilibili: service, BilibiliDownload: downloader, Douyin: dyService, DouyinDownload: dyDownloader, ActiveDownloads: func() int { return downloader.ActiveCount() + dyDownloader.ActiveCount() }})
	defer shell.Wait()
	mygo.App.WhenReady(shell.OpenWindow)

	return mygo.App.Run()
}

// Portable installations keep data beside the executable.
// go run places its binary in go-build*/.../exe; persist outside that temp tree.
func defaultDataDir(executable, cwd, tempDir string) string {
	rel, err := filepath.Rel(tempDir, executable)
	if err == nil && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) && !filepath.IsAbs(rel) {
		first, _, _ := strings.Cut(rel, string(filepath.Separator))
		if strings.HasPrefix(first, "go-build") {
			return cwd
		}
	}

	return filepath.Dir(executable)
}

// openAppData creates first-run files without replacing existing preferences or icons.
func openAppData(directory string) (*storage.Store, error) {
	store, err := storage.Open(directory)
	if err != nil {
		return nil, err
	}
	if err = store.InitDefaults("./download"); err == nil {
		err = ensureAppIcon(directory)
	}
	if err != nil {
		_ = store.Close()
		return nil, err
	}
	return store, nil
}
func ensureAppIcon(directory string) error {
	f, err := os.OpenFile(filepath.Join(directory, "appicon.png"), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
	if errors.Is(err, os.ErrExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("创建应用图标: %w", err)
	}
	_, writeErr := f.Write(assets.AppIconPNG)
	return errors.Join(writeErr, f.Close())
}
