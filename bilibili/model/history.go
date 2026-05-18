package model

// DownloadHistoryItem 下载历史记录
type DownloadHistoryItem struct {
	Bvid         string `json:"bvid"`
	Cid          int64  `json:"cid"`
	Title        string `json:"title"`
	Cover        string `json:"cover"`
	Duration     int    `json:"duration"`
	UpperName    string `json:"upperName"`
	Play         int    `json:"play"`
	Danmaku      int    `json:"danmaku"`
	Pubtime      int    `json:"pubtime"`
	SourceName   string `json:"sourceName"`
	SourceKind   string `json:"sourceKind"`
	Path         string `json:"path"`
	DownloadKind string `json:"downloadKind"`
	// Wails 绑定生成不支持直接暴露 time.Time，保存为 RFC3339 字符串给前端解析。
	Downloaded string `json:"downloaded"`
}
