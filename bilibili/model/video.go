package model

// VideoListData UP主视频列表数据
type VideoListData struct {
	List           VideoList      `json:"list"`
	Page           VideoListPage  `json:"page"`
	EpisodicButton EpisodicButton `json:"episodic_button"`
	IsRisk         bool           `json:"is_risk"`
	GaiaResType    int            `json:"gaia_res_type"`
	GaiaData       any            `json:"gaia_data"`
}
type VideoList struct {
	Slist []any       `json:"slist"`
	Tlist any         `json:"tlist"`
	Vlist []VListItem `json:"vlist"`
}

type VListItem struct {
	Comment          int                `json:"comment"` // 视频评论数
	Typeid           int                `json:"typeid"`
	Play             int                `json:"play"` // 视频播放数
	Pic              string             `json:"pic"`  // 视频封面url
	Subtitle         string             `json:"subtitle"`
	Description      string             `json:"description"`
	Copyright        string             `json:"copyright"`
	Title            string             `json:"title"` // 视频标题
	Review           int                `json:"review"`
	Author           string             `json:"author"`
	Mid              int64              `json:"mid"`     // 视频作者的mid
	Created          int                `json:"created"` // 视频创建时间
	Length           string             `json:"length"`  // 视频时长，格式为"mm:ss"
	VideoReview      int                `json:"video_review"`
	Aid              int64              `json:"aid"`  // 视频av号
	Bvid             string             `json:"bvid"` // 视频bv号
	HideClick        bool               `json:"hide_click"`
	IsPay            int                `json:"is_pay"`
	IsUnionVideo     int                `json:"is_union_video"` // 是否是联合视频，0：否，1：是
	IsSteinsGate     int                `json:"is_steins_gate"`
	IsLivePlayback   int                `json:"is_live_playback"`
	IsLessonVideo    int                `json:"is_lesson_video"`
	IsLessonFinished int                `json:"is_lesson_finished"`
	LessonUpdateInfo string             `json:"lesson_update_info"`
	JumpUrl          string             `json:"jump_url"`
	Meta             VideoVListItemMeta `json:"meta"`
	IsAvoided        int                `json:"is_avoided"`
	SeasonId         int                `json:"season_id"`
	Attribute        int                `json:"attribute"`
	IsChargingArc    bool               `json:"is_charging_arc"`
	ElecArcType      int                `json:"elec_arc_type"`
	ElecArcBadge     string             `json:"elec_arc_badge"`
	Vt               int                `json:"vt"`
	EnableVt         int                `json:"enable_vt"`
	VtDisplay        string             `json:"vt_display"`
	PlaybackPosition int                `json:"playback_position"`
	IsSelfView       bool               `json:"is_self_view"`
	ViewSelfType     int                `json:"view_self_type"`
}

type VideoVListItemMeta struct {
	Id        int               `json:"id"`
	Title     string            `json:"title"`
	Cover     string            `json:"cover"`
	Mid       int64             `json:"mid"`
	Intro     string            `json:"intro"`
	SignState int               `json:"sign_state"`
	Attribute int               `json:"attribute"`
	Stat      VListItemMetaStat `json:"stat"`
	EpCount   int               `json:"ep_count"`
	FirstAid  int64             `json:"first_aid"`
	Ptime     int               `json:"ptime"`
	EpNum     int               `json:"ep_num"`
	Show      int               `json:"show"`
}

type VListItemMetaStat struct {
	SeasonId int `json:"season_id"`
	View     int `json:"view"`    // 播放数
	Danmaku  int `json:"danmaku"` // 弹幕数
	Reply    int `json:"reply"`
	Favorite int `json:"favorite"` // 收藏数
	Coin     int `json:"coin"`     // 硬币数
	Share    int `json:"share"`
	Like     int `json:"like"` // 点赞数
	Mtime    int `json:"mtime"`
	Vt       int `json:"vt"`
	Vv       int `json:"vv"`
}

type VideoListPage struct {
	Pn    int `json:"pn"`
	Ps    int `json:"ps"`
	Count int `json:"count"`
}

type EpisodicButton struct {
	Text string `json:"text"`
	Uri  string `json:"uri"`
}
