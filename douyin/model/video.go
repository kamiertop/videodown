package model

// ImageVideo 动图Video结构
type ImageVideo struct {
	Cover        Cover         `json:"cover"`          // 视频封面
	OriginCover  Cover         `json:"origin_cover"`   // 视频原始封面
	BitRate      []BitRateItem `json:"bit_rate"`       // 视频不同质量的播放地址信息列表, 和play_addr中的一样
	Duration     int64         `json:"duration"`       // 视频时长, 单位为秒
	Height       uint32        `json:"height"`         // 视频高度
	Width        uint32        `json:"width"`          // 视频宽度
	PlayAddrH264 PlayInfo      `json:"play_addr_h264"` // 视频播放地址信息
	PlayAddr     PlayInfo      `json:"play_addr"`      // 视频播放地址信息
	HasWatermark bool          `json:"has_watermark"`  // 是否有水印
}

// AwemeItem 视频信息
type AwemeItem struct {
	AwemeId                  string          `json:"aweme_id"`    // 视频ID
	Desc                     string          `json:"desc"`        // 视频标题+#标签
	CreateTime               int             `json:"create_time"` // 视频创建时间
	Author                   Author          `json:"author"`      // 视频作者信息
	Music                    Music           `json:"music"`
	Video                    Video           `json:"video"` // 视频信息
	Statistics               VideoStatistics `json:"statistics"`
	Duration                 int             `json:"duration"` // 视频时长, 单位为毫秒
	AwemeType                int             `json:"aweme_type"`
	AuthorUserId             int64           `json:"author_user_id"` // 视频作者的用户ID
	Region                   string          `json:"region"`         // 视频发布的地区:CN等
	GroupId                  string          `json:"group_id"`
	PreventDownload          bool            `json:"prevent_download"`
	IsMomentHistory          int             `json:"is_moment_history"`
	IsMomentStory            int             `json:"is_moment_story"`
	SecItemId                string          `json:"sec_item_id"`
	ItemAigcFollowShot       int             `json:"item_aigc_follow_shot"`
	Images                   []ImageItem     `json:"images"` // 图文类型视频使用这个字段中的图片列表
	OriginDuetResourceUri    string          `json:"origin_duet_resource_uri"`
	IsImageBeat              bool            `json:"is_image_beat"`
	IsLifeItem               bool            `json:"is_life_item"`
	AuthorMaskTag            int             `json:"author_mask_tag"`
	UserRecommendStatus      int             `json:"user_recommend_status"`
	CollectionCornerMark     int             `json:"collection_corner_mark"`
	IsSharePost              bool            `json:"is_share_post"`
	AuthenticationToken      string          `json:"authentication_token"` // 视频的认证token
	MediaType                int             `json:"media_type"`           // 视频类型，2图文，4视频
	ActivityVideoType        int             `json:"activity_video_type"`
	BoostStatus              int             `json:"boost_status"`
	Caption                  string          `json:"caption"`    // 视频标题，带标签，可能为空
	ItemTitle                string          `json:"item_title"` // 子标题，可能为空
	Original                 int             `json:"original"`
	LunaVideoCandidateStatus string          `json:"luna_video_candidate_status,omitempty"`
	IsMultiContent           int             `json:"is_multi_content,omitempty"`
	ImageItemQualityLevel    int             `json:"image_item_quality_level,omitempty"`
	IsLivePhoto              int             `json:"is_live_photo"` // 是否是动图，1是，0不是
	IsSliedes                bool            `json:"is_slides"`     // 有这个字段，一定是动图，并且是多张动图，没有这个的也不一定不是动图
}

