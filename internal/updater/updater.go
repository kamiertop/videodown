// Package updater checks for application updates without blocking startup.
package updater

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/imroc/req/v3"
	"github.com/kamiertop/videodown/internal/storage"
	"github.com/kamiertop/videodown/logger"
	"github.com/wailsapp/wails/v3/pkg/application"
	"golang.org/x/mod/semver"
)

const (
	latestReleaseURL = "https://api.github.com/repos/kamiertop/videodown/releases/latest"
	checkTimeout     = 15 * time.Second
	UnSupportedOS    = "不支持的操作系统或架构"
)

// Version is set by build flags, for example:
// -X github.com/kamiertop/videodown/internal/updater.Version=1.0.0
var Version string

type Result struct {
	Available      bool   `json:"available"`      // 是否有新版本可用
	CurrentVersion string `json:"currentVersion"` // 当前版本
	LatestVersion  string `json:"latestVersion"`  // 最新版本
	DownloadURL    string `json:"downloadURL"`    // 下载链接
	ReleaseNotes   string `json:"releaseNotes"`   // 发布说明
}

type Updater struct {
	logger *logger.Logger
	store  *storage.Store
	events *application.EventManager
	client *req.Client

	resultMu sync.RWMutex
	result   *Result
}

func New(log *logger.Logger, store *storage.Store, events *application.EventManager) *Updater {
	return &Updater{
		logger: log.WithName("Updater"),
		store:  store,
		events: events,
		client: req.C().SetCommonRetryCount(2),
	}
}

// ServiceStartup starts the network check in the background and returns
// immediately so application startup is never delayed by GitHub.
func (u *Updater) ServiceStartup(ctx context.Context, _ application.ServiceOptions) error {
	go u.checkOnStartup(ctx)

	return nil
}

func (u *Updater) checkOnStartup(parent context.Context) {
	u.logger.Debugf("starting update check, current version: %s", Version)
	enabled, err := u.store.AutoUpdate()
	if err != nil {
		u.logger.Errorf("read auto-update setting: %v", err)
		return
	}
	if !enabled {
		u.logger.Infof("auto-update is disabled, skipping check")
		return
	}

	ctx, cancel := context.WithTimeout(parent, checkTimeout)
	defer cancel()

	result, err := u.check(ctx)
	if err != nil {
		if !errors.Is(err, context.Canceled) {
			u.logger.Errorf("check update: %v", err)
		}
		return
	}
	u.setResult(result)
	if result.Available && u.events != nil {
		u.events.Emit("update-available", result)
	}
}

func (u *Updater) check(ctx context.Context) (Result, error) {
	resp, err := u.client.
		R().
		SetContext(ctx).
		SetHeaders(map[string]string{
			"Accept":     "application/vnd.github+json",
			"User-Agent": "videodown/" + Version,
		}).
		Get(latestReleaseURL)
	if err == nil && resp.StatusCode == http.StatusOK {
		var release releaseResponse
		if err = resp.Into(&release); err != nil {
			u.logger.Debugf("decode GitHub API response failed: %v", err)
		} else if !semver.IsValid(release.TagName) {
			u.logger.Debugf("GitHub API returned invalid version: %q", release.TagName)
		} else {
			u.logger.Infof("latest release: %s, current version: %s", release.TagName, Version)
			return Result{
				Available:      semver.Compare(release.TagName, Version) > 0,
				CurrentVersion: Version,
				LatestVersion:  release.TagName,
				DownloadURL:    ParseDownloadURL(release.TagName),
				ReleaseNotes:   release.Body,
			}, nil
		}
	}

	latestVersion, redirectErr := u.parseVersionByRedirect(ctx)
	if redirectErr != nil {
		return Result{}, fmt.Errorf("API check failed and redirect parse failed: %w", redirectErr)
	}
	if !semver.IsValid(latestVersion) {
		return Result{}, fmt.Errorf("invalid release version %q", latestVersion)
	}
	return Result{
		Available:      semver.Compare(latestVersion, Version) > 0,
		CurrentVersion: Version,
		LatestVersion:  latestVersion,
		DownloadURL:    ParseDownloadURL(latestVersion),
	}, nil
}

func (u *Updater) GetVersion() string {
	return Version
}

// NeedUpdate performs an explicit update check for frontend callers.
func (u *Updater) NeedUpdate() (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), checkTimeout)
	defer cancel()
	result, err := u.check(ctx)
	if err != nil {
		return false, err
	}
	u.setResult(result)

	return result.Available, nil
}

