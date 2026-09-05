package download

// Task 是前端提交给后端的最小下载任务；流地址已经由前端按用户选择的画质/音质确定。
type Task struct {
	SourceName string `json:"sourceName"`
	SourceKind string `json:"sourceKind"`
	UpperName  string `json:"upperName"`
	Bvid       string `json:"bvid"`
	// Cid 为分 P 稿件某一 P 的 cid；单 P 或未填时为 0，行为与旧版一致。
	Cid      int64  `json:"cid"`
	Title    string `json:"title"`
	Cover    string `json:"cover"`
	Duration int    `json:"duration"`
	Play     int    `json:"play"`
	Danmaku  int    `json:"danmaku"`
	Pubtime  int    `json:"pubtime"`
	VideoURL string `json:"videoURL"`
	AudioURL string `json:"audioURL"`
}

// Result 记录单个任务的结果，前端据此移除已完成的视频并保留失败项。
type Result struct {
	Bvid  string `json:"bvid"`
	Cid   int64  `json:"cid"`
	Title string `json:"title"`
	Path  string `json:"path"`
	Error string `json:"error"`
}

// BatchResult 汇总批量下载结果；失败不会中断整批任务。
type BatchResult struct {
	Results []Result `json:"results"`
	Success int      `json:"success"`
	Failed  int      `json:"failed"`
}

type progress struct {
	Bvid           string  `json:"bvid"`
	Cid            int64   `json:"cid"`
	Title          string  `json:"title"`
	Phase          string  `json:"phase"`
	Downloaded     int64   `json:"downloaded"`
	Total          int64   `json:"total"`
	Percent        float64 `json:"percent"`
	SleepRemaining int64   `json:"sleepRemaining"`
	SleepTotal     int64   `json:"sleepTotal"`
}
