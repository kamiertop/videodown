package model

type SeasonsSeriesResponse struct {
	Code    int               `json:"code"`
	Message string            `json:"message"`
	Ttl     int               `json:"ttl"`
	Data    SeasonsSeriesData `json:"data"`
}

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
	Cover       string `json:"cover"`
	Description string `json:"description"`
	Mid         int    `json:"mid"`
	Name        string `json:"name"`
	Ptime       int    `json:"ptime"`
	SeasonId    int    `json:"season_id"`
	Total       int    `json:"total"`
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
