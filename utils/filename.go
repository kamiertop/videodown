package utils

import (
	"errors"
	"fmt"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var templateToken = regexp.MustCompile(`\{(title|author|source|folder|collection|publish_date|date|time|id|author_id)\}`)
var unsupportedToken = regexp.MustCompile(`\{[^{}]*\}`)
var emptyFieldSeparator = regexp.MustCompile(`\s*[-_|]\s*\x00(?:\s*[-_|]\s*\x00)*|\x00\s*[-_|]\s*`)

// ApplyFilenameTemplate replaces supported tokens and sanitizes the result.
func ApplyFilenameTemplate(template string, values map[string]string) string {
	name := templateToken.ReplaceAllStringFunc(template, func(token string) string {
		value := values[token[1:len(token)-1]]
		if strings.TrimSpace(value) == "" {
			return "\x00"
		}
		return value
	})
	// Remove unknown placeholders instead of writing them into filenames.
	name = unsupportedToken.ReplaceAllString(name, "")
	// Empty optional fields must not leave a dangling separator.
	name = emptyFieldSeparator.ReplaceAllString(name, " ")
	name = strings.ReplaceAll(name, "\x00", "")
	name = FileNamePreserveSpaces(name)
	// FileName normalizes whitespace to underscores; restore the readable
	// separators selected by the template (for example " - ").
	name = regexp.MustCompile(`_+[-]_+`).ReplaceAllString(name, " - ")
	name = regexp.MustCompile(`_+\|_+`).ReplaceAllString(name, " | ")
	return strings.Trim(name, "-| .")
}

// SupportedFilenameTemplate reports whether a template contains at least one supported field.
func SupportedFilenameTemplate(template string) bool { return templateToken.MatchString(template) }

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
	return fileName(rawName, false)
}

func FileNamePreserveSpaces(rawName string) string {
	return fileName(rawName, true)
}

func fileName(rawName string, preserveSpaces bool) string {
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

	// 保留标题中的普通空格（仅折叠连续空白），避免中文标题被改成下划线风格。
	separator := "_"
	if preserveSpaces {
		separator = " "
	}
	s = strings.Join(strings.Fields(b.String()), separator)
	s = strings.Trim(s, " .")
	if isReservedWindowsName(s) {
		s = "_" + s
	}
	return s
}

func isReservedWindowsName(name string) bool {
	base := strings.ToUpper(strings.TrimSpace(strings.SplitN(name, ".", 2)[0]))
	switch base {
	case "CON", "PRN", "AUX", "NUL", "COM1", "COM2", "COM3", "COM4", "COM5", "COM6", "COM7", "COM8", "COM9", "LPT1", "LPT2", "LPT3", "LPT4", "LPT5", "LPT6", "LPT7", "LPT8", "LPT9":
		return true
	}
	return false
}

// MaxFileNameBytes 文件/目录名单部分的安全字节上限：Linux 单个名字上限为
// 255 字节（NAME_MAX），再预留扩展名（如 ".mp4"）和 UniqueFilePath 的
// "(N)" 后缀空间。超过上限的名字必须由调用方退回短视频 ID 之类的短名字，
// 否则 os.Stat/os.Create 会返回 ENAMETOOLONG。
const MaxFileNameBytes = 200

// UniqueFilePath 在给定路径已存在时，生成一个唯一的文件路径。
// 注意：stat 返回“不存在”以外的错误（例如文件名过长的 ENAMETOOLONG）时
// 必须直接返回，交给后续 os.Create 暴露真实原因，否则会在这里无限循环。
func UniqueFilePath(path string) string {
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		return path
	}
	ext := filepath.Ext(path)
	base := strings.TrimSuffix(path, ext)
	for i := 1; ; i++ {
		candidate := fmt.Sprintf("%s(%d)%s", base, i, ext)
		_, err := os.Stat(candidate)
		switch {
		case err == nil:
			continue // 已存在，继续尝试下一个序号
		case errors.Is(err, os.ErrNotExist):
			return candidate
		default:
			// 无法确认存在性（权限、文件名过长等）：原样返回，让 Create 报错。
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
