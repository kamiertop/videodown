package utils

import "strings"

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
