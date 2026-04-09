package model

// VideoDetailConciseData 只保留必要的基本信息
type VideoDetailConciseData struct {
	View VideoDetailView `json:"view"`
}

// VideoDetailFullData 视频详情
type VideoDetailFullData struct {
	View                   VideoDetailView    `json:"View"`
	Card                   VideoCard          `json:"Card"`
	Tags                   []VideoTag         `json:"Tags"`
	Reply                  VideoReply         `json:"Reply"`
	Related                []VideoRelatedItem `json:"Related"` // 相关视频
	Spec                   any                `json:"Spec"`
	HotShare               VideoHotShare      `json:"hot_share"`
	Elec                   any                `json:"elec"`
	Emergency              VideoEmergency     `json:"emergency"`
	ViewAddit              VideoViewAddit     `json:"view_addit"`
	Guide                  any                `json:"guide"`
	QueryTags              any                `json:"query_tags"`
	Participle             []string           `json:"participle"`
	ModuleCtrl             any                `json:"module_ctrl"`
	ReplaceRecommend       bool               `json:"replace_recommend"`
	IsHitLabourDayActivity bool               `json:"is_hit_labour_day_activity"`
}

// VideoCard UP主信息卡片
type VideoCard struct {
	Card         VideoCardEntry `json:"card"`
	Space        VideoCardSpace `json:"space"`
	Following    bool           `json:"following"`
	ArchiveCount int            `json:"archive_count"`
	ArticleCount int            `json:"article_count"`
	Follower     int            `json:"follower"`
	LikeNum      int            `json:"like_num"`
}

type VideoDetailView struct {
	Pages                   []VideoPage          `json:"pages"` // 视频分P信息
	Cid                     int64                `json:"cid"`   // 视频1P cid
	Bvid                    string               `json:"bvid"`
	Aid                     int64                `json:"aid"`
	Duration                int                  `json:"duration"`   // 总时长（所有分P），单位为秒
	Videos                  int                  `json:"videos"`     // 稿件分P总数
	UgcSeason               UgcSeason            `json:"ugc_season"` // 视频合集信息，不在合集中的视频无此项
	Pic                     string               `json:"pic"`        // 视频封面url
	Title                   string               `json:"title"`      // 视频标题
	Pubdate                 int                  `json:"pubdate"`    // 发布时间
	Ctime                   int                  `json:"ctime"`      // 用户投稿时间
	Tid                     int                  `json:"tid"`        // 分区id
	TidV2                   int                  `json:"tid_v2"`     // 分区tid（v2）
	Tname                   string               `json:"tname"`      // 子分区名称
	TnameV2                 string               `json:"tname_v2"`   // 子分区名称（v2）
	Copyright               int                  `json:"copyright"`  // 视频类型。1： 原创，2： 转载，3：未填写
	Desc                    string               `json:"desc"`       // 视频简介
	DescV2                  []VideoDescV2        `json:"desc_v2"`    // 新版视频简介
	State                   int                  `json:"state"`      // 视频状态
	MissionId               int                  `json:"mission_id"` // 稿件参与的活动id
	Rights                  VideoViewRights      `json:"rights"`     // 视频属性信息
	Owner                   VideoOwner           `json:"owner"`      // UP主信息
	Stat                    VideoViewStat        `json:"stat"`       // 视频状态
	ArgueInfo               VideoDetailArgueInfo `json:"argue_info"` // 争议/警告信息
	Dynamic                 string               `json:"dynamic"`    // 视频同步发布的动态文字内容
	Dimension               Dimension            `json:"dimension"`  // 视频1P 分辨率
	SeasonId                int                  `json:"season_id"`  //
	Premiere                any                  `json:"premiere"`
	TeenageMode             int                  `json:"teenage_mode"`
	IsChargeableSeason      bool                 `json:"is_chargeable_season"`
	IsStory                 bool                 `json:"is_story"`            // 是否为动态视频
	IsUpowerExclusive       bool                 `json:"is_upower_exclusive"` // 是否为充电专属视频
	IsUpowerPlay            bool                 `json:"is_upower_play"`      // 充电专属视频是否支持试看
	IsUpowerPreview         bool                 `json:"is_upower_preview"`
	EnableVt                int                  `json:"enable_vt"`
	VtDisplay               string               `json:"vt_display"`
	IsUpowerExclusiveWithQa bool                 `json:"is_upower_exclusive_with_qa"`
	NoCache                 bool                 `json:"no_cache"`
	Subtitle                VideoSubTitle        `json:"subtitle"`
	IsSeasonDisplay         bool                 `json:"is_season_display"`
	UserGarb                UserGarb             `json:"user_garb"` // 用户装扮信息
	HonorReply              any                  `json:"honor_reply"`
	LikeIcon                string               `json:"like_icon"`
	NeedJumpBv              bool                 `json:"need_jump_bv"`
	DisableShowUpInfo       bool                 `json:"disable_show_up_info"`
	IsStoryPlay             int                  `json:"is_story_play"`
	IsViewSelf              bool                 `json:"is_view_self"` // 是否为作者本人观看
}

