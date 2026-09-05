package download

// Task 是前端提交给后端的最小任务协议。
// 前端已经完成清晰度选择，所以后端只消费最终 videoURL；图片合集则使用 ImageURLs。
type Task struct {
	AwemeID      string   `json:"awemeId"`
	SourceKind   string   `json:"sourceKind"`
	SourceName   string   `json:"sourceName"`
	Title        string   `json:"title"`
	Cover        string   `json:"cover"`
	Duration     int      `json:"duration"`
	AuthorName   string   `json:"authorName"`
	PublishTime  int      `json:"publishTime"`
	DiggCount    int      `json:"diggCount"`
	CollectCount int      `json:"collectCount"`
	VideoURL     string   `json:"videoURL"`
	ImageURLs    []string `json:"imageURLs"`
	Assets       []Asset  `json:"assets"`
	MusicURL     string   `json:"musicURL"`
}

type Asset struct {
	URL  string `json:"url"`
	Kind string `json:"kind"`
	Ext  string `json:"ext"`
}

type Result struct {
	AwemeID string `json:"awemeId"`
	Title   string `json:"title"`
	Path    string `json:"path"`
	Error   string `json:"error"`
}

type BatchResult struct {
	Results []Result `json:"results"`
	Success int      `json:"success"`
	Failed  int      `json:"failed"`
}

// progress 通过 Wails 事件推给前端，前端按 awemeId 更新对应卡片进度。
type progress struct {
	AwemeID        string  `json:"awemeId"`
	Title          string  `json:"title"`
	Phase          string  `json:"phase"`
	Downloaded     int64   `json:"downloaded"`
	Total          int64   `json:"total"`
	Percent        float64 `json:"percent"`
	SleepRemaining int64   `json:"sleepRemaining"`
	SleepTotal     int64   `json:"sleepTotal"`
}
