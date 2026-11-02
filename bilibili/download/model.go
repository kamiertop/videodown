package download

// Task describes a download with the selected video and audio streams.
type Task struct {
	SourceName string
	UpperName  string
	Bvid       string
	// Cid 为分 P 稿件某一 P 的 cid；单 P 或未填时为 0，行为与旧版一致。
	Cid      int64
	Title    string
	Cover    string
	Duration int
	Play     int
	Danmaku  int
	Pubtime  int
	VideoURL string
	AudioURL string
}

// Result records the outcome of one download task.
type Result struct {
	Bvid  string
	Cid   int64
	Title string
	Path  string
	Error string
}

// DownloadKey 标识一个待判断是否已下载的视频：多 P 收藏条目带具体 cid，普通列表项常只有 bvid。
type DownloadKey struct {
	Bvid string
	Cid  int64
}

// BatchResult 汇总批量下载结果；失败不会中断整批任务。
type BatchResult struct {
	Results []Result
	Success int
	Failed  int
}

type Progress struct {
	Bvid           string
	Cid            int64
	Title          string
	Phase          string
	Downloaded     int64
	Total          int64
	Percent        float64
	SleepRemaining int64
	SleepTotal     int64
}

func (p Progress) withPhase(phase string) Progress {
	p.Phase = phase
	return p
}

func (p Progress) withPercent(percent float64) Progress {
	p.Percent = percent
	return p
}
