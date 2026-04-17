package model

type SeasonsSeriesData struct {
	ItemsLists SeasonsSeriesItemsLists `json:"items_lists"`
}

type SeasonsSeriesItemsLists struct {
	Page        SeasonsSeriesPage `json:"page"`
	SeasonsList []SeasonsItem     `json:"seasons_list"`
	SeriesList  []SeriesItem      `json:"series_list"`
}

type SeasonsSeriesPage struct {
	PageNum  int `json:"page_num"`
	PageSize int `json:"page_size"`
	// Total 合集+系列的总数
	Total int `json:"total"` // 合集+系列
}

type SeasonsItem struct {
	Archives   []SeasonsSeriesArchivesItem `json:"archives"`
	Meta       SeasonsItemMeta             `json:"meta"`
	RecentAids []int64                     `json:"recent_aids"`
}

type SeasonsItemMeta struct {
	Category    int    `json:"category"`
	Cover       string `json:"cover"`       // 合集封面
	Description string `json:"description"` // 合集描述
	Mid         int    `json:"mid"`
	Name        string `json:"name"`
	Ptime       int    `json:"ptime"`
	SeasonId    int    `json:"season_id"`
	Total       int    `json:"total"` // 合集总量
	Title       string `json:"title"`
}

type SeasonsSeriesArchivesItem struct {
	Aid              int                       `json:"aid"`
	Bvid             string                    `json:"bvid"`
	Ctime            int                       `json:"ctime"`
	Duration         int                       `json:"duration"`
	EnableVt         bool                      `json:"enable_vt"`
	InteractiveVideo bool                      `json:"interactive_video"`
	Pic              string                    `json:"pic"`
	PlaybackPosition int                       `json:"playback_position"`
	Pubdate          int                       `json:"pubdate"`
	Stat             SeasonsSeriesArchivesStat `json:"stat"`
	State            int                       `json:"state"`
	Title            string                    `json:"title"`
	UgcPay           int                       `json:"ugc_pay"`
	VtDisplay        string                    `json:"vt_display"`
	IsLessonVideo    int                       `json:"is_lesson_video"`
}

type SeasonsSeriesArchivesStat struct {
	View    int `json:"view"`
	Vt      int `json:"vt"`
	Danmaku int `json:"danmaku"`
}

type SeriesItem struct {
	Archives   []SeasonsSeriesArchivesItem `json:"archives"`
	Meta       SeriesItemMeta              `json:"meta"`
	RecentAids []int                       `json:"recent_aids"`
}

type SeriesItemMeta struct {
	Category     int      `json:"category"`
	Cover        string   `json:"cover"`
	Creator      string   `json:"creator"`
	Ctime        int      `json:"ctime"`
	Description  string   `json:"description"`
	Keywords     []string `json:"keywords"`
	LastUpdateTs int      `json:"last_update_ts"`
	Mid          int      `json:"mid"`
	Mtime        int      `json:"mtime"`
	Name         string   `json:"name"`
	RawKeywords  string   `json:"raw_keywords"`
	SeriesId     int      `json:"series_id"`
	State        int      `json:"state"`
	Total        int      `json:"total"`
}

// SeasonsArchivesData 一个合集内的视频信息
type SeasonsArchivesData struct {
	Aids     []uint64              `json:"aids"`
	Page     SeasonsArchivesPage   `json:"page"`
	Meta     SeasonsArchivesMeta   `json:"meta"`
	Archives []SeasonsArchivesItem `json:"archives"`
}

type SeasonsArchivesItem struct {
	Aid              int                       `json:"aid"`
	Bvid             string                    `json:"bvid"`
	Ctime            int                       `json:"ctime"`
	Duration         int                       `json:"duration"`
	EnableVt         bool                      `json:"enable_vt"`
	InteractiveVideo bool                      `json:"interactive_video"`
	Pic              string                    `json:"pic"`
	PlaybackPosition int                       `json:"playback_position"`
	Pubdate          int                       `json:"pubdate"`
	State            int                       `json:"state"`
	Stat             SeasonsSeriesArchivesStat `json:"stat"`
	Title            string                    `json:"title"`
	UgcPay           int                       `json:"ugc_pay"`
	VtDisplay        string                    `json:"vt_display"`
	IsLessonVideo    int                       `json:"is_lesson_video"`
}

type SeasonsArchivesMeta struct {
	Cover       string `json:"cover"`
	Description string `json:"description"`
	Title       string `json:"title"`
	Category    int    `json:"category"`
	Mid         int    `json:"mid"`
	Ptime       int    `json:"ptime"`
	SeasonId    int    `json:"season_id"`
	Total       int    `json:"total"` // 总数，和Page中的Total一致
}
type SeasonsArchivesPage struct {
	PageNum  int
	PageSize int
	Total    int
}

type SeriesArchivesData struct {
	Aids     []int                `json:"aids"`
	Page     SeriesListPage       `json:"page"`
	Archives []SeriesArchivesItem `json:"archives"`
}

type SeriesArchivesItem struct {
	Aid              int                       `json:"aid"`
	Title            string                    `json:"title"`
	PubDate          int                       `json:"pubdate"`
	Ctime            int                       `json:"ctime"`
	State            int                       `json:"state"`
	Pic              string                    `json:"pic"`
	Duration         int                       `json:"duration"`
	Bvid             string                    `json:"bvid"`
	UgcPay           int                       `json:"ugc_pay"`
	InteractiveVideo bool                      `json:"interactive_video"`
	EnableVt         int                       `json:"enable_vt"`
	VtDisplay        string                    `json:"vt_display"`
	Desc             string                    `json:"desc"`
	UpMid            int                       `json:"up_mid"`
	Stat             SeasonsSeriesArchivesStat `json:"stat"`
}

type SeriesArchivesStat struct {
	Vies int `json:"view"`
}

type SeriesListPage struct {
	Num   int `json:"num"`   // 当前页
	Size  int `json:"size"`  // 每页数量
	Total int `json:"total"` // 总数
}
