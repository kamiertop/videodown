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

// VideoRelatedItem 推荐视频
type VideoRelatedItem struct {
	Aid       int64  `json:"aid"`
	Videos    int    `json:"videos"`
	Tid       int    `json:"tid"`
	Tname     string `json:"tname"`
	Copyright int    `json:"copyright"`
	Pic       string `json:"pic"`
	Title     string `json:"title"`
	Pubdate   int    `json:"pubdate"`
	Ctime     int    `json:"ctime"`
	Desc      string `json:"desc"`
	State     int    `json:"state"`
	Duration  int    `json:"duration"`
	Rights    struct {
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
		NoBackground  int `json:"no_background"`
		ArcPay        int `json:"arc_pay"`
		PayFreeWatch  int `json:"pay_free_watch"`
	} `json:"rights"`
	Owner       VideoOwner    `json:"owner"`
	Stat        VideoStat     `json:"stat"`
	Dynamic     string        `json:"dynamic"`
	Cid         int64         `json:"cid"`
	Dimension   Dimension     `json:"dimension"`
	ShortLinkV2 string        `json:"short_link_v2"`
	FirstFrame  string        `json:"first_frame"`
	PubLocation string        `json:"pub_location"`
	Cover43     string        `json:"cover43"`
	Tidv2       int           `json:"tidv2"`
	Tnamev2     string        `json:"tnamev2"`
	PidV2       int           `json:"pid_v2"`
	PidNameV2   string        `json:"pid_name_v2"`
	Bvid        string        `json:"bvid"`
	SeasonType  int           `json:"season_type"`
	IsOgv       bool          `json:"is_ogv"`
	OgvInfo     any           `json:"ogv_info"`
	RcmdReason  string        `json:"rcmd_reason"`
	EnableVt    int           `json:"enable_vt"`
	SeasonId    int           `json:"season_id,omitempty"`
	AiRcmd      RelatedAiRcmd `json:"ai_rcmd"`
	UpFromV2    int           `json:"up_from_v2,omitempty"`
	MissionId   int           `json:"mission_id,omitempty"`
}

type RelatedAiRcmd struct {
	Id      int64  `json:"id"`
	Goto    string `json:"goto"`
	Trackid string `json:"trackid"`
	UniqId  string `json:"uniq_id"`
}

type VideoTag struct {
	TagId   int    `json:"tag_id"`
	TagName string `json:"tag_name"`
	MusicId string `json:"music_id"`
	TagType string `json:"tag_type"`
	JumpUrl string `json:"jump_url"`
}

type VideoCardSpace struct {
	SImg string `json:"s_img"`
	LImg string `json:"l_img"`
}

