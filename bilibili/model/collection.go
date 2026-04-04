package model

type CollectionResponse struct {
	Code    int            `json:"code"`
	Message string         `json:"message"`
	TTL     int            `json:"ttl"`
	Data    CollectionData `json:"data"`
}

// CollectionData 合集列表元信息
type CollectionData struct {
	Count   int                  `json:"count"`
	List    []CollectionDataList `json:"list"`
	HasMore bool                 `json:"has_more"`
}
type CollectionUpper struct {
	Mid      int    `json:"mid"`
	Name     string `json:"name"`
	Face     string `json:"face"`
	JumpLink string `json:"jump_link"`
}
type CollectionDataList struct {
	ID         int             `json:"id"`
	Fid        int             `json:"fid"`
	Mid        int             `json:"mid"`
	Attr       int             `json:"attr"`
	AttrDesc   string          `json:"attr_desc"`
	Title      string          `json:"title"`
	Cover      string          `json:"cover"`
	Upper      CollectionUpper `json:"upper"`
	CoverType  int             `json:"cover_type"`
	Intro      string          `json:"intro"`
	Ctime      int             `json:"ctime"`
	Mtime      int             `json:"mtime"`
	State      int             `json:"state"`
	FavState   int             `json:"fav_state"`
	MediaCount int             `json:"media_count"`
	ViewCount  int             `json:"view_count"`
	Vt         int             `json:"vt"`
	IsTop      bool            `json:"is_top"`
	RecentFav  any             `json:"recent_fav"`
	PlaySwitch int             `json:"play_switch"`
	Type       int             `json:"type"`
	Link       string          `json:"link"`
	Bvid       string          `json:"bvid"`
}

type CollectionItemResponse struct {
	Code    int                `json:"code"`
	Data    CollectionItemData `json:"data"`
	TTL     int                `json:"ttl"`
	Message string             `json:"message"`
}

// CollectionItemData 合集元信息
type CollectionItemData struct {
	Info   CollectionItemInfo     `json:"info"`
	Medias []CollectionItemMedias `json:"medias"`
}
type CollectionItemInfo struct {
	ID         int                   `json:"id"`
	SeasonType int                   `json:"season_type"`
	Title      string                `json:"title"`
	Cover      string                `json:"cover"`
	Upper      CollectionItemUpper   `json:"upper"`
	CntInfo    CollectionItemCntInfo `json:"cnt_info"`
	MediaCount int                   `json:"media_count"`
	Intro      string                `json:"intro"`
	EnableVt   int                   `json:"enable_vt"`
}
type CollectionItemUpper struct {
	Mid  int    `json:"mid"`
	Name string `json:"name"`
}
type CollectionItemCntInfo struct {
	Collect int `json:"collect"`
	Play    int `json:"play"`
	Danmaku int `json:"danmaku"`
	Vt      int `json:"vt"`
}
type CollectionItemMedias struct {
	ID         int64                      `json:"id"`
	Title      string                     `json:"title"`
	Cover      string                     `json:"cover"`
	Duration   int                        `json:"duration"`
	Pubtime    int                        `json:"pubtime"`
	Bvid       string                     `json:"bvid"`
	Upper      CollectionItemMediaUpper   `json:"upper"`
	CntInfo    CollectionItemMediaCntInfo `json:"cnt_info"`
	EnableVt   int                        `json:"enable_vt"`
	VtDisplay  string                     `json:"vt_display"`
	IsSelfView bool                       `json:"is_self_view"`
}

type CollectionItemMediaUpper struct {
	Mid  int    `json:"mid"`
	Name string `json:"name"`
}
type CollectionItemMediaCntInfo struct {
	Collect int `json:"collect"`
	Play    int `json:"play"`
	Danmaku int `json:"danmaku"`
	Vt      int `json:"vt"`
}
