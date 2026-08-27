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

type PlayHistoryCursor struct {
	Max      int    `json:"max"`
	ViewAt   int    `json:"view_at"`
	Business string `json:"business"`
	Ps       int    `json:"ps"`
}

type PlayHistoryTab struct {
	// archive: 视频
	// live： 直播
	// article： 专栏
	Type string `json:"type"`
	Name string `json:"name"`
}

// PlayHistoryData 播放历史记录
type PlayHistoryData struct {
	Cursor PlayHistoryCursor `json:"cursor"`
	Tab    []PlayHistoryTab  `json:"tab"`
	List   []PlayHistoryItem `json:"list"`
}

type PlayHistory struct {
	Oid      int64  `json:"oid"`
	Epid     int    `json:"epid"`
	Bvid     string `json:"bvid"`
	Page     int    `json:"page"`
	Cid      int64  `json:"cid"`
	Part     string `json:"part"`
	Business string `json:"business"`
	Dt       int    `json:"dt"`
}

type PlayHistoryItem struct {
	Title      string      `json:"title"` // 标题
	LongTitle  string      `json:"long_title"`
	Cover      string      `json:"cover"` // 封面URL
	History    PlayHistory `json:"history"`
	Videos     int         `json:"videos"`
	AuthorName string      `json:"author_name"` // 作者名称
	AuthorFace string      `json:"author_face"`
	AuthorMid  int64       `json:"author_mid"`
	ViewAt     int         `json:"view_at"`
	Progress   int         `json:"progress"`
	Badge      string      `json:"badge"`
	ShowTitle  string      `json:"show_title"`
	Current    string      `json:"current"`
	Duration   int         `json:"duration"`
	Total      int         `json:"total"`
	IsFinish   int         `json:"is_finish"`
	IsFav      int         `json:"is_fav"`
	Kid        int64       `json:"kid"`
	NewDesc    string      `json:"new_desc"`
	TagName    string      `json:"tag_name"`    // 视频标签名称，例如：数码
	LiveStatus int         `json:"live_status"` // 直播状态
}
