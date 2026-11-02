package youtube

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/kamiertop/videodown/youtube/internal/extractor"
)

var mediaTransport = func() *http.Transport {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.MaxIdleConnsPerHost = 8
	transport.MaxConnsPerHost = 8
	return transport
}()

const streamChunkSize int64 = 4 << 20

const streamWorkers = 4

func (s *Service) downloadStream(ctx context.Context, client *extractor.Client, video *extractor.Video, format *extractor.Format, path string, progress func(int64, int64)) error {
	rawURL, err := client.GetStreamURLContext(ctx, video, format)
	if err != nil {
		return err
	}
	u, err := url.Parse(rawURL)
	if err != nil || u.Scheme != "https" || !(u.Hostname() == "googlevideo.com" || strings.HasSuffix(u.Hostname(), ".googlevideo.com")) {
		return errors.New("无效的视频下载地址")
	}
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	transport := s.transport
	if transport == nil {
		transport = mediaTransport
	}
	// Long transfers have no overall deadline; the user's cancellation applies to every worker.
	httpClient := &http.Client{Timeout: 0, Transport: transport}
	if format.ContentLength > 0 {
		if err := downloadRanges(ctx, httpClient, u, client.UserAgent(), file, format.ContentLength, progress); err != nil {
			return err
		}
	} else {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
		if err != nil {
			return err
		}
		req.Header.Set("User-Agent", client.UserAgent())
		resp, err := httpClient.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			return fmt.Errorf("视频下载失败：HTTP %d", resp.StatusCode)
		}
		var done int64
		writer := &progressWriter{Writer: file, onWrite: func(n int) { done += int64(n); progress(done, resp.ContentLength) }}
		if _, err := io.Copy(writer, resp.Body); err != nil {
			return err
		}
		if done == 0 || (resp.ContentLength > 0 && done != resp.ContentLength) {
			return io.ErrUnexpectedEOF
		}
	}
	return file.Close()
}

func downloadRanges(parent context.Context, client *http.Client, base *url.URL, userAgent string, file *os.File, total int64, progress func(int64, int64)) error {
	ctx, cancel := context.WithCancel(parent)
	defer cancel()
	count := int((total + streamChunkSize - 1) / streamChunkSize)
	workers := min(streamWorkers, count)
	if total < 8<<20 {
		workers = 1
	}
	type part struct {
		index      int
		start, end int64
	}
	jobs := make(chan part, count)
	for i := range count {
		start := int64(i) * streamChunkSize
		jobs <- part{i, start, min(start+streamChunkSize, total) - 1}
	}
	close(jobs)
	var mu sync.Mutex
	var done int64
	partial := make([]int64, count)
	var firstErr error
	report := func(index int, offset int64) {
		mu.Lock()
		defer mu.Unlock()
		if offset > partial[index] {
			done += offset - partial[index]
			partial[index] = offset
			progress(done, total)
		}
	}
	var wait sync.WaitGroup
	for range workers {
		wait.Go(func() {
			for job := range jobs {
				if ctx.Err() != nil {
					return
				}
				err := downloadRange(ctx, client, base, userAgent, file, job.start, job.end, func(n int64) { report(job.index, n) })
				if err != nil {
					mu.Lock()
					if firstErr == nil {
						firstErr = err
					}
					mu.Unlock()
					cancel()
					return
				}
			}
		})
	}
	wait.Wait()
	if parent.Err() != nil {
		return parent.Err()
	}
	if firstErr != nil {
		return firstErr
	}
	if done != total {
		return io.ErrUnexpectedEOF
	}
	return nil
}

func downloadRange(ctx context.Context, client *http.Client, base *url.URL, userAgent string, file *os.File, start, end int64, progress func(int64)) error {
	const attempts = 3
	var last error
	for attempt := range attempts {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if attempt > 0 {
			timer := time.NewTimer(time.Duration(attempt) * 300 * time.Millisecond)
			select {
			case <-timer.C:
			case <-ctx.Done():
				timer.Stop()
				return ctx.Err()
			}
		}
		u := *base
		q := u.Query()
		q.Set("range", fmt.Sprintf("%d-%d", start, end))
		u.RawQuery = q.Encode()
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
		if err != nil {
			return err
		}
		req.Header.Set("User-Agent", userAgent)
		resp, err := client.Do(req)
		if err != nil {
			last = err
			continue
		}
		if resp.StatusCode != 200 && resp.StatusCode != 206 {
			resp.Body.Close()
			last = fmt.Errorf("视频下载失败：HTTP %d", resp.StatusCode)
			if resp.StatusCode < 500 {
				return last
			}
			continue
		}
		expected := end - start + 1
		if resp.ContentLength >= 0 && resp.ContentLength != expected {
			resp.Body.Close()
			return fmt.Errorf("视频分片长度不匹配：期望 %d，实际 %d", expected, resp.ContentLength)
		}
		var written int64
		writer := &progressWriter{Writer: io.NewOffsetWriter(file, start), onWrite: func(n int) { written += int64(n); progress(written) }}
		_, err = io.CopyN(writer, resp.Body, expected)
		if err == nil {
			var extra [1]byte
			n, readErr := resp.Body.Read(extra[:])
			if n != 0 {
				err = errors.New("服务器未正确处理分片范围")
			} else if readErr != nil && !errors.Is(readErr, io.EOF) {
				err = readErr
			}
		}
		resp.Body.Close()
		if err == nil {
			return nil
		}
		last = err
	}
	return last
}

type progressWriter struct {
	io.Writer
	onWrite func(int)
}

func (w *progressWriter) Write(p []byte) (int, error) {
	n, err := w.Writer.Write(p)
	if n > 0 {
		w.onWrite(n)
	}
	return n, err
}