// Deprecated: JingXuan 精选视频结构
type JingXuan struct {
	EntertainmentVideoType int `json:"entertainmentVideoType"`
	ArticleInfo            struct {
		IsCartoon bool `json:"isCartoon"`
		HasMore   bool `json:"hasMore"`
	} `json:"articleInfo"`
	AwemeId     string `json:"awemeId"`
	AwemeType   int    `json:"awemeType"`
	GroupId     string `json:"groupId"`
	ChapterInfo struct {
		Status bool `json:"status"`
		List   []struct {
			Desc      string `json:"desc"`
			Detail    string `json:"detail"`
			Timestamp int    `json:"timestamp"`
			Url       string `json:"url"`
		} `json:"list"`
		ChapterAbstract     string `json:"chapterAbstract"`
		IsPoint             bool   `json:"isPoint"`
		UseRecommendChapter bool   `json:"useRecommendChapter"`
		ShowAiTag           bool   `json:"showAiTag"`
	} `json:"chapterInfo"`
	AuthorInfo struct {
		Uid                    string `json:"uid"`
		IsAdFake               bool   `json:"isAdFake"`
		SecUid                 string `json:"secUid"`
		Nickname               string `json:"nickname"`
		RemarkName             string `json:"remarkName"`
		AvatarUri              string `json:"avatarUri"`
		FollowerCount          int    `json:"followerCount"`
		TotalFavorited         int    `json:"totalFavorited"`
		FollowStatus           int    `json:"followStatus"`
		FollowerStatus         int    `json:"followerStatus"`
		EnterpriseVerifyReason string `json:"enterpriseVerifyReason"`
		CustomVerify           string `json:"customVerify"`
		RoomData               struct {
		} `json:"roomData"`
		AvatarThumb struct {
			Height  int      `json:"height"`
			Width   int      `json:"width"`
			Uri     string   `json:"uri"`
			UrlList []string `json:"urlList"`
		} `json:"avatarThumb"`
		RoleTitle       string `json:"roleTitle"`
		Secret          int    `json:"secret"`
		AccountCertInfo struct {
		} `json:"accountCertInfo"`
	} `json:"authorInfo"`
	Desc                string `json:"desc"`
	ItemTitle           string `json:"itemTitle"`
	Caption             string `json:"caption"`
	AuthorUserId        int64  `json:"authorUserId"`
	AuthenticationToken string `json:"authenticationToken"`
	PcNeedLogin         bool   `json:"pcNeedLogin"`
	MediaType           int    `json:"mediaType"`
	CreateTime          int    `json:"createTime"`
	Video               struct {
		Width    int    `json:"width"`
		Height   int    `json:"height"`
		Ratio    string `json:"ratio"`
		Duration int    `json:"duration"`
		DataSize int    `json:"dataSize"`
		Uri      string `json:"uri"`
		PlayAddr []struct {
			Src string `json:"src"`
		} `json:"playAddr"`
		PlayAddrSize     int    `json:"playAddrSize"`
		PlayAddrFileHash string `json:"playAddrFileHash"`
		PlayApi          string `json:"playApi"`
		PlayAddrH265     []struct {
			Src string `json:"src"`
		} `json:"playAddrH265"`
		PlayAddrH265Size     int    `json:"playAddrH265Size"`
		PlayAddrH265FileHash string `json:"playAddrH265FileHash"`
		PlayApiH265          string `json:"playApiH265"`
		BitRateList          []struct {
			Uri      string `json:"uri"`
			DataSize int    `json:"dataSize"`
			Width    int    `json:"width"`
			Height   int    `json:"height"`
			PlayAddr []struct {
				Src string `json:"src"`
			} `json:"playAddr"`
			PlayApi         string `json:"playApi"`
			IsH265          int    `json:"isH265"`
			QualityType     int    `json:"qualityType"`
			BitRate         int    `json:"bitRate"`
			VideoFormat     string `json:"videoFormat"`
			GearName        string `json:"gearName"`
			Fps             int    `json:"fps"`
			PlayerAccessKey string `json:"playerAccessKey"`
			HDRBit          string `json:"HDRBit"`
			HDRType         string `json:"HDRType"`
			FeatureId       string `json:"featureId"`
			Format          string `json:"format"`
			FileId          string `json:"fileId"`
			PktOffsetMap    []struct {
				Time   int `json:"time"`
				Offset int `json:"offset"`
			} `json:"pktOffsetMap"`
			RealBitrate int `json:"realBitrate"`
			Mvmaf       struct {
				MvmafSrV1080  float64 `json:"mvmaf_sr_v1080,omitempty"`
				MvmafSrV960   float64 `json:"mvmaf_sr_v960,omitempty"`
				MvmafSrV864   float64 `json:"mvmaf_sr_v864,omitempty"`
				MvmafSrV720   float64 `json:"mvmaf_sr_v720,omitempty"`
				MvmafOriV1080 float64 `json:"mvmaf_ori_v1080,omitempty"`
				MvmafOriV960  float64 `json:"mvmaf_ori_v960,omitempty"`
				MvmafOriV864  float64 `json:"mvmaf_ori_v864,omitempty"`
				MvmafOriV720  float64 `json:"mvmaf_ori_v720,omitempty"`
			} `json:"mvmaf"`
			Ufq struct {
				Enh      float64 `json:"enh,omitempty"`
				Playback struct {
					Ori  float64 `json:"ori"`
					Srv1 float64 `json:"srv1"`
				} `json:"playback,omitempty"`
				Src     float64 `json:"src,omitempty"`
				Trans   float64 `json:"trans,omitempty"`
				Version string  `json:"version,omitempty"`
			} `json:"ufq"`
			SrSharpnessStrength int    `json:"srSharpnessStrength,omitempty"`
			IndexRange          string `json:"indexRange,omitempty"`
			InitRange           string `json:"initRange,omitempty"`
			AudioFileId         string `json:"audioFileId,omitempty"`
			FirstSegmentRange   string `json:"firstSegmentRange,omitempty"`
			AudioChannels       string `json:"audioChannels,omitempty"`
			AudioSampleRate     string `json:"audioSampleRate,omitempty"`
		} `json:"bitRateList"`
		BitRateAudioList []struct {
			RealBitrate  int    `json:"realBitrate"`
			AudioQuality int    `json:"audioQuality"`
			Bitrate      int    `json:"bitrate"`
			CodecType    string `json:"codecType"`
			FileHash     string `json:"fileHash"`
			FileId       string `json:"fileId"`
			LogoType     string `json:"logoType"`
			MediaType    string `json:"mediaType"`
			Quality      string `json:"quality"`
			Size         int    `json:"size"`
			QualityDesc  string `json:"qualityDesc"`
			UrlList      []struct {
				Src string `json:"src"`
			} `json:"urlList"`
			IndexRange        string `json:"indexRange"`
			InitRange         string `json:"initRange"`
			CheckInfo         string `json:"checkInfo"`
			FirstSegmentRange string `json:"firstSegmentRange"`
		} `json:"bitRateAudioList"`
		Cover              string        `json:"cover"`
		CoverUrlList       []string      `json:"coverUrlList"`
		Cover169UrlList    []interface{} `json:"cover169UrlList"`
		Cover169BigUrlList []interface{} `json:"cover169BigUrlList"`
		CoverUri           string        `json:"coverUri"`
		DynamicCover       string        `json:"dynamicCover"`
		OriginCover        string        `json:"originCover"`
		RawCover           string        `json:"rawCover"`
		OriginCoverUrlList []string      `json:"originCoverUrlList"`
		GaussianCover      string        `json:"gaussianCover"`
		Meta               struct {
			AudioBitrateSet       string `json:"audio_bitrate_set"`
			AudioBitrateValues    string `json:"audio_bitrate_values"`
			BrightRatioMean       string `json:"bright_ratio_mean"`
			BrightnessMean        string `json:"brightness_mean"`
			DiffOverexposureRatio string `json:"diff_overexposure_ratio"`
			EnableManualLadder    string `json:"enable_manual_ladder"`
			Format                string `json:"format"`
			GearVqm               string `json:"gear_vqm"`
			Hrids                 string `json:"hrids"`
			IsSpatialVideo        string `json:"is_spatial_video"`
			Isad                  string `json:"isad"`
			Loudness              string `json:"loudness"`
			OverexposureRatioMean string `json:"overexposure_ratio_mean"`
			Peak                  string `json:"peak"`
			PlayTimeProbDist      string `json:"play_time_prob_dist"`
			Qprf                  string `json:"qprf"`
			Sdgs                  string `json:"sdgs"`
			SrPotential           string `json:"sr_potential"`
			SrScore               string `json:"sr_score"`
			StdBrightness         string `json:"std_brightness"`
			StrategyTokens        string `json:"strategy_tokens"`
			TitleInfo             string `json:"title_info"`
			VideoMeta             string `json:"video_meta"`
			VolumeInfo            string `json:"volume_info"`
			VqsOrigin             string `json:"vqs_origin"`
		} `json:"meta"`
		BigThumbs []struct {
			Duration float64  `json:"duration"`
			Fext     string   `json:"fext"`
			ImgNum   int      `json:"img_num"`
			ImgUrl   string   `json:"img_url"`
			ImgUrls  []string `json:"img_urls"`
			ImgXLen  int      `json:"img_x_len"`
			ImgXSize int      `json:"img_x_size"`
			ImgYLen  int      `json:"img_y_len"`
			ImgYSize int      `json:"img_y_size"`
			Interval int      `json:"interval"`
			Uri      string   `json:"uri"`
			Uris     []string `json:"uris"`
		} `json:"bigThumbs"`
		VideoModel interface{} `json:"videoModel"`
	} `json:"video"`
	MixInfo struct {
		Cover             string `json:"cover"`
		CoverUri          string `json:"coverUri"`
		MixId             string `json:"mixId"`
		MixName           string `json:"mixName"`
		Desc              string `json:"desc"`
		Status            int    `json:"status"`
		PlayVV            int    `json:"playVV"`
		CollectVV         int    `json:"collectVV"`
		CurrentEpisode    int    `json:"currentEpisode"`
		TotalEpisode      int    `json:"totalEpisode"`
		HasUpdatedEpisode int    `json:"hasUpdatedEpisode"`
		IsCollected       int    `json:"isCollected"`
		WatchedItem       string `json:"watchedItem"`
		UpdateTime        string `json:"updateTime"`
		Author            struct {
			Uid         string `json:"uid"`
			Secret      int    `json:"secret"`
			IsBlockedV2 bool   `json:"isBlockedV2"`
			UserNotSee  int    `json:"userNotSee"`
		} `json:"author"`
		IsSeries2Mix bool `json:"isSeries2Mix"`
		ChargeInfo   struct {
			BuySchema         string `json:"buySchema"`
			ChargeType        int    `json:"chargeType"`
			ChargeCount       int    `json:"chargeCount"`
			HasPaid           bool   `json:"hasPaid"`
			PromiseUpdateTime int    `json:"promiseUpdateTime"`
			SellType          int    `json:"sellType"`
			UnpaidCount       int    `json:"unpaidCount"`
			UseDemotion       bool   `json:"useDemotion"`
		} `json:"chargeInfo"`
	} `json:"mixInfo"`
	SeriesInfo struct {
		ContentSubType    int    `json:"contentSubType"`
		IsSeries2Mix      bool   `json:"isSeries2Mix"`
		LightIconUrl      string `json:"lightIconUrl"`
		DarkIconUrl       string `json:"darkIconUrl"`
		Cover             string `json:"cover"`
		HorizontalCover   string `json:"horizontalCover"`
		SeriesId          string `json:"seriesId"`
		SeriesName        string `json:"seriesName"`
		Status            int    `json:"status"`
		PlayVV            int    `json:"playVV"`
		CollectVV         int    `json:"collectVV"`
		CurrentEpisode    int    `json:"currentEpisode"`
		TotalEpisode      int    `json:"totalEpisode"`
		HasUpdatedEpisode int    `json:"hasUpdatedEpisode"`
		IsCollected       int    `json:"isCollected"`
		WatchedItem       string `json:"watchedItem"`
		Author            struct {
			Uid         string `json:"uid"`
			Secret      int    `json:"secret"`
			IsBlockedV2 bool   `json:"isBlockedV2"`
			UserNotSee  int    `json:"userNotSee"`
			Nickname    string `json:"nickname"`
		} `json:"author"`
		Desc               string `json:"desc"`
		SeriesContentTypes []struct {
			SeriesContentType int    `json:"seriesContentType"`
			Name              string `json:"name"`
		} `json:"seriesContentTypes"`
		IsCharge    bool `json:"isCharge"`
		IsIaa       bool `json:"isIaa"`
		IsExclusive bool `json:"isExclusive"`
		Stats       struct {
			CurrentEpisode   int `json:"currentEpisode"`
			TotalEpisode     int `json:"totalEpisode"`
			UpdatedToEpisode int `json:"updatedToEpisode"`
		} `json:"stats"`
		ChargeInfo struct {
		} `json:"chargeInfo"`
	} `json:"seriesInfo"`
	IsPrivate     bool `json:"isPrivate"`
	IsFriendLimit bool `json:"isFriendLimit"`
	Download      struct {
		Url           string   `json:"url"`
		UrlList       []string `json:"urlList"`
		AllowDownload bool     `json:"allowDownload"`
		DataSize      int      `json:"dataSize"`
	} `json:"download"`
	ImpressionData string `json:"impressionData"`
	Stats          struct {
		CommentCount   int `json:"commentCount"`
		DiggCount      int `json:"diggCount"`
		ShareCount     int `json:"shareCount"`
		PlayCount      int `json:"playCount"`
		CollectCount   int `json:"collectCount"`
		DownloadCount  int `json:"downloadCount"`
		ForwardCount   int `json:"forwardCount"`
		LiveWatchCount int `json:"liveWatchCount"`
		RecommendCount int `json:"recommendCount"`
	} `json:"stats"`
	ShareInfo struct {
		ShareUrl      string `json:"shareUrl"`
		ShareLinkDesc string `json:"shareLinkDesc"`
	} `json:"shareInfo"`
	Status struct {
		AllowShare    bool `json:"allowShare"`
		IsReviewing   bool `json:"isReviewing"`
		IsDelete      bool `json:"isDelete"`
		IsProhibited  bool `json:"isProhibited"`
		PrivateStatus int  `json:"privateStatus"`
		ReviewStatus  int  `json:"reviewStatus"`
		PartSee       int  `json:"partSee"`
		VideoMute     struct {
			IsMute   bool   `json:"isMute"`
			MuteDesc string `json:"muteDesc"`
		} `json:"videoMute"`
		ReviewResult struct {
			ReviewStatus int `json:"reviewStatus"`
		} `json:"reviewResult"`
		AllowFriendRecommendGuide  bool `json:"allowFriendRecommendGuide"`
		AllowSelfRecommendToFriend bool `json:"allowSelfRecommendToFriend"`
	} `json:"status"`
	Music struct {
		Id         int64  `json:"id"`
		IdStr      string `json:"idStr"`
		Mid        string `json:"mid"`
		Author     string `json:"author"`
		Title      string `json:"title"`
		CoverThumb struct {
			UrlList []string `json:"urlList"`
			Uri     string   `json:"uri"`
			Width   int      `json:"width"`
			Height  int      `json:"height"`
		} `json:"coverThumb"`
		CoverMedium struct {
			UrlList []string `json:"urlList"`
			Uri     string   `json:"uri"`
			Width   int      `json:"width"`
			Height  int      `json:"height"`
		} `json:"coverMedium"`
		PlayUrl struct {
			UrlList []string `json:"urlList"`
			Uri     string   `json:"uri"`
			Width   int      `json:"width"`
			Height  int      `json:"height"`
			UrlKey  string   `json:"urlKey"`
		} `json:"playUrl"`
		SecUid      string `json:"secUid"`
		AvatarThumb struct {
			UrlList []string `json:"urlList"`
			Uri     string   `json:"uri"`
			Width   int      `json:"width"`
			Height  int      `json:"height"`
		} `json:"avatarThumb"`
		OwnerNickname     string `json:"ownerNickname"`
		CollectStat       int    `json:"collectStat"`
		BindedChallengeId int    `json:"bindedChallengeId"`
		Status            int    `json:"status"`
		CanNotPlay        bool   `json:"canNotPlay"`
		MusicName         string `json:"musicName"`
		IsOriginal        bool   `json:"isOriginal"`
		Duration          int    `json:"duration"`
		UserCount         int    `json:"userCount"`
	} `json:"music"`
	Images     []interface{} `json:"images"`
	ImageInfos string        `json:"imageInfos"`
	ImgBitrate []interface{} `json:"imgBitrate"`
	DyQ        string        `json:"dyQ"`
	Rate       int           `json:"rate"`
}
