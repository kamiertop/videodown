package model

// VideoDetailConciseData 只保留必要的基本信息
type VideoDetailConciseData struct {
	View VideoView `json:"view"`
}

type VideoView struct {
	Bvid     string              `json:"bvid"`
	Aid      int64               `json:"aid"`
	Pic      string              `json:"pic"`
	Title    string              `json:"title"`
	Pubdate  int                 `json:"pubdate"`
	Duration int                 `json:"duration"`
	Cid      int64               `json:"cid"`
	SeasonId int                 `json:"season_id"`
	Pages    []VideoViewPageItem `json:"pages"`
}

// VideoDetailResponse 完整的响应
type VideoDetailResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Ttl     int    `json:"ttl"`
	Data    struct {
		View struct {
			Bvid      string `json:"bvid"`
			Aid       int64  `json:"aid"`
			Videos    int    `json:"videos"`
			Tid       int    `json:"tid"`
			TidV2     int    `json:"tid_v2"`
			Tname     string `json:"tname"`
			TnameV2   string `json:"tname_v2"`
			Copyright int    `json:"copyright"`
			Pic       string `json:"pic"`
			Title     string `json:"title"`
			Pubdate   int    `json:"pubdate"`
			Ctime     int    `json:"ctime"`
			Desc      string `json:"desc"`
			DescV2    []struct {
				RawText string `json:"raw_text"`
				Type    int    `json:"type"`
				BizId   int    `json:"biz_id"`
			} `json:"desc_v2"`
			State     int `json:"state"`
			Duration  int `json:"duration"`
			MissionId int `json:"mission_id"`
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
				CleanMode     int `json:"clean_mode"`
				IsSteinGate   int `json:"is_stein_gate"`
				Is360         int `json:"is_360"`
				NoShare       int `json:"no_share"`
				ArcPay        int `json:"arc_pay"`
				FreeWatch     int `json:"free_watch"`
			} `json:"rights"`
			Owner VideoDetailOwner `json:"owner"`
			Stat  struct {
				Aid        int64  `json:"aid"`
				View       int    `json:"view"`
				Danmaku    int    `json:"danmaku"`
				Reply      int    `json:"reply"`
				Favorite   int    `json:"favorite"`
				Coin       int    `json:"coin"`
				Share      int    `json:"share"`
				NowRank    int    `json:"now_rank"`
				HisRank    int    `json:"his_rank"`
				Like       int    `json:"like"`
				Dislike    int    `json:"dislike"`
				Evaluation string `json:"evaluation"`
				Vt         int    `json:"vt"`
			} `json:"stat"`
			ArgueInfo struct {
				ArgueMsg  string `json:"argue_msg"`
				ArgueType int    `json:"argue_type"`
				ArgueLink string `json:"argue_link"`
			} `json:"argue_info"`
			Dynamic                 string              `json:"dynamic"`
			Cid                     int64               `json:"cid"`
			Dimension               Dimension           `json:"dimension"`
			SeasonId                int                 `json:"season_id"`
			Premiere                interface{}         `json:"premiere"`
			TeenageMode             int                 `json:"teenage_mode"`
			IsChargeableSeason      bool                `json:"is_chargeable_season"`
			IsStory                 bool                `json:"is_story"`
			IsUpowerExclusive       bool                `json:"is_upower_exclusive"`
			IsUpowerPlay            bool                `json:"is_upower_play"`
			IsUpowerPreview         bool                `json:"is_upower_preview"`
			EnableVt                int                 `json:"enable_vt"`
			VtDisplay               string              `json:"vt_display"`
			IsUpowerExclusiveWithQa bool                `json:"is_upower_exclusive_with_qa"`
			NoCache                 bool                `json:"no_cache"`
			Pages                   []VideoViewPageItem `json:"pages"`
			Subtitle                VideoViewSubTitle   `json:"subtitle"`
			UgcSeason               UgcSeason           `json:"ugc_season"`
			IsSeasonDisplay         bool                `json:"is_season_display"`
			UserGarb                struct {
				UrlImageAniCut string `json:"url_image_ani_cut"`
			} `json:"user_garb"`
			HonorReply struct {
			} `json:"honor_reply"`
			LikeIcon          string `json:"like_icon"`
			NeedJumpBv        bool   `json:"need_jump_bv"`
			DisableShowUpInfo bool   `json:"disable_show_up_info"`
			IsStoryPlay       int    `json:"is_story_play"`
			IsViewSelf        bool   `json:"is_view_self"`
		} `json:"View"`
		Card struct {
			Card struct {
				Mid         string        `json:"mid"`
				Name        string        `json:"name"`
				Approve     bool          `json:"approve"`
				Sex         string        `json:"sex"`
				Rank        string        `json:"rank"`
				Face        string        `json:"face"`
				FaceNft     int           `json:"face_nft"`
				FaceNftType int           `json:"face_nft_type"`
				DisplayRank string        `json:"DisplayRank"`
				Regtime     int           `json:"regtime"`
				Spacesta    int           `json:"spacesta"`
				Birthday    string        `json:"birthday"`
				Place       string        `json:"place"`
				Description string        `json:"description"`
				Article     int           `json:"article"`
				Attentions  []interface{} `json:"attentions"`
				Fans        int           `json:"fans"`
				Friend      int           `json:"friend"`
				Attention   int           `json:"attention"`
				Sign        string        `json:"sign"`
				LevelInfo   struct {
					CurrentLevel int `json:"current_level"`
					CurrentMin   int `json:"current_min"`
					CurrentExp   int `json:"current_exp"`
					NextExp      int `json:"next_exp"`
				} `json:"level_info"`
				Pendant struct {
					Pid               int    `json:"pid"`
					Name              string `json:"name"`
					Image             string `json:"image"`
					Expire            int    `json:"expire"`
					ImageEnhance      string `json:"image_enhance"`
					ImageEnhanceFrame string `json:"image_enhance_frame"`
					NPid              int    `json:"n_pid"`
				} `json:"pendant"`
				Nameplate struct {
					Nid        int    `json:"nid"`
					Name       string `json:"name"`
					Image      string `json:"image"`
					ImageSmall string `json:"image_small"`
					Level      string `json:"level"`
					Condition  string `json:"condition"`
				} `json:"nameplate"`
				Official struct {
					Role  int    `json:"role"`
					Title string `json:"title"`
					Desc  string `json:"desc"`
					Type  int    `json:"type"`
				} `json:"Official"`
				OfficialVerify struct {
					Type int    `json:"type"`
					Desc string `json:"desc"`
				} `json:"official_verify"`
				Vip struct {
					Type       int `json:"type"`
					Status     int `json:"status"`
					DueDate    int `json:"due_date"`
					VipPayType int `json:"vip_pay_type"`
					ThemeType  int `json:"theme_type"`
					Label      struct {
						Path                  string      `json:"path"`
						Text                  string      `json:"text"`
						LabelTheme            string      `json:"label_theme"`
						TextColor             string      `json:"text_color"`
						BgStyle               int         `json:"bg_style"`
						BgColor               string      `json:"bg_color"`
						BorderColor           string      `json:"border_color"`
						UseImgLabel           bool        `json:"use_img_label"`
						ImgLabelUriHans       string      `json:"img_label_uri_hans"`
						ImgLabelUriHant       string      `json:"img_label_uri_hant"`
						ImgLabelUriHansStatic string      `json:"img_label_uri_hans_static"`
						ImgLabelUriHantStatic string      `json:"img_label_uri_hant_static"`
						LabelId               int         `json:"label_id"`
						LabelGoto             interface{} `json:"label_goto"`
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
					OttInfo struct {
						VipType      int    `json:"vip_type"`
						PayType      int    `json:"pay_type"`
						PayChannelId string `json:"pay_channel_id"`
						Status       int    `json:"status"`
						OverdueTime  int    `json:"overdue_time"`
					} `json:"ott_info"`
					SuperVip struct {
						IsSuperVip bool `json:"is_super_vip"`
					} `json:"super_vip"`
					VipType   int `json:"vipType"`
					VipStatus int `json:"vipStatus"`
				} `json:"vip"`
				IsSeniorMember int         `json:"is_senior_member"`
				NameRender     interface{} `json:"name_render"`
			} `json:"card"`
			Space struct {
				SImg string `json:"s_img"`
				LImg string `json:"l_img"`
			} `json:"space"`
			Following    bool `json:"following"`
			ArchiveCount int  `json:"archive_count"`
			ArticleCount int  `json:"article_count"`
			Follower     int  `json:"follower"`
			LikeNum      int  `json:"like_num"`
		} `json:"Card"`
		Tags []struct {
			TagId   int    `json:"tag_id"`
			TagName string `json:"tag_name"`
			MusicId string `json:"music_id"`
			TagType string `json:"tag_type"`
			JumpUrl string `json:"jump_url"`
		} `json:"Tags"`
		Reply struct {
			Page    interface{} `json:"page"`
			Replies []struct {
				Rpid       int         `json:"rpid"`
				Oid        int         `json:"oid"`
				Type       int         `json:"type"`
				Mid        int         `json:"mid"`
				Root       int         `json:"root"`
				Parent     int         `json:"parent"`
				Dialog     int         `json:"dialog"`
				Count      int         `json:"count"`
				Rcount     int         `json:"rcount"`
				State      int         `json:"state"`
				Fansgrade  int         `json:"fansgrade"`
				Attr       int         `json:"attr"`
				Ctime      int         `json:"ctime"`
				Like       int         `json:"like"`
				Action     int         `json:"action"`
				Content    interface{} `json:"content"`
				Replies    interface{} `json:"replies"`
				Assist     int         `json:"assist"`
				ShowFollow bool        `json:"show_follow"`
			} `json:"replies"`
		} `json:"Reply"`
		Related []struct {
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
			Owner VideoDetailOwner `json:"owner"`
			Stat  struct {
				Aid      int64 `json:"aid"`
				View     int   `json:"view"`
				Danmaku  int   `json:"danmaku"`
				Reply    int   `json:"reply"`
				Favorite int   `json:"favorite"`
				Coin     int   `json:"coin"`
				Share    int   `json:"share"`
				NowRank  int   `json:"now_rank"`
				HisRank  int   `json:"his_rank"`
				Like     int   `json:"like"`
				Dislike  int   `json:"dislike"`
				Vt       int   `json:"vt"`
				Vv       int   `json:"vv"`
				FavG     int   `json:"fav_g"`
				LikeG    int   `json:"like_g"`
			} `json:"stat"`
			Dynamic     string      `json:"dynamic"`
			Cid         int64       `json:"cid"`
			Dimension   Dimension   `json:"dimension"`
			SeasonId    int         `json:"season_id,omitempty"`
			ShortLinkV2 string      `json:"short_link_v2"`
			FirstFrame  string      `json:"first_frame"`
			PubLocation string      `json:"pub_location"`
			Cover43     string      `json:"cover43"`
			Tidv2       int         `json:"tidv2"`
			Tnamev2     string      `json:"tnamev2"`
			PidV2       int         `json:"pid_v2"`
			PidNameV2   string      `json:"pid_name_v2"`
			Bvid        string      `json:"bvid"`
			SeasonType  int         `json:"season_type"`
			IsOgv       bool        `json:"is_ogv"`
			OgvInfo     interface{} `json:"ogv_info"`
			RcmdReason  string      `json:"rcmd_reason"`
			EnableVt    int         `json:"enable_vt"`
			AiRcmd      struct {
				Id      int64  `json:"id"`
				Goto    string `json:"goto"`
				Trackid string `json:"trackid"`
				UniqId  string `json:"uniq_id"`
			} `json:"ai_rcmd"`
			UpFromV2  int `json:"up_from_v2,omitempty"`
			MissionId int `json:"mission_id,omitempty"`
		} `json:"Related"` // 推荐视频列表
		Spec     interface{} `json:"Spec"`
		HotShare struct {
			Show bool          `json:"show"`
			List []interface{} `json:"list"`
		} `json:"hot_share"`
		Elec      interface{} `json:"elec"`
		Emergency struct {
			NoLike  bool `json:"no_like"`
			NoCoin  bool `json:"no_coin"`
			NoFav   bool `json:"no_fav"`
			NoShare bool `json:"no_share"`
		} `json:"emergency"`
		ViewAddit struct {
			Field1 bool `json:"63"`
			Field2 bool `json:"64"`
			Field3 bool `json:"69"`
			Field4 bool `json:"71"`
			Field5 bool `json:"72"`
		} `json:"view_addit"`
		Guide                  interface{} `json:"guide"`
		QueryTags              interface{} `json:"query_tags"`
		Participle             []string    `json:"participle"`
		ModuleCtrl             interface{} `json:"module_ctrl"`
		ReplaceRecommend       bool        `json:"replace_recommend"`
		IsHitLabourDayActivity bool        `json:"is_hit_labour_day_activity"`
	} `json:"data"`
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
	SeasonId int    `json:"season_id"`
	Id       int    `json:"id"`
	Title    string `json:"title"`
	Type     int    `json:"type"`
	Episodes []struct {
		SeasonId  int    `json:"season_id"`
		SectionId int    `json:"section_id"`
		Id        int    `json:"id"`
		Aid       int64  `json:"aid"`
		Cid       int64  `json:"cid"`
		Title     string `json:"title"`
		Attribute int    `json:"attribute"`
		Arc       struct {
			Aid       int64  `json:"aid"`
			Videos    int    `json:"videos"`
			TypeId    int    `json:"type_id"`
			TypeName  string `json:"type_name"`
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
				ArcPay        int `json:"arc_pay"`
				FreeWatch     int `json:"free_watch"`
			} `json:"rights"`
			Author struct {
				Mid  int    `json:"mid"`
				Name string `json:"name"`
				Face string `json:"face"`
			} `json:"author"`
			Stat struct {
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
			} `json:"stat"`
			Dynamic            string      `json:"dynamic"`
			Dimension          Dimension   `json:"dimension"`
			DescV2             interface{} `json:"desc_v2"`
			IsChargeableSeason bool        `json:"is_chargeable_season"`
			IsBlooper          bool        `json:"is_blooper"`
			EnableVt           int         `json:"enable_vt"`
			VtDisplay          string      `json:"vt_display"`
			TypeIdV2           int         `json:"type_id_v2"`
			TypeNameV2         string      `json:"type_name_v2"`
			IsLessonVideo      int         `json:"is_lesson_video"`
		} `json:"arc"`
		Page  UgcSeasonPageItem   `json:"page"`
		Bvid  string              `json:"bvid"`
		Pages []UgcSeasonPageItem `json:"pages"`
	} `json:"episodes"`
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

type VideoViewPageItem struct {
	Cid        int64     `json:"cid"`
	Page       int       `json:"page"`
	From       string    `json:"from"`
	Part       string    `json:"part"`
	Duration   int       `json:"duration"`
	Vid        string    `json:"vid"`
	Weblink    string    `json:"weblink"`
	Dimension  Dimension `json:"dimension"`
	FirstFrame string    `json:"first_frame"`
	Ctime      int       `json:"ctime"`
}

type VideoViewSubTitle struct {
	AllowSubmit bool  `json:"allow_submit"`
	List        []any `json:"list"`
}