type VideoViewStat struct {
	Aid        int64  `json:"aid"`     // 视频aid
	View       int    `json:"view"`    // 视频播放量
	Danmaku    int    `json:"danmaku"` // 视频弹幕数
	Reply      int    `json:"reply"`
	Favorite   int    `json:"favorite"` // 视频收藏数
	Coin       int    `json:"coin"`     // 视频投币数
	Share      int    `json:"share"`
	NowRank    int    `json:"now_rank"`
	HisRank    int    `json:"his_rank"`
	Like       int    `json:"like"`    //	视频点赞数
	Dislike    int    `json:"dislike"` //	视频点踩数
	Evaluation string `json:"evaluation"`
	Vt         int    `json:"vt"`
}

// VideoViewRights 视频属性信息
type VideoViewRights struct {
	Bp            int `json:"bp"`       // 是否允许承包
	Elec          int `json:"elec"`     // 是否支持充电
	Download      int `json:"download"` // 是否允许下载
	Movie         int `json:"movie"`    // 是否电影
	Pay           int `json:"pay"`
	Hd5           int `json:"hd5"`        // 是否有高码率
	NoReprint     int `json:"no_reprint"` // 是否显示“禁止转载”信息
	Autoplay      int `json:"autoplay"`   // 是否自动播放
	UgcPay        int `json:"ugc_pay"`
	IsCooperation int `json:"is_cooperation"` // 是否为合作视频
	UgcPayPreview int `json:"ugc_pay_preview"`
	NoBackground  int `json:"no_background"`
	CleanMode     int `json:"clean_mode"`
	IsSteinGate   int `json:"is_stein_gate"` // 是否为全景视频
	Is360         int `json:"is_360"`
	NoShare       int `json:"no_share"`
	ArcPay        int `json:"arc_pay"`
	FreeWatch     int `json:"free_watch"`
}

type VideoViewAddit struct {
	Field1 bool `json:"63"`
	Field2 bool `json:"64"`
	Field3 bool `json:"69"`
	Field4 bool `json:"71"`
	Field5 bool `json:"72"`
}

// VideoEmergency 视频操作按钮信息
type VideoEmergency struct {
	NoLike  bool `json:"no_like"`  // 是否不显示点赞按钮
	NoCoin  bool `json:"no_coin"`  // 是否不显示投币按钮
	NoFav   bool `json:"no_fav"`   // 是否不显示收藏按钮
	NoShare bool `json:"no_share"` // 是否不显示分享按钮
}

type VideoHotShare struct {
	Show bool  `json:"show"`
	List []any `json:"list"`
}

type VideoReply struct {
	Page    any              `json:"page"`
	Replies []VideoReplyItem `json:"replies"`
}

type VideoReplyItem struct {
	Rpid       int  `json:"rpid"`
	Oid        int  `json:"oid"`
	Type       int  `json:"type"`
	Mid        int  `json:"mid"`
	Root       int  `json:"root"`
	Parent     int  `json:"parent"`
	Dialog     int  `json:"dialog"`
	Count      int  `json:"count"`
	Rcount     int  `json:"rcount"`
	State      int  `json:"state"`
	Fansgrade  int  `json:"fansgrade"`
	Attr       int  `json:"attr"`
	Ctime      int  `json:"ctime"`
	Like       int  `json:"like"`
	Action     int  `json:"action"`
	Content    any  `json:"content"`
	Replies    any  `json:"replies"`
	Assist     int  `json:"assist"`
	ShowFollow bool `json:"show_follow"`
}

// UgcSeason 视频合集信息，不在合集中的视频无此项
type UgcSeason struct {
	Id          int                    `json:"id"`
	Title       string                 `json:"title"`
	Cover       string                 `json:"cover"`
	Mid         int                    `json:"mid"`
	Intro       string                 `json:"intro"`
	SignState   int                    `json:"sign_state"`
	Attribute   int                    `json:"attribute"`
	Sections    []UgcSeasonSectionItem `json:"sections"`
	Stat        UgcSeasonStat          `json:"stat"`
	EpCount     int                    `json:"ep_count"`
	SeasonType  int                    `json:"season_type"`
	IsPaySeason bool                   `json:"is_pay_season"`
	EnableVt    int                    `json:"enable_vt"`
}

type UgcSeasonStat struct {
	SeasonId int `json:"season_id"`
	View     int `json:"view"`
	Danmaku  int `json:"danmaku"`
	Reply    int `json:"reply"`
	Fav      int `json:"fav"`
	Coin     int `json:"coin"`
	Share    int `json:"share"`
	NowRank  int `json:"now_rank"`
	HisRank  int `json:"his_rank"`
	Like     int `json:"like"`
	Vt       int `json:"vt"`
	Vv       int `json:"vv"`
}

