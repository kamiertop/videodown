package utils

import "strings"

// FileName 修改视频文件名为合法的文件名
func FileName(rawName string) string {
	// 移除首尾空格
	rawName = strings.TrimSpace(rawName)
	strings.NewReplacer()
	// 替换 '/'
	rawName = strings.ReplaceAll(rawName, "/", " ")

	rawName = strings.ReplaceAll(rawName, "\\", " ")

	rawName = strings.Join(strings.Fields(rawName), "_")

	return rawName
}
