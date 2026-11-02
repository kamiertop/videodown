package download

// Task describes selected video streams, image collections, and related assets.
type Task struct {
	AwemeID      string
	SourceName   string
	Title        string
	Cover        string
	Duration     int
	AuthorName   string
	PublishTime  int
	DiggCount    int
	CollectCount int
	VideoURL     string
	ImageURLs    []string
	Assets       []Asset
	MusicURL     string
}

type Asset struct {
	URL  string
	Kind string
	Ext  string
}

type Result struct {
	AwemeID string
	Title   string
	Path    string
	Error   string
}

type BatchResult struct {
	Results []Result
	Success int
	Failed  int
}

// Progress describes download state published through the native listener.
type Progress struct {
	AwemeID        string
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