type UgcSeasonSectionItem struct {
	SeasonId int       `json:"season_id"`
	Id       int       `json:"id"`
	Title    string    `json:"title"`
	Type     int       `json:"type"`
	Episodes []Episode `json:"episodes"`
}

type Episode struct {
	SeasonId  int                 `json:"season_id"`
	SectionId int                 `json:"section_id"`
	Id        int                 `json:"id"`
	Aid       int64               `json:"aid"`
	Cid       int64               `json:"cid"`
	Title     string              `json:"title"`
	Attribute int                 `json:"attribute"`
	Arc       Arc                 `json:"arc"`
	Page      UgcSeasonPageItem   `json:"page"`
	Bvid      string              `json:"bvid"`
	Pages     []UgcSeasonPageItem `json:"pages"`
}

type Arc struct {
	Aid                int64      `json:"aid"`
	Videos             int        `json:"videos"`
	TypeId             int        `json:"type_id"`
	TypeName           string     `json:"type_name"`
	Copyright          int        `json:"copyright"`
	Pic                string     `json:"pic"`
	Title              string     `json:"title"`
	Pubdate            int        `json:"pubdate"`
	Ctime              int        `json:"ctime"`
	Desc               string     `json:"desc"`
	State              int        `json:"state"`
	Duration           int        `json:"duration"`
	Rights             ArcRights  `json:"rights"`
	Author             VideoOwner `json:"author"`
	Stat               VideoStat  `json:"stat"`
	Dynamic            string     `json:"dynamic"`
	Dimension          Dimension  `json:"dimension"`
	DescV2             any        `json:"desc_v2"`
	IsChargeableSeason bool       `json:"is_chargeable_season"`
	IsBlooper          bool       `json:"is_blooper"`
	EnableVt           int        `json:"enable_vt"`
	VtDisplay          string     `json:"vt_display"`
	TypeIdV2           int        `json:"type_id_v2"`
	TypeNameV2         string     `json:"type_name_v2"`
	IsLessonVideo      int        `json:"is_lesson_video"`
}
type ArcRights struct {
	Bp            int `json:"bp"`
	Elec          int `json:"elec"`
	Download      int `json:"download"`
	Movie         int `json:"movie"`
	Pay           int `json:"pay"`
	Hd5           int `json:"hd5"`
	NoReprint     int `json:"no_reprint"`
	Autoplay      int `json:"autoplay"`
	UgcPay        int `json:"ugc_pay"`
	IsCooperation int `json:"is_cooperation"`
	UgcPayPreview int `json:"ugc_pay_preview"`
	ArcPay        int `json:"arc_pay"`
	FreeWatch     int `json:"free_watch"`
}

type VideoStat struct {
	Aid        int64  `json:"aid"`
	View       int    `json:"view"`
	Danmaku    int    `json:"danmaku"`
	Reply      int    `json:"reply"`
	Fav        int    `json:"fav"`
	Coin       int    `json:"coin"`
	Share      int    `json:"share"`
	NowRank    int    `json:"now_rank"`
	HisRank    int    `json:"his_rank"`
	Like       int    `json:"like"`
	Dislike    int    `json:"dislike"`
	Evaluation string `json:"evaluation"`
	ArgueMsg   string `json:"argue_msg"`
	Vt         int    `json:"vt"`
	Vv         int    `json:"vv"`
}

type UgcSeasonPageItem struct {
	Cid       int64     `json:"cid"`
	Page      int       `json:"page"`
	From      string    `json:"from"`
	Part      string    `json:"part"`
	Duration  int       `json:"duration"`
	Vid       string    `json:"vid"`
	Weblink   string    `json:"weblink"`
	Dimension Dimension `json:"dimension"`
}

type Dimension struct {
	Width  int `json:"width"`
	Height int `json:"height"`
	Rotate int `json:"rotate"`
}

type VideoPage struct {
	Cid        int64     `json:"cid"`
	Page       int       `json:"page"`     // 分P页码，从1开始
	From       string    `json:"from"`     // 视频来源，如 "vupload"（UP主上传）， hunan：芒果TV，qq： 腾讯视频
	Part       string    `json:"part"`     // 分P标题
	Duration   int       `json:"duration"` // 分P时长，单位秒
	Vid        string    `json:"vid"`      // 站外视频vid
	Weblink    string    `json:"weblink"`  // 站外视频跳转url
	Dimension  Dimension `json:"dimension"`
	FirstFrame string    `json:"first_frame"`
	Ctime      int       `json:"ctime"`
}

type VideoSubTitle struct {
	AllowSubmit bool  `json:"allow_submit"`
	List        []any `json:"list"`
}
