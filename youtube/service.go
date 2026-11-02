// Package youtube implements video, audio and playlist downloads without an external downloader.
package youtube

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/kamiertop/videodown/internal/storage"
	"github.com/kamiertop/videodown/utils"
	"github.com/kamiertop/videodown/youtube/internal/extractor"
)

type Format struct {
	ID      int
	Label   string
	Summary string
	Height  int
	Audio   bool
	Mime    string
}
type Video struct {
	ID, Title, Author, Cover string
	Duration                 time.Duration
	Formats                  []Format
	AudioFormats             []Format
}
type Progress struct {
	Phase             string
	Downloaded, Total int64
	Percent           float64
	Speed             float64
}
type History struct {
	ID         string `json:"id"`
	Title      string `json:"title"`
	Author     string `json:"authorName"`
	Path       string `json:"path"`
	Downloaded string `json:"downloaded"`
	Kind       string `json:"kind,omitempty"`
}

var errAuthenticatedPlayback = errors.New("YouTube 已确认账号登录，但当前解析器未通过播放器校验")

type cachedVideo struct {
	client *extractor.Client
	video  *extractor.Video
	at     time.Time
}

type Service struct {
	parseMu   sync.Mutex
	publishMu sync.Mutex
	cache     map[string]cachedVideo
	store     *storage.Store
	transport http.RoundTripper
}

func New(store *storage.Store) *Service { return &Service{store: store} }
func (s *Service) DownloadConcurrency() int {
	return max(1, min(10, s.store.Config().Download.Concurrency))
}
func (s *Service) Cookie() string { return s.store.Config().YouTube.Cookie }
func (s *Service) SetCookie(value string) error {
	value = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(value), "Cookie:"))
	if strings.ContainsAny(value, "\r\n") {
		return errors.New("请粘贴单行 Cookie 请求头")
	}
	err := s.store.UpdateConfig(func(c *storage.Config) error { c.YouTube.Cookie = value; return nil })
	if err == nil {
		s.parseMu.Lock()
		s.cache = nil
		s.parseMu.Unlock()
	}
	return err
}

var videoIDPattern = regexp.MustCompile(`^[A-Za-z0-9_-]{11}$`)

func VideoID(input string) (string, error) {
	input = strings.TrimSpace(input)
	if videoIDPattern.MatchString(input) {
		return input, nil
	}
	u, err := url.Parse(input)
	if err != nil || (u.Scheme != "https" && u.Scheme != "http") {
		return "", errors.New("请输入有效的 YouTube 视频链接")
	}
	host := strings.ToLower(u.Hostname())
	var id string
	switch host {
	case "youtu.be":
		id = strings.Trim(u.Path, "/")
	case "youtube.com", "www.youtube.com", "m.youtube.com", "music.youtube.com":
		if u.Path == "/watch" {
			id = u.Query().Get("v")
		} else {
			parts := strings.Split(strings.Trim(u.Path, "/"), "/")
			if len(parts) == 2 && (parts[0] == "shorts" || parts[0] == "embed" || parts[0] == "live") {
				id = parts[1]
			}
		}
	}
	if !videoIDPattern.MatchString(id) {
		return "", errors.New("不支持的 YouTube 链接，请使用单个视频链接")
	}
	return id, nil
}
func (s *Service) client() *extractor.Client {
	return &extractor.Client{Cookie: s.Cookie(), HTTPClient: &http.Client{Timeout: 30 * time.Second, Transport: s.transport}}
}
func describe(v *extractor.Video) Video {
	out := Video{ID: v.ID, Title: v.Title, Author: v.Author, Duration: v.Duration}
	if len(v.Thumbnails) > 0 {
		for _, thumb := range v.Thumbnails {
			if thumb.URL != "" {
				out.Cover = thumb.URL
			}
		}
	}
	if out.Cover == "" && v.ID != "" {
		out.Cover = "https://i.ytimg.com/vi/" + v.ID + "/hqdefault.jpg"
	}
	for _, f := range v.Formats {
		if !strings.HasPrefix(f.MimeType, "video/") || (f.URL == "" && f.Cipher == "") {
			continue
		}
		label := formatLabel(f)

		out.Formats = append(out.Formats, Format{ID: f.ItagNo, Label: label, Summary: formatSummary(f), Height: f.Height, Audio: f.AudioChannels > 0, Mime: f.MimeType})
	}
	for _, f := range selectableAudioFormats(v.Formats) {
		out.AudioFormats = append(out.AudioFormats, audioFormat(*f))
	}
	sort.SliceStable(out.Formats, func(i, j int) bool { return out.Formats[i].Height > out.Formats[j].Height })
	return out
}

