package download

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/kamiertop/videodown/douyin/model"
	"github.com/kamiertop/videodown/internal/constant"
	"github.com/kamiertop/videodown/utils"
)

// DownloadCover 从多个抖音封面候选中选择最高质量的一张并保存到作者目录。
func (d *Service) DownloadCover(covers []model.Cover, task Task) (string, error) {
	coverURL := bestDouyinCoverURL(covers)
	if coverURL == "" {
		return "", errors.New("封面地址无效")
	}

	storagePath, err := d.store.StoragePath()
	if err != nil {
		return "", err
	}
	targetDir, err := d.resolveDownloadDir(storagePath, task)
	if err != nil {
		return "", err
	}
	if err = os.MkdirAll(targetDir, 0o755); err != nil {
		return "", errors.New("创建下载目录失败")
	}

	headers, err := d.publicHeaders()
	if err != nil {
		return "", err
	}
	headers[constant.Accept] = "image/avif,image/webp,image/apng,image/svg+xml,image/*,*/*;q=0.8"

	resp, err := d.client.R().
		SetHeaders(headers).
		Get(coverURL)
	if err != nil {
		return "", err
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return "", fmt.Errorf("封面下载失败: %s", resp.Status)
	}

	fileName := utils.FileName(task.Title)
	if fileName == "" {
		fileName = "cover"
	}
	ext := utils.ImageExtFromResponse(coverURL, resp.Response)
	outPath := utils.UniqueFilePath(filepath.Join(targetDir, fileName+ext))
	if err = os.WriteFile(outPath, resp.Bytes(), 0o644); err != nil {
		return "", err
	}
	if task.AwemeID != "" {
		task.AwemeID += ":cover"
		d.markDownloaded(task, outPath, false, 0, kindCover)
	}

	return outPath, nil
}

func bestDouyinCoverURL(covers []model.Cover) string {
	for _, cover := range covers {
		for _, rawURL := range cover.UrlList {
			rawURL = normalizeDouyinHTTPURL(rawURL)
			if !strings.HasPrefix(rawURL, "http://") && !strings.HasPrefix(rawURL, "https://") {
				continue
			}
			return rawURL
		}
	}
	return ""
}