// InstallUpdate starts the standalone updater and exits the current process.
func (u *Updater) InstallUpdate(packagePath string) error {
	if packagePath == "" {
		return errors.New("更新包路径为空")
	}
	if _, err := os.Stat(packagePath); err != nil {
		return fmt.Errorf("更新包不存在: %w", err)
	}
	current, err := os.Executable()
	if err != nil {
		return fmt.Errorf("获取当前程序路径失败: %w", err)
	}
	name := "videodown-updater"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	updaterPath := filepath.Join(filepath.Dir(current), name)
	if _, err := os.Stat(updaterPath); err != nil {
		return fmt.Errorf("更新程序不存在: %w", err)
	}
	cmd := exec.Command(updaterPath, "--package", packagePath, "--target", current)
	cmd.Dir = filepath.Dir(current)
	logPath := filepath.Join(filepath.Dir(current), "videodown-updater.log")
	if logFile, logErr := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644); logErr == nil {
		cmd.Stdout, cmd.Stderr = logFile, logFile
	}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("启动更新程序失败: %w", err)
	}
	go func() { application.Get().Quit() }()
	return nil
}

// DownloadUpdate downloads an update package and emits progress events.
func (u *Updater) DownloadUpdate(ctx context.Context, downloadURL string) (string, error) {
	// Download URLs from GitHub Releases normally return a 302 to the asset;
	// explicitly use the normal redirect policy in case the client was used for
	// a no-redirect version lookup previously.
	resp, err := u.client.Clone().SetRedirectPolicy(req.DefaultRedirectPolicy()).R().SetContext(ctx).Get(downloadURL)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download returned HTTP %d", resp.StatusCode)
	}
	ext := ".tar.gz"
	if strings.HasSuffix(downloadURL, ".zip") {
		ext = ".zip"
	}
	// Keep the package beside the running application so users can inspect it
	// and the standalone updater can access it reliably.
	appPath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("获取当前程序路径失败: %w", err)
	}
	f, err := os.CreateTemp(filepath.Dir(appPath), "videodown-update-*"+ext)
	if err != nil {
		return "", err
	}
	defer f.Close()
	total := resp.ContentLength
	var downloaded int64
	buf := make([]byte, 32*1024)
	for {
		n, readErr := resp.Body.Read(buf)
		if n > 0 {
			if _, err = f.Write(buf[:n]); err != nil {
				return "", err
			}
			downloaded += int64(n)
			if u.events != nil {
				u.events.Emit("update-download-progress", map[string]int64{"downloaded": downloaded, "total": total})
			}
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return "", readErr
		}
	}
	return f.Name(), nil
}

func (u *Updater) parseVersionByRedirect(ctx context.Context) (string, error) {
	// Clone so the no-redirect policy does not affect subsequent downloads.
	resp, err := u.client.Clone().
		SetTimeout(checkTimeout).
		SetRedirectPolicy(req.NoRedirectPolicy()).
		R().
		SetContext(ctx).
		Get("https://github.com/kamiertop/videodown/releases/latest")
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusFound {
		return "", errors.New("访问GitHub最新版本页面失败，可能是网络问题或GitHub访问限制")
	}
	location, err := url.Parse(resp.GetHeader("Location"))
	if err != nil {
		return "", err
	}
	// 返回版本
	return path.Base(location.Path), nil
}

// ParseDownloadURL 根据版本生成下载链接
func ParseDownloadURL(version string) string {
	switch runtime.GOOS {
	case "windows":
		return fmt.Sprintf("https://github.com/kamiertop/videodown/releases/download/%s/videodown-windows-amd64.zip", version)
	case "darwin":
		return fmt.Sprintf("https://github.com/kamiertop/videodown/releases/download/%s/videodown-macos-universal.tar.gz", version)
	case "linux":
		return fmt.Sprintf("https://github.com/kamiertop/videodown/releases/download/%s/videodown-linux-amd64.tar.gz", version)
	default:
		return UnSupportedOS
	}
}

func (u *Updater) setResult(result Result) {
	u.resultMu.Lock()
	defer u.resultMu.Unlock()
	u.result = &result
}

// LatestResult lets the frontend recover a result if the startup event was
// emitted before its event listener was registered.
func (u *Updater) LatestResult() *Result {
	u.resultMu.RLock()
	defer u.resultMu.RUnlock()
	if u.result == nil {
		return nil
	}
	result := *u.result
	return &result
}