// formatLabel exposes the stream characteristics rather than just its container.
func formatLabel(f extractor.Format) string {
	media, params, _ := mime.ParseMediaType(f.MimeType)
	codec := params["codecs"]
	for _, entry := range strings.Split(codec, ",") {
		entry = strings.TrimSpace(entry)
		switch {
		case strings.HasPrefix(entry, "avc1"), strings.HasPrefix(entry, "avc3"):
			codec = "H.264"
		case strings.HasPrefix(entry, "av01"):
			codec = "AV1"
		case strings.HasPrefix(entry, "vp09"), strings.HasPrefix(entry, "vp9"):
			codec = "VP9"
		case strings.HasPrefix(entry, "hev1"), strings.HasPrefix(entry, "hvc1"):
			codec = "H.265"
		default:
			continue
		}
		break
	}
	if codec == "" {
		codec = "编码未知"
	}
	label := fmt.Sprintf("%dp · %s · %s", f.Height, codec, strings.ToUpper(strings.TrimPrefix(media, "video/")))
	if f.FPS > 0 {
		label += fmt.Sprintf(" · %d帧", f.FPS)
	}
	bitrate := f.AverageBitrate
	if bitrate <= 0 {
		bitrate = f.Bitrate
	}
	if bitrate > 0 {
		label += fmt.Sprintf(" · %.2f Mbps", float64(bitrate)/1e6)
	}
	if f.AudioChannels > 0 {
		label += " · 含音频"
	} else {
		label += " · 需合并音频"
	}
	// The itag distinguishes variants even when YouTube reports identical metadata.
	return label + fmt.Sprintf(" · #%d", f.ItagNo)
}

