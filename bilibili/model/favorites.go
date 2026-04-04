package model

// FavoritesResponse 收藏夹列表，包含了用户的所有收藏夹信息，但不包含收藏夹内的视频列表等详细信息
type FavoritesResponse struct {
	Code    int64         `json:"code"`
	Message string        `json:"message"`
	TTL     int64         `json:"ttl"`
	Data    FavoritesData `json:"data"`
}

type FavoritesData struct {
	Count  int64          `json:"count"`
	List   []FavoriteItem `json:"list"`
	Season any            `json:"season"`
}

type FavoriteItem struct {
	ID         int64  `json:"id"`
	Fid        int64  `json:"fid"`
	Mid        int64  `json:"mid"`
	Attr       int64  `json:"attr"`
	Title      string `json:"title"`
	FavState   int64  `json:"fav_state"`
	MediaCount int64  `json:"media_count"`
}

// FavoriteResponse 收藏夹详情，包含了收藏夹内的视频列表等信息
type FavoriteResponse struct {
	Code    int          `json:"code"`
	Message string       `json:"message"`
	TTL     int          `json:"ttl"`
	Data    FavoriteData `json:"data"`
}
type FavoriteData struct {
	Info    FavoriteInfo     `json:"info"`
	Medias  []FavoriteMedias `json:"medias"`
	HasMore bool             `json:"has_more"`
	TTL     int              `json:"ttl"`
}

// FavoriteInfoUpper info中的用户信息
type FavoriteInfoUpper struct {
	Mid       int    `json:"mid"`
	Name      string `json:"name"`
	Face      string `json:"face"`
	Followed  bool   `json:"followed"`
	VipType   int    `json:"vip_type"`
	VipStatue int    `json:"vip_statue"`
}
type FavoriteInfoCntInfo struct {
	Collect int `json:"collect"`  // 收藏数
	Play    int `json:"play"`     // 播放数
	ThumbUp int `json:"thumb_up"` // 点赞数
	Share   int `json:"share"`    // 分享数
}

// FavoriteInfo 收藏夹元数据
type FavoriteInfo struct {
	ID         int64               `json:"id"`    // 收藏夹mlid（完整id）	收藏夹原始id+创建者mid尾号2位
	Fid        int                 `json:"fid"`   // 收藏夹原始id
	Mid        int                 `json:"mid"`   // 创建者mid
	Attr       int                 `json:"attr"`  // 0：正常，1： 失效
	Title      string              `json:"title"` // 收藏夹名称
	Cover      string              `json:"cover"` // 封面图片
	Upper      FavoriteInfoUpper   `json:"upper"` // 收藏夹用户信息
	CoverType  int                 `json:"cover_type"`
	CntInfo    FavoriteInfoCntInfo `json:"cnt_info"` // 收藏夹状态数
	Type       int                 `json:"type"`
	Intro      string              `json:"intro"` // 备注
	Ctime      int                 `json:"ctime"` // 创建时间
	Mtime      int                 `json:"mtime"` // 收藏时间
	State      int                 `json:"state"`
	FavState   int                 `json:"fav_state"`
	LikeState  int                 `json:"like_state"`
	MediaCount int                 `json:"media_count"` // 内容数量
	IsTop      bool                `json:"is_top"`
}
type FavoriteMediaUpper struct {
	Mid      int64  `json:"mid"`
	Name     string `json:"name"`
	Face     string `json:"face"`
	JumpLink string `json:"jump_link"`
}
type FavoriteMediaCntInfo struct {
	Collect    int    `json:"collect"` // 收藏数
	Play       int    `json:"play"`    // 播放数
	Danmaku    int    `json:"danmaku"` // 弹幕数
	Vt         int    `json:"vt"`
	PlaySwitch int    `json:"play_switch"`
	Reply      int    `json:"reply"`
	ViewText1  string `json:"view_text_1"` // 播放数文本, 比如：12.3万
}
type FavoriteMediaUgc struct {
	FirstCid int `json:"first_cid"`
}

// FavoriteMedias 收藏夹内的视频信息，包含了视频的基本信息和一些状态信息
type FavoriteMedias struct {
	ID            int                  `json:"id"`       // 内容id，视频稿件：视频稿件avid ; 音频：音频auid ;视频合集：视频合集id
	Type          int                  `json:"type"`     // 内容类型， 2：视频稿件； 12：音频；21：视频合集
	Title         string               `json:"title"`    // 标题
	Cover         string               `json:"cover"`    // 封面url
	Intro         string               `json:"intro"`    // 简介
	Page          int                  `json:"page"`     // 视频分P数
	Duration      int                  `json:"duration"` // 视频/音频时长
	Upper         FavoriteMediaUpper   `json:"upper"`    // up主信息
	Attr          int                  `json:"attr"`     // 0: 正常; 9: up自己删除；1： 其他原因删除
	CntInfo       FavoriteMediaCntInfo `json:"cnt_info"` // 状态数
	Link          string               `json:"link"`     // 跳转uri，bilibili://video/xxxxx
	Ctime         int                  `json:"ctime"`    // 创建时间
	PubTime       int                  `json:"pubtime"`  // 发布时间
	FavTime       int                  `json:"fav_time"` // 收藏时间
	BvID          string               `json:"bv_id"`    // 视频稿件bvid
	Bvid          string               `json:"bvid"`     // 视频稿件bvid
	Season        any                  `json:"season"`
	Ogv           any                  `json:"ogv"`
	Ugc           FavoriteMediaUgc     `json:"ugc"`
	MediaListLink string               `json:"media_list_link"` // bilibili://music/playlist/playpage/3997344669?page_type=3\u0026oid=1305850968\u0026otype=2
}
