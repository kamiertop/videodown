package api

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/kamiertop/videodown/utils"
)

var invalidFilenameChars = regexp.MustCompile(`[<>:"/\\|?*\x00-\x1f]`)

func streamBaseURL(v string) string {
	return strings.TrimSpace(v)
}

func sanitizeFilename(name string) string {
	t := strings.TrimSpace(name)
	if t == "" {
		return "video"
	}
	t = invalidFilenameChars.ReplaceAllString(t, "_")
	t = strings.Trim(t, " .")
	if t == "" {
		return "video"
	}
	return t
}

func sanitizeDirName(name string) string {
	t := strings.TrimSpace(name)
	if t == "" {
		return ""
	}
	t = invalidFilenameChars.ReplaceAllString(t, "_")
	t = strings.Trim(t, " .")
	return t
}

func uniqueFilePath(path string) string {
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		return path
	}
	ext := filepath.Ext(path)
	base := strings.TrimSuffix(path, ext)
	for i := 1; ; i++ {
		candidate := fmt.Sprintf("%s(%d)%s", base, i, ext)
		if _, err := os.Stat(candidate); errors.Is(err, os.ErrNotExist) {
			return candidate
		}
	}
}

func (b *BiliBili) downloadToFile(rawURL, targetPath, bvid, cookies string) error {
	u, err := url.Parse(rawURL)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") {
		return errors.New("流地址无效")
	}

	req, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set(UserAgent, userAgent())
	req.Header.Set(Referer, fmt.Sprintf("https://www.bilibili.com/video/%s", strings.TrimSpace(bvid)))
	req.Header.Set(Origin, BiliBiliUrl)
	req.Header.Set(Cookie, cookies)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("下载失败: %s", resp.Status)
	}

	f, err := os.Create(targetPath)
	if err != nil {
		return err
	}
	defer func() {
		_ = f.Close()
	}()

	if _, err = io.Copy(f, resp.Body); err != nil {
		return err
	}
	return nil
}

// DownloadVideoByDash 下载单个视频，videoURL/audioURL 来自 DASH 流地址。
func (b *BiliBili) DownloadVideoByDash(sourceName, bvid, title, videoURL, audioURL string) (string, error) {
	videoURL = streamBaseURL(videoURL)
	audioURL = streamBaseURL(audioURL)
	if videoURL == "" {
		return "", errors.New("视频流地址为空")
	}

	cookies, err := b.getCookies()
	if err != nil {
		return "", err
	}

	storagePath, err := b.settings.GetStorage()
	if err != nil {
		return "", err
	}

	targetDir := storagePath
	if group := sanitizeDirName(sourceName); group != "" {
		targetDir = filepath.Join(storagePath, group)
	}

	if err = os.MkdirAll(targetDir, 0o755); err != nil {
		return "", errors.New("创建下载目录失败")
	}

	fileName := sanitizeFilename(title)
	if strings.TrimSpace(bvid) == "" {
		bvid = "BV_UNKNOWN"
	}

	outPath := uniqueFilePath(filepath.Join(targetDir, fileName+".mp4"))

	tmpDir := filepath.Join(storagePath, ".tmp", fmt.Sprintf("%s-%d", bvid, time.Now().UnixNano()))
	if err = os.MkdirAll(tmpDir, 0o755); err != nil {
		return "", errors.New("创建临时目录失败")
	}
	defer func() {
		_ = os.RemoveAll(tmpDir)
	}()

	videoTmp := filepath.Join(tmpDir, "video.m4s")
	if err = b.downloadToFile(videoURL, videoTmp, bvid, cookies); err != nil {
		return "", err
	}

	ff := utils.NewFFmpeg()
	if audioURL == "" {
		if err = ff.Remux(videoTmp, outPath); err != nil {
			return "", err
		}
		return outPath, nil
	}

	audioTmp := filepath.Join(tmpDir, "audio.m4s")
	if err = b.downloadToFile(audioURL, audioTmp, bvid, cookies); err != nil {
		return "", err
	}
	if err = ff.Merge(videoTmp, audioTmp, outPath); err != nil {
		return "", err
	}

	return outPath, nil
}