func (s *Service) Resolve(ctx context.Context, input string) (Video, error) {
	id, err := VideoID(input)
	if err != nil {
		return Video{}, err
	}
	_, raw, err := s.fetchVideo(ctx, id)
	if err != nil {
		return Video{}, s.parseError(err)
	}
	v := describe(raw)
	if len(v.Formats) == 0 && len(v.AudioFormats) == 0 {
		return Video{}, errors.New("未找到可下载的视频格式，可能需要登录或播放器规则已变化")
	}
	return v, nil
}
func (s *Service) History() ([]History, error) {
	rows, err := s.store.HistoryRows("youtube")
	if err != nil {
		return nil, err
	}
	var result []History
	for _, row := range rows {
		var item History
		if json.Unmarshal([]byte(row), &item) == nil {
			result = append(result, item)
		}
	}
	return result, nil
}
func (s *Service) Download(ctx context.Context, video Video, formatID int, notify func(Progress)) (string, error) {
	var progressMu sync.Mutex
	var phase string
	var sampleAt time.Time
	var sampleBytes int64
	var speed float64
	emit := func(p Progress) {
		progressMu.Lock()
		defer progressMu.Unlock()
		now := time.Now()
		if p.Phase != phase {
			phase = p.Phase
			sampleAt = now
			sampleBytes = 0
			speed = 0
		}
		if elapsed := now.Sub(sampleAt); elapsed >= 250*time.Millisecond {
			speed = float64(p.Downloaded-sampleBytes) / elapsed.Seconds()
			sampleAt = now
			sampleBytes = p.Downloaded
		}
		p.Speed = speed
		if notify != nil {
			notify(p)
		}
	}
	cfg := s.store.Config()
	if formatID <= 0 {
		formats := video.Formats
		if formatID == -1 {
			formats = video.AudioFormats
		}
		if len(formats) == 0 {
			resolved, err := s.Resolve(ctx, video.ID)
			if err != nil {
				return "", err
			}
			formats = resolved.Formats
			if formatID == -1 {
				formats = resolved.AudioFormats
			}
		}
		if len(formats) == 0 {
			return "", errors.New("没有可下载的对应格式")
		}
		formatID = formats[0].ID
	}
	key := "youtube:downloaded:" + video.ID + fmt.Sprintf(":%d", formatID)
	if cfg.Download.SkipDownloaded {
		if raw, e := s.store.History(key); e == nil {
			var old History
			if json.Unmarshal([]byte(raw), &old) == nil {
				if stat, e := os.Stat(old.Path); e == nil && !stat.IsDir() {
					emit(Progress{Phase: "已跳过", Percent: 100})
					return old.Path, nil
				}
			}
		}
	}
	emit(Progress{Phase: "正在获取下载地址"})
	client, raw, err := s.fetchVideo(ctx, video.ID)
	if err != nil {
		return "", err
	}
	var selected *extractor.Format
	audioFormats := selectableAudioFormats(raw.Formats)
	for _, f := range audioFormats {
		if audioID(f) == formatID {
			selected = f
			break
		}
	}
	if selected == nil {
		for i := range raw.Formats {
			f := &raw.Formats[i]
			if f.AudioTrack != nil && !f.AudioTrack.AudioIsDefault {
				continue
			}
			if f.ItagNo == formatID && strings.HasPrefix(f.MimeType, "video/") {
				selected = f
				break
			}
		}
	}
	if selected == nil {
		return "", errors.New("所选画质已不可用，请重新解析")
	}
	audioOnly := strings.HasPrefix(selected.MimeType, "audio/")
	var audio *extractor.Format
	if !audioOnly && selected.AudioChannels == 0 {
		if len(audioFormats) > 0 {
			audio = audioFormats[0]
		}
		if audio == nil {
			return "", errors.New("未找到配套音频")
		}
		if _, _, err := utils.CheckFFmpeg(cfg.Download.FFmpegPath); err != nil {
			return "", err
		}
	}
	dir := s.store.DownloadDirectory()
	if cfg.Download.Grouping && cfg.Download.GroupingRule != "none" {
		switch cfg.Download.GroupingRule {
		case "author":
			dir = filepath.Join(dir, utils.FileName(raw.Author))
		case "source":
			dir = filepath.Join(dir, "YouTube")
		case "author_source":
			dir = filepath.Join(dir, utils.FileName(raw.Author), "YouTube")
		case "date":
			dir = filepath.Join(dir, time.Now().Format("2006-01-02"))
		}
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	name := utils.ApplyFilenameTemplate(cfg.Download.FilenameTemplate, map[string]string{"title": raw.Title, "author": raw.Author, "id": raw.ID, "author_id": raw.ChannelID, "source": "YouTube", "date": time.Now().Format("2006-01-02"), "time": time.Now().Format("15-04-05"), "publish_date": publishDate(raw.PublishDate)})
	if name == "" || len(name) > 200 {
		name = raw.ID
	}
	ext := ".mp4"
	if strings.Contains(selected.MimeType, "webm") {
		ext = ".webm"
	}
	if audioOnly && strings.Contains(selected.MimeType, "mp4") {
		ext = ".m4a"
	}
	if audio != nil {
		ext = ".mkv"
	}
	temp, err := os.MkdirTemp(dir, ".youtube-")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(temp)
	videoPath := filepath.Join(temp, "video"+ext)
	err = s.downloadStream(ctx, client, raw, selected, videoPath, func(done, total int64) {
		p := float64(0)
		if total > 0 {
			p = float64(done) / float64(total) * 90
		}
		phase := "下载视频"
		if audioOnly {
			phase = "下载音频"
		}
		emit(Progress{Phase: phase, Downloaded: done, Total: total, Percent: p})
	})
	if err != nil {
		return "", err
	}
	ready := videoPath
	if audio != nil {
		audioPath := filepath.Join(temp, "audio")
		if err = s.downloadStream(ctx, client, raw, audio, audioPath, func(done, total int64) {
			p := float64(90)
			if total > 0 {
				p += float64(done) / float64(total) * 8
			}
			emit(Progress{Phase: "下载音频", Downloaded: done, Total: total, Percent: p})
		}); err != nil {
			return "", err
		}
		emit(Progress{Phase: "合并音视频", Percent: 98})
		ready = filepath.Join(temp, "merged.mkv")
		if err = utils.NewFFmpeg(cfg.Download.FFmpegPath).MergeContext(ctx, videoPath, audioPath, ready); err != nil {
			return "", err
		}
	}
	if err = ctx.Err(); err != nil {
		return "", err
	}
	s.publishMu.Lock()
	final := utils.UniqueFilePath(filepath.Join(dir, name+ext))
	err = os.Rename(ready, final)
	s.publishMu.Unlock()
	if err != nil {
		return "", err
	}
	kind := "video"
	if audioOnly {
		kind = "audio"
	}
	payload, _ := json.Marshal(History{Kind: kind, ID: raw.ID, Title: raw.Title, Author: raw.Author, Path: final, Downloaded: time.Now().UTC().Format(time.RFC3339)})
	if err = s.store.SaveHistory(key, string(payload)); err != nil {
		return final, fmt.Errorf("视频已保存，但历史记录写入失败：%w", err)
	}
	emit(Progress{Phase: "已完成", Percent: 100})
	return final, nil
}
func publishDate(date time.Time) string {
	if date.IsZero() {
		return ""
	}
	return date.Format("2006-01-02")
}

// fetchVideo retries a public request without account cookies when the account
// request is rejected; a successful authenticated response is always preferred.
func (s *Service) fetchVideo(ctx context.Context, id string) (*extractor.Client, *extractor.Video, error) {
	s.parseMu.Lock()
	defer s.parseMu.Unlock()
	if err := ctx.Err(); err != nil {
		return nil, nil, err
	}
	key := fmt.Sprintf("%s:%x", id, sha256.Sum256([]byte(s.Cookie())))
	if hit, ok := s.cache[key]; ok && time.Since(hit.at) < 5*time.Minute {
		return hit.client.Clone(), hit.video, nil
	}
	remember := func(client *extractor.Client, video *extractor.Video) (*extractor.Client, *extractor.Video, error) {
		if s.cache == nil {
			s.cache = map[string]cachedVideo{}
		}
		if len(s.cache) >= 8 {
			var oldest string
			var at time.Time
			for k, v := range s.cache {
				if at.IsZero() || v.at.Before(at) {
					oldest, at = k, v.at
				}
			}
			delete(s.cache, oldest)
		}
		s.cache[key] = cachedVideo{client: client.Clone(), video: video, at: time.Now()}
		return client, video, nil
	}

	client := s.client()
	video, err := client.GetVideoContext(ctx, id)
	if err == nil && usableFormats(video) {
		return remember(client, video)
	}
	if ctx.Err() != nil {
		return nil, nil, ctx.Err()
	}
	if client.Cookie != "" {
		anonymous := &extractor.Client{HTTPClient: client.HTTPClient}
		fallback, fallbackErr := anonymous.GetVideoContext(ctx, id)
		if fallbackErr == nil && usableFormats(fallback) {
			return remember(anonymous, fallback)
		}
	}
	if err != nil {
		if loggedIn, known := client.AccountState(); known && loggedIn {
			err = fmt.Errorf("%w：%w", errAuthenticatedPlayback, err)
		}
		return client, nil, err
	}
	return nil, nil, errors.New("播放器未提供可下载的直链格式，当前精简解析器尚不支持该响应")
}
func (s *Service) parseError(err error) error {
	if errors.Is(err, errAuthenticatedPlayback) {
		return fmt.Errorf("%w。Cookie 有效，无需重复填写；当前精简解析器尚未支持该播放器响应", err)
	}
	var status *extractor.ErrPlayabiltyStatus
	if errors.As(err, &status) && strings.Contains(strings.ToLower(status.Reason), "page needs to be reloaded") {
		if s.Cookie() != "" {
			return fmt.Errorf("已读取保存的 Cookie，但 YouTube 未通过当前客户端的播放校验：%s。当前精简解析器尚未适配此验证，请确认程序与浏览器使用相同网络出口", status.Reason)
		}
		return fmt.Errorf("YouTube 未通过当前客户端的播放校验：%s", status.Reason)
	}

	if errors.Is(err, extractor.ErrLoginRequired) {
		if s.Cookie() != "" {
			return fmt.Errorf("已读取保存的 Cookie，但仍要求登录或验证网络：%w。请确认 Cookie 来自当前已登录会话，且程序与浏览器使用相同的代理或网络出口；当前解析器无法自动完成验证", err)
		}
		return fmt.Errorf("YouTube 解析失败：%w，可在账号页填写已登录浏览器的 Cookie", err)
	}
	return fmt.Errorf("YouTube 解析失败：%w", err)
}

func formatSummary(f extractor.Format) string {
	parts := strings.Split(formatLabel(f), " · ")
	summary := strings.Join(parts[:min(2, len(parts))], " · ")
	if f.FPS > 30 {
		summary += fmt.Sprintf(" · %d帧", f.FPS)
	}
	return summary
}
