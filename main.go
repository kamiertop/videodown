package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/egoist/mygo"

	biliapi "github.com/kamiertop/videodown/bilibili/api"
	bilidownload "github.com/kamiertop/videodown/bilibili/download"
	dyapi "github.com/kamiertop/videodown/douyin/api"
	dydownload "github.com/kamiertop/videodown/douyin/download"
	"github.com/kamiertop/videodown/internal/apppaths"
	"github.com/kamiertop/videodown/internal/gui"
	"github.com/kamiertop/videodown/internal/storage"
	"github.com/kamiertop/videodown/internal/updater"
	"github.com/kamiertop/videodown/logger"
	yt "github.com/kamiertop/videodown/youtube"
)

func main() {
	flag.Parse()
	if err := Run(); err != nil {
		log.Print(err)
		os.Exit(1)
	}
}

func Run() (err error) {
	dataDir, err := apppaths.Directory()
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
	shell := gui.NewShellWithOptions(gui.Options{
		YouTube:          yt.New(store),
		UpdateChecker:    updater.New(vdlog, store),
		Preferences:      store,
		Settings:         store,
		Bilibili:         service,
		BilibiliDownload: downloader,
		Douyin:           dyService,
		DouyinDownload:   dyDownloader,
		ActiveDownloads:  func() int { return downloader.ActiveCount() + dyDownloader.ActiveCount() },
	})
	defer shell.Wait()
	mygo.App.WhenReady(shell.OpenWindow)

	return mygo.App.Run()
}

// openAppData creates first-run data files without replacing existing preferences.
func openAppData(directory string) (*storage.Store, error) {
	store, err := storage.Open(directory)
	if err != nil {
		return nil, err
	}
	if err = store.InitDefaults("./download"); err != nil {
		_ = store.Close()
		return nil, err
	}

	return store, nil
}
