package utils

import (
	"errors"
	"fmt"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// isIllegalChar 判断是否为文件名中的非法字符。
func isIllegalChar(r rune) bool {
	if r <= 0x1f {
		return true
	}
	switch r {
	case '<', '>', ':', '"', '/', '\\', '|', '?', '*':
		return true
	}
	return false
}

// FileName 清理文件名中的非法字符，返回合法的文件/目录名。
// 空字符串输入或清理后为空时返回空字符串，由调用方决定默认值。
func FileName(rawName string) string {
	s := strings.TrimSpace(rawName)
	if s == "" {
		return ""
	}

	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if isIllegalChar(r) {
			b.WriteByte(' ')
		} else {
			b.WriteRune(r)
		}
	}

	s = strings.Join(strings.Fields(b.String()), "_")
	s = strings.Trim(s, " .")
	return s
}

// UniqueFilePath 在给定路径已存在时，生成一个唯一的文件路径
func UniqueFilePath(path string) string {
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

// ImageExtFromResponse 从URL和HTTP响应中推断图片文件扩展名，优先使用URL路径中的扩展名，如果无效则使用Content-Type头部信息
func ImageExtFromResponse(rawURL string, resp *http.Response) string {
	if u, err := url.Parse(rawURL); err == nil {
		if ext := strings.ToLower(filepath.Ext(u.Path)); ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".webp" || ext == ".gif" {
			if ext == ".jpeg" {
				return ".jpg"
			}
			return ext
		}
	}

	if resp != nil {
		contentType := resp.Header.Get("Content-Type")
		if contentType != "" {
			if exts, err := mime.ExtensionsByType(strings.Split(contentType, ";")[0]); err == nil && len(exts) > 0 {
				ext := strings.ToLower(exts[0])
				if ext == ".jpeg" || ext == ".jpe" {
					return ".jpg"
				}
				return ext
			}
		}
	}

	return ".jpg"
}
