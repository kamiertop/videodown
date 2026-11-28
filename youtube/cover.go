package youtube

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/kamiertop/videodown/internal/mediafile"
)

// DownloadCover saves the original thumbnail independently of video downloads.
func (s *Service) DownloadCover(ctx context.Context, video Video) (string, error) {
	s.active.Add(1)
	defer s.active.Add(-1)

	id, err := VideoID(video.ID)
	if err != nil {
		return "", err
	}
	fallback := "https://i.ytimg.com/vi/" + id + "/hqdefault.jpg"
	data, ext, err := s.readCover(ctx, video.Cover)
	if err != nil && video.Cover != fallback && ctx.Err() == nil {
		data, ext, err = s.readCover(ctx, fallback)
	}
	if err != nil {
		return "", err
	}
	if err := ctx.Err(); err != nil {
		return "", err
	}

	dir := s.mediaDirectory(s.store.Config().Download, video.Author)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("创建封面目录失败: %w", err)
	}
	name := mediafile.FileName(video.Title)
	if name == "" || len(name) > mediafile.MaxFileNameBytes {
		name = id
	}

	temp, err := os.CreateTemp(dir, ".youtube-cover-*.tmp")
	if err != nil {
		return "", err
	}
	defer os.Remove(temp.Name())
	_, writeErr := temp.Write(data)
	if err := errors.Join(writeErr, temp.Close()); err != nil {
		return "", err
	}
	path, err := mediafile.PublishFile(temp.Name(), filepath.Join(dir, name+ext))
	if err != nil {
		return "", err
	}

	payload, _ := json.Marshal(History{
		Kind: "cover", ID: id, Title: video.Title, Author: video.Author,
		Path: path, Downloaded: time.Now().UTC().Format(time.RFC3339),
	})
	if err := s.store.SaveHistory("youtube:downloaded:"+id+":cover", string(payload)); err != nil {
		return path, fmt.Errorf("封面已保存，但历史记录写入失败: %w", err)
	}
	return path, nil
}

func (s *Service) readCover(ctx context.Context, rawURL string) ([]byte, string, error) {
	u, err := url.Parse(rawURL)
	if err != nil || u.Scheme != "https" || u.User != nil ||
		!(u.Hostname() == "img.youtube.com" || strings.HasSuffix(u.Hostname(), ".ytimg.com")) {
		return nil, "", errors.New("无效的 YouTube 封面地址")
	}
	resp, err := s.requestClient().R().SetContext(ctx).SetHeader("Referer", "https://www.youtube.com/").Get(u.String())
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("下载封面失败: HTTP %d", resp.StatusCode)
	}
	const limit = 8 << 20
	data, err := io.ReadAll(io.LimitReader(resp.Body, limit+1))
	if err != nil {
		return nil, "", err
	}
	if len(data) > limit {
		return nil, "", errors.New("YouTube 封面超过大小限制")
	}
	var ext string
	switch http.DetectContentType(data) {
	case "image/jpeg":
		ext = ".jpg"
	case "image/png":
		ext = ".png"
	case "image/webp":
		ext = ".webp"
	case "image/gif":
		ext = ".gif"
	default:
		return nil, "", errors.New("YouTube 返回了无法保存的封面格式")
	}
	return data, ext, nil
}