type VideoCardEntry struct {
	Mid            string              `json:"mid"`
	Name           string              `json:"name"`
	Approve        bool                `json:"approve"`
	Sex            string              `json:"sex"`
	Rank           string              `json:"rank"`
	Face           string              `json:"face"`
	FaceNft        int                 `json:"face_nft"`
	FaceNftType    int                 `json:"face_nft_type"`
	DisplayRank    string              `json:"DisplayRank"`
	Regtime        int                 `json:"regtime"`
	Spacesta       int                 `json:"spacesta"`
	Birthday       string              `json:"birthday"`
	Place          string              `json:"place"`
	Description    string              `json:"description"`
	Article        int                 `json:"article"`
	Attentions     []any               `json:"attentions"`
	Fans           int                 `json:"fans"`
	Friend         int                 `json:"friend"`
	Attention      int                 `json:"attention"`
	Sign           string              `json:"sign"`
	LevelInfo      VideoLevelInfo      `json:"level_info"`
	Pendant        VideoPendant        `json:"pendant"`
	Nameplate      VideoNameplate      `json:"nameplate"`
	Official       VideoOfficial       `json:"Official"`
	OfficialVerify VideoOfficialVerify `json:"official_verify"`
	Vip            struct {
		Type       int `json:"type"`
		Status     int `json:"status"`
		DueDate    int `json:"due_date"`
		VipPayType int `json:"vip_pay_type"`
		ThemeType  int `json:"theme_type"`
		Label      struct {
			Path                  string `json:"path"`
			Text                  string `json:"text"`
			LabelTheme            string `json:"label_theme"`
			TextColor             string `json:"text_color"`
			BgStyle               int    `json:"bg_style"`
			BgColor               string `json:"bg_color"`
			BorderColor           string `json:"border_color"`
			UseImgLabel           bool   `json:"use_img_label"`
			ImgLabelUriHans       string `json:"img_label_uri_hans"`
			ImgLabelUriHant       string `json:"img_label_uri_hant"`
			ImgLabelUriHansStatic string `json:"img_label_uri_hans_static"`
			ImgLabelUriHantStatic string `json:"img_label_uri_hant_static"`
			LabelId               int    `json:"label_id"`
			LabelGoto             any    `json:"label_goto"`
		} `json:"label"`
		AvatarSubscript    int    `json:"avatar_subscript"`
		NicknameColor      string `json:"nickname_color"`
		Role               int    `json:"role"`
		AvatarSubscriptUrl string `json:"avatar_subscript_url"`
		TvVipStatus        int    `json:"tv_vip_status"`
		TvVipPayType       int    `json:"tv_vip_pay_type"`
		TvDueDate          int    `json:"tv_due_date"`
		AvatarIcon         struct {
			IconResource struct {
			} `json:"icon_resource"`
		} `json:"avatar_icon"`
		OttInfo   OttInfo  `json:"ott_info"`
		SuperVip  SuperVip `json:"super_vip"`
		VipType   int      `json:"vipType"`
		VipStatus int      `json:"vipStatus"`
	} `json:"vip"`
	IsSeniorMember int `json:"is_senior_member"`
	NameRender     any `json:"name_render"`
}

type AvatarIcon struct {
}

type SuperVip struct {
	IsSuperVip bool `json:"is_super_vip"`
}

type OttInfo struct {
	VipType      int    `json:"vip_type"`
	PayType      int    `json:"pay_type"`
	PayChannelId string `json:"pay_channel_id"`
	Status       int    `json:"status"`
	OverdueTime  int    `json:"overdue_time"`
}

type VideoLevelInfo struct {
	CurrentLevel int `json:"current_level"`
	CurrentMin   int `json:"current_min"`
	CurrentExp   int `json:"current_exp"`
	NextExp      int `json:"next_exp"`
}

type VideoPendant struct {
	Pid               int    `json:"pid"`
	Name              string `json:"name"`
	Image             string `json:"image"`
	Expire            int    `json:"expire"`
	ImageEnhance      string `json:"image_enhance"`
	ImageEnhanceFrame string `json:"image_enhance_frame"`
	NPid              int    `json:"n_pid"`
}

type VideoNameplate struct {
	Nid        int    `json:"nid"`
	Name       string `json:"name"`
	Image      string `json:"image"`
	ImageSmall string `json:"image_small"`
	Level      string `json:"level"`
	Condition  string `json:"condition"`
}

type VideoOfficial struct {
	Role  int    `json:"role"`
	Title string `json:"title"`
	Desc  string `json:"desc"`
	Type  int    `json:"type"`
}

type VideoOfficialVerify struct {
	Type int    `json:"type"`
	Desc string `json:"desc"`
}

type VideoDetailConcise struct {
}

type UserGarb struct {
	UrlImageAniCut string `json:"url_image_ani_cut"`
}

type VideoDescV2 struct {
	RawText string `json:"raw_text"`
	Type    int    `json:"type"`
	BizId   int    `json:"biz_id"`
}

type VideoOwner struct {
	Mid  int    `json:"mid"`
	Name string `json:"name"`
	Face string `json:"face"`
}

type VideoDetailArgueInfo struct {
	ArgueMsg  string `json:"argue_msg"`
	ArgueType int    `json:"argue_type"`
	ArgueLink string `json:"argue_link"`
}
