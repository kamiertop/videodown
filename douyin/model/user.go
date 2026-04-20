package model

type UserData struct {
	User       User `json:"user"`
	StatusCode int  `json:"status_code"`
}

type User struct {
	AvatarLarger  Avatar `json:"avatar_larger"`
	AvatarMedium  Avatar `json:"avatar_medium"`
	AvatarThumb   Avatar `json:"avatar_thumb"`
	City          string `json:"city"`
	AwemeCount    int    `json:"aweme_count"` // 作品数量
	Country       string `json:"country"`
	FollowerCount int    `json:"follower_count"`
	IpLocation    string `json:"ip_location"`
	NickName      string `json:"nickname"`
	Province      string `json:"province"`  // 省份
	Signature     string `json:"signature"` // 签名
	Uid           string `json:"uid"`
	UniqueId      string `json:"unique_id"` // 抖音号
	SecUid        string `json:"sec_uid"`   // 安全uid，某些接口需要传这个
}

type UserVideoList struct {
	StatusCode int   `json:"status_code"`
	MinCursor  int   `json:"min_cursor"`
	MaxCursor  int64 `json:"max_cursor"`
	HasMore    int   `json:"has_more"`
	AwemeList  []struct {
		AwemeId    string `json:"aweme_id"`
		Desc       string `json:"desc"`
		CreateTime int    `json:"create_time"`
		Author     struct {
			Uid                   string      `json:"uid"`
			CardEntriesNotDisplay interface{} `json:"card_entries_not_display"`
			Nickname              string      `json:"nickname"`
			PersonalTagList       interface{} `json:"personal_tag_list"`
			CanSetGeofencing      interface{} `json:"can_set_geofencing"`
			AvatarThumb           struct {
				Uri     string   `json:"uri"`
				UrlList []string `json:"url_list"`
				Width   int      `json:"width"`
				Height  int      `json:"height"`
			} `json:"avatar_thumb"`
			CreatorTagList      interface{} `json:"creator_tag_list"`
			RiskNoticeText      string      `json:"risk_notice_text"`
			FollowStatus        int         `json:"follow_status"`
			DataLabelList       interface{} `json:"data_label_list"`
			Story25Comment      int         `json:"story25_comment"`
			NotSeenItemIdListV2 interface{} `json:"not_seen_item_id_list_v2"`
			CustomVerify        string      `json:"custom_verify"`
			BanUserFunctions    interface{} `json:"ban_user_functions"`
			CfList              interface{} `json:"cf_list"`
			InterestTags        interface{} `json:"interest_tags"`
			PrivateRelationList interface{} `json:"private_relation_list"`
			NeedPoints          interface{} `json:"need_points"`
			ShareInfo           struct {
				ShareUrl       string `json:"share_url"`
				ShareWeiboDesc string `json:"share_weibo_desc"`
				ShareDesc      string `json:"share_desc"`
				ShareTitle     string `json:"share_title"`
				ShareQrcodeUrl struct {
					Uri     string   `json:"uri"`
					UrlList []string `json:"url_list"`
					Width   int      `json:"width"`
					Height  int      `json:"height"`
				} `json:"share_qrcode_url"`
				ShareTitleMyself string `json:"share_title_myself"`
				ShareTitleOther  string `json:"share_title_other"`
				ShareDescInfo    string `json:"share_desc_info"`
			} `json:"share_info"`
			WhiteCoverUrl             interface{} `json:"white_cover_url"`
			EndorsementInfoList       interface{} `json:"endorsement_info_list"`
			HomepageBottomToast       interface{} `json:"homepage_bottom_toast"`
			EnterpriseVerifyReason    string      `json:"enterprise_verify_reason"`
			IsAdFake                  bool        `json:"is_ad_fake"`
			StoryInteractive          int         `json:"story_interactive"`
			UserPermissions           interface{} `json:"user_permissions"`
			DisplayInfo               interface{} `json:"display_info"`
			AccountCertInfo           string      `json:"account_cert_info"`
			OfflineInfoList           interface{} `json:"offline_info_list"`
			ContrailList              interface{} `json:"contrail_list"`
			CardSortPriority          interface{} `json:"card_sort_priority"`
			PreventDownload           bool        `json:"prevent_download"`
			FamiliarVisitorUser       interface{} `json:"familiar_visitor_user"`
			UserTags                  interface{} `json:"user_tags"`
			TextExtra                 interface{} `json:"text_extra"`
			FollowerStatus            int         `json:"follower_status"`
			VerificationPermissionIds interface{} `json:"verification_permission_ids"`
			AvatarSchemaList          interface{} `json:"avatar_schema_list"`
			ProfileMobParams          interface{} `json:"profile_mob_params"`
			CoverUrl                  []struct {
				Uri     string   `json:"uri"`
				UrlList []string `json:"url_list"`
				Width   int      `json:"width"`
				Height  int      `json:"height"`
			} `json:"cover_url"`
			ImRoleIds                              interface{} `json:"im_role_ids"`
			NotSeenItemIdList                      interface{} `json:"not_seen_item_id_list"`
			SpecialPeopleLabels                    interface{} `json:"special_people_labels"`
			SignatureExtra                         interface{} `json:"signature_extra"`
			FollowerListSecondaryInformationStruct interface{} `json:"follower_list_secondary_information_struct"`
			CardEntries                            interface{} `json:"card_entries"`
			StoryTtl                               int         `json:"story_ttl"`
			BatchUnfollowRelationDesc              interface{} `json:"batch_unfollow_relation_desc"`
			ProfileComponentDisabled               interface{} `json:"profile_component_disabled"`
			IdentityLabels                         interface{} `json:"identity_labels"`
			LinkItemList                           interface{} `json:"link_item_list"`
			BatchUnfollowContainTabs               interface{} `json:"batch_unfollow_contain_tabs"`
			SecUid                                 string      `json:"sec_uid"`
		} `json:"author"`
		Music struct {
			Id      int64  `json:"id"`
			IdStr   string `json:"id_str"`
			Title   string `json:"title"`
			Author  string `json:"author"`
			Album   string `json:"album"`
			CoverHd struct {
				Uri     string   `json:"uri"`
				UrlList []string `json:"url_list"`
				Width   int      `json:"width"`
				Height  int      `json:"height"`
			} `json:"cover_hd,omitempty"`
			CoverLarge struct {
				Uri     string   `json:"uri"`
				UrlList []string `json:"url_list"`
				Width   int      `json:"width"`
				Height  int      `json:"height"`
			} `json:"cover_large,omitempty"`
			CoverMedium struct {
				Uri     string   `json:"uri"`
				UrlList []string `json:"url_list"`
				Width   int      `json:"width"`
				Height  int      `json:"height"`
			} `json:"cover_medium,omitempty"`
			CoverThumb struct {
				Uri     string   `json:"uri"`
				UrlList []string `json:"url_list"`
				Width   int      `json:"width"`
				Height  int      `json:"height"`
			} `json:"cover_thumb"`
			PlayUrl struct {
				Uri     string   `json:"uri"`
				UrlList []string `json:"url_list"`
				Width   int      `json:"width"`
				Height  int      `json:"height"`
				UrlKey  string   `json:"url_key"`
			} `json:"play_url"`
			SchemaUrl                 string        `json:"schema_url"`
			SourcePlatform            int           `json:"source_platform"`
			StartTime                 int           `json:"start_time"`
			EndTime                   int           `json:"end_time"`
			Duration                  int           `json:"duration"`
			Extra                     string        `json:"extra"`
			UserCount                 int           `json:"user_count"`
			Position                  interface{}   `json:"position"`
			CollectStat               int           `json:"collect_stat"`
			Status                    int           `json:"status"`
			OfflineDesc               string        `json:"offline_desc"`
			OwnerId                   string        `json:"owner_id"`
			OwnerNickname             string        `json:"owner_nickname"`
			IsOriginal                bool          `json:"is_original"`
			Mid                       string        `json:"mid"`
			BindedChallengeId         int           `json:"binded_challenge_id"`
			Redirect                  bool          `json:"redirect"`
			IsRestricted              bool          `json:"is_restricted"`
			AuthorDeleted             bool          `json:"author_deleted"`
			IsDelVideo                bool          `json:"is_del_video"`
			IsVideoSelfSee            bool          `json:"is_video_self_see"`
			OwnerHandle               string        `json:"owner_handle"`
			AuthorPosition            interface{}   `json:"author_position"`
			PreventDownload           bool          `json:"prevent_download"`
			UnshelveCountries         interface{}   `json:"unshelve_countries"`
			PreventItemDownloadStatus int           `json:"prevent_item_download_status"`
			ExternalSongInfo          []interface{} `json:"external_song_info"`
			SecUid                    string        `json:"sec_uid"`
			AvatarThumb               struct {
				Uri     string   `json:"uri"`
				UrlList []string `json:"url_list"`
				Width   int      `json:"width"`
				Height  int      `json:"height"`
			} `json:"avatar_thumb"`
			AvatarMedium struct {
				Uri     string   `json:"uri"`
				UrlList []string `json:"url_list"`
				Width   int      `json:"width"`
				Height  int      `json:"height"`
			} `json:"avatar_medium"`
			AvatarLarge struct {
				Uri     string   `json:"uri"`
				UrlList []string `json:"url_list"`
				Width   int      `json:"width"`
				Height  int      `json:"height"`
			} `json:"avatar_large"`
			PreviewStartTime int  `json:"preview_start_time"`
			PreviewEndTime   int  `json:"preview_end_time"`
			IsCommerceMusic  bool `json:"is_commerce_music"`
			IsOriginalSound  bool `json:"is_original_sound"`
			AuditionDuration int  `json:"audition_duration"`
			ShootDuration    int  `json:"shoot_duration"`
			ReasonType       int  `json:"reason_type"`
			Artists          []struct {
				Uid      string `json:"uid"`
				SecUid   string `json:"sec_uid"`
				NickName string `json:"nick_name"`
				Handle   string `json:"handle"`
				Avatar   struct {
					Uri     string   `json:"uri"`
					UrlList []string `json:"url_list"`
				} `json:"avatar"`
				IsVerified bool `json:"is_verified"`
				EnterType  int  `json:"enter_type"`
			} `json:"artists"`
			LyricShortPosition   interface{} `json:"lyric_short_position"`
			MuteShare            bool        `json:"mute_share"`
			TagList              interface{} `json:"tag_list"`
			DmvAutoShow          bool        `json:"dmv_auto_show"`
			IsPgc                bool        `json:"is_pgc"`
			IsMatchedMetadata    bool        `json:"is_matched_metadata"`
			IsAudioUrlWithCookie bool        `json:"is_audio_url_with_cookie"`
			MusicChartRanks      interface{} `json:"music_chart_ranks"`
			CanBackgroundPlay    bool        `json:"can_background_play"`
			MusicStatus          int         `json:"music_status"`
			VideoDuration        int         `json:"video_duration"`
			PgcMusicType         int         `json:"pgc_music_type"`
			AuthorStatus         int         `json:"author_status"`
			SearchImpr           struct {
				EntityId string `json:"entity_id"`
			} `json:"search_impr"`
			ArtistUserInfos                interface{} `json:"artist_user_infos"`
			DspStatus                      int         `json:"dsp_status"`
			MusicianUserInfos              interface{} `json:"musician_user_infos"`
			MusicCollectCount              int         `json:"music_collect_count"`
			MusicCoverAtmosphereColorValue string      `json:"music_cover_atmosphere_color_value"`
			ShowOriginClip                 bool        `json:"show_origin_clip"`
			StrongBeatUrl                  struct {
				Uri     string   `json:"uri"`
				UrlList []string `json:"url_list"`
				Width   int      `json:"width"`
				Height  int      `json:"height"`
			} `json:"strong_beat_url,omitempty"`
			CoverColorHsv struct {
				H int `json:"h"`
				S int `json:"s"`
				V int `json:"v"`
			} `json:"cover_color_hsv,omitempty"`
			Song struct {
				Id      int64       `json:"id"`
				IdStr   string      `json:"id_str"`
				Title   string      `json:"title,omitempty"`
				Artists interface{} `json:"artists"`
				Chorus  struct {
					StartMs    int `json:"start_ms"`
					DurationMs int `json:"duration_ms"`
				} `json:"chorus,omitempty"`
				ChorusV3Infos interface{} `json:"chorus_v3_infos"`
			} `json:"song,omitempty"`
			MusicImageBeats struct {
				MusicImageBeatsUrl struct {
					Uri     string   `json:"uri"`
					UrlList []string `json:"url_list"`
					Width   int      `json:"width"`
					Height  int      `json:"height"`
				} `json:"music_image_beats_url"`
			} `json:"music_image_beats,omitempty"`
			MatchedPgcSound struct {
				Author      string `json:"author"`
				Title       string `json:"title"`
				MixedTitle  string `json:"mixed_title"`
				MixedAuthor string `json:"mixed_author"`
				CoverMedium struct {
					Uri     string   `json:"uri"`
					UrlList []string `json:"url_list"`
					Width   int      `json:"width"`
					Height  int      `json:"height"`
				} `json:"cover_medium"`
			} `json:"matched_pgc_sound,omitempty"`
		} `json:"music"`
		DiversionBarInfo []interface{} `json:"diversion_bar_info"`
		Video            struct {
			PlayAddr struct {
				Uri      string   `json:"uri"`
				UrlList  []string `json:"url_list"`
				Width    int      `json:"width"`
				Height   int      `json:"height"`
				UrlKey   string   `json:"url_key"`
				DataSize int      `json:"data_size,omitempty"`
				FileHash string   `json:"file_hash,omitempty"`
				FileCs   string   `json:"file_cs,omitempty"`
			} `json:"play_addr"`
			Cover struct {
				Uri     string   `json:"uri"`
				UrlList []string `json:"url_list"`
				Width   int      `json:"width"`
				Height  int      `json:"height"`
			} `json:"cover"`
			Height       int `json:"height"`
			Width        int `json:"width"`
			DynamicCover struct {
				Uri     string   `json:"uri"`
				UrlList []string `json:"url_list"`
				Width   int      `json:"width"`
				Height  int      `json:"height"`
			} `json:"dynamic_cover,omitempty"`
			OriginCover struct {
				Uri     string   `json:"uri"`
				UrlList []string `json:"url_list"`
				Width   int      `json:"width"`
				Height  int      `json:"height"`
			} `json:"origin_cover"`
			Ratio       string `json:"ratio"`
			Format      string `json:"format,omitempty"`
			Meta        string `json:"meta"`
			IsSourceHDR int    `json:"is_source_HDR,omitempty"`
			BitRate     []struct {
				GearName    string `json:"gear_name"`
				QualityType int    `json:"quality_type"`
				BitRate     int    `json:"bit_rate"`
				PlayAddr    struct {
					Uri      string   `json:"uri"`
					UrlList  []string `json:"url_list"`
					Width    int      `json:"width"`
					Height   int      `json:"height"`
					UrlKey   string   `json:"url_key"`
					DataSize int      `json:"data_size"`
					FileHash string   `json:"file_hash"`
					FileCs   string   `json:"file_cs,omitempty"`
				} `json:"play_addr"`
				IsH265     int    `json:"is_h265"`
				IsBytevc1  int    `json:"is_bytevc1"`
				HDRType    string `json:"HDR_type"`
				HDRBit     string `json:"HDR_bit"`
				FPS        int    `json:"FPS"`
				VideoExtra string `json:"video_extra"`
				Format     string `json:"format"`
			} `json:"bit_rate"`
			Duration     int `json:"duration"`
			BitRateAudio []struct {
				AudioMeta struct {
					UrlList struct {
						MainUrl     string `json:"main_url"`
						BackupUrl   string `json:"backup_url"`
						FallbackUrl string `json:"fallback_url"`
					} `json:"url_list"`
					EncodedType string `json:"encoded_type"`
					MediaType   string `json:"media_type"`
					LogoType    string `json:"logo_type"`
					Quality     string `json:"quality"`
					QualityDesc string `json:"quality_desc"`
					Format      string `json:"format"`
					Bitrate     int    `json:"bitrate"`
					CodecType   string `json:"codec_type"`
					Size        int    `json:"size"`
					Fps         int    `json:"fps"`
					FileId      string `json:"file_id"`
					FileHash    string `json:"file_hash"`
					SubInfo     string `json:"sub_info"`
				} `json:"audio_meta"`
				AudioQuality int    `json:"audio_quality"`
				AudioExtra   string `json:"audio_extra"`
			} `json:"bit_rate_audio"`
			GaussianCover struct {
				Uri     string   `json:"uri"`
				UrlList []string `json:"url_list"`
				Width   int      `json:"width"`
				Height  int      `json:"height"`
			} `json:"gaussian_cover,omitempty"`
			PlayAddr265 struct {
				Uri      string   `json:"uri"`
				UrlList  []string `json:"url_list"`
				Width    int      `json:"width"`
				Height   int      `json:"height"`
				UrlKey   string   `json:"url_key"`
				DataSize int      `json:"data_size"`
				FileHash string   `json:"file_hash"`
				FileCs   string   `json:"file_cs"`
			} `json:"play_addr_265,omitempty"`
			Audio struct {
				OriginalSoundInfos interface{} `json:"original_sound_infos"`
			} `json:"audio"`
			PlayAddrH264 struct {
				Uri      string   `json:"uri"`
				UrlList  []string `json:"url_list"`
				Width    int      `json:"width"`
				Height   int      `json:"height"`
				UrlKey   string   `json:"url_key"`
				DataSize int      `json:"data_size"`
				FileHash string   `json:"file_hash"`
				FileCs   string   `json:"file_cs"`
			} `json:"play_addr_h264,omitempty"`
			RawCover struct {
				Uri     string   `json:"uri"`
				UrlList []string `json:"url_list"`
				Width   int      `json:"width"`
				Height  int      `json:"height"`
			} `json:"raw_cover,omitempty"`
			AnimatedCover struct {
				Uri     string   `json:"uri"`
				UrlList []string `json:"url_list"`
			} `json:"animated_cover,omitempty"`
			HorizontalType    int    `json:"horizontal_type,omitempty"`
			MiscDownloadAddrs string `json:"misc_download_addrs,omitempty"`
			BigThumbs         []struct {
				ImgNum   int      `json:"img_num"`
				Uri      string   `json:"uri"`
				ImgUrl   string   `json:"img_url"`
				ImgXSize int      `json:"img_x_size"`
				ImgYSize int      `json:"img_y_size"`
				ImgXLen  int      `json:"img_x_len"`
				ImgYLen  int      `json:"img_y_len"`
				Duration float64  `json:"duration"`
				Interval float64  `json:"interval"`
				Fext     string   `json:"fext"`
				Uris     []string `json:"uris"`
				ImgUrls  []string `json:"img_urls"`
			} `json:"big_thumbs"`
			VideoModel     string `json:"video_model,omitempty"`
			UseStaticCover bool   `json:"use_static_cover,omitempty"`
			IsLongVideo    int    `json:"is_long_video,omitempty"`
		} `json:"video"`
		ShareUrl   string `json:"share_url"`
		UserDigged int    `json:"user_digged"`
		Statistics struct {
			RecommendCount int `json:"recommend_count"`
			CommentCount   int `json:"comment_count"`
			DiggCount      int `json:"digg_count"`
			AdmireCount    int `json:"admire_count"`
			PlayCount      int `json:"play_count"`
			ShareCount     int `json:"share_count"`
			CollectCount   int `json:"collect_count"`
		} `json:"statistics"`
		Status struct {
			NotAllowSoftDelReason string `json:"not_allow_soft_del_reason"`
			IsDelete              bool   `json:"is_delete"`
			AllowShare            bool   `json:"allow_share"`
			ReviewResult          struct {
				ReviewStatus int `json:"review_status"`
			} `json:"review_result"`
			AllowFriendRecommendGuide  bool `json:"allow_friend_recommend_guide"`
			PartSee                    int  `json:"part_see"`
			PrivateStatus              int  `json:"private_status"`
			ListenVideoStatus          int  `json:"listen_video_status"`
			InReviewing                bool `json:"in_reviewing"`
			AllowSelfRecommendToFriend bool `json:"allow_self_recommend_to_friend"`
			AllowFriendRecommend       bool `json:"allow_friend_recommend"`
			IsProhibited               bool `json:"is_prohibited"`
			EnableSoftDelete           int  `json:"enable_soft_delete"`
		} `json:"status"`
		PublishPlusAlienation struct {
			AlienationType int `json:"alienation_type"`
		} `json:"publish_plus_alienation"`
		TextExtra []struct {
			Start         int    `json:"start"`
			End           int    `json:"end"`
			Type          int    `json:"type"`
			CaptionStart  int    `json:"caption_start,omitempty"`
			CaptionEnd    int    `json:"caption_end,omitempty"`
			SearchText    string `json:"search_text,omitempty"`
			SearchQueryId string `json:"search_query_id,omitempty"`
			HashtagName   string `json:"hashtag_name,omitempty"`
			HashtagId     string `json:"hashtag_id,omitempty"`
			IsCommerce    bool   `json:"is_commerce,omitempty"`
			UserId        string `json:"user_id,omitempty"`
			SecUid        string `json:"sec_uid,omitempty"`
		} `json:"text_extra"`
		IsTop                  int `json:"is_top"`
		EntertainmentVideoType int `json:"entertainment_video_type"`
		ShareInfo              struct {
			ShareUrl      string `json:"share_url"`
			ShareLinkDesc string `json:"share_link_desc"`
		} `json:"share_info"`
		ProductGenreInfo struct {
			ProductGenreType        int   `json:"product_genre_type"`
			MaterialGenreSubTypeSet []int `json:"material_genre_sub_type_set"`
			SpecialInfo             struct {
				RecommendGroupName int `json:"recommend_group_name"`
			} `json:"special_info"`
		} `json:"product_genre_info"`
		VideoLabels          interface{} `json:"video_labels"`
		IsFromAdAuth         bool        `json:"is_from_ad_auth"`
		IsAds                bool        `json:"is_ads"`
		Duration             int         `json:"duration"`
		AwemeType            int         `json:"aweme_type"`
		Is25Story            int         `json:"is_25_story"`
		EnableDecoratedEmoji bool        `json:"enable_decorated_emoji"`
		ImageInfos           interface{} `json:"image_infos"`
		RiskInfos            struct {
			Vote     bool   `json:"vote"`
			Warn     bool   `json:"warn"`
			RiskSink bool   `json:"risk_sink"`
			Type     int    `json:"type"`
			Content  string `json:"content"`
		} `json:"risk_infos"`
		SeriesBasicInfo struct {
		} `json:"series_basic_info"`
		MvInfo                          interface{}   `json:"mv_info"`
		Position                        interface{}   `json:"position"`
		UniqidPosition                  interface{}   `json:"uniqid_position"`
		CommentList                     interface{}   `json:"comment_list"`
		AuthorUserId                    int64         `json:"author_user_id"`
		IsSubtitled                     int           `json:"is_subtitled"`
		Geofencing                      []interface{} `json:"geofencing"`
		NearbyHotComment                interface{}   `json:"nearby_hot_comment"`
		SelectAnchorExpandedContent     int           `json:"select_anchor_expanded_content"`
		Region                          string        `json:"region"`
		VideoText                       interface{}   `json:"video_text"`
		AwemeTypeTags                   string        `json:"aweme_type_tags"`
		CollectStat                     int           `json:"collect_stat"`
		LabelTopText                    interface{}   `json:"label_top_text"`
		Promotions                      []interface{} `json:"promotions"`
		GroupId                         string        `json:"group_id"`
		PreventDownload                 bool          `json:"prevent_download"`
		NicknamePosition                interface{}   `json:"nickname_position"`
		ChallengePosition               interface{}   `json:"challenge_position"`
		CfAssetsType                    int           `json:"cf_assets_type"`
		PersonalPageBottonDiagnoseStyle int           `json:"personal_page_botton_diagnose_style"`
		FlashMobTrends                  int           `json:"flash_mob_trends"`
		LongVideo                       interface{}   `json:"long_video"`
		IsMomentHistory                 int           `json:"is_moment_history"`
		IsMomentStory                   int           `json:"is_moment_story"`
		EntLogExtra                     struct {
			LogExtra string `json:"log_extra"`
		} `json:"ent_log_extra"`
		SecItemId           string      `json:"sec_item_id"`
		InteractionStickers interface{} `json:"interaction_stickers"`
		ShootWay            string      `json:"shoot_way"`
		OriginCommentIds    interface{} `json:"origin_comment_ids"`
		CommerceConfigData  interface{} `json:"commerce_config_data"`
		EffectInflowEffects interface{} `json:"effect_inflow_effects"`
		VideoControl        struct {
			AllowDownload            bool `json:"allow_download"`
			ShareType                int  `json:"share_type"`
			ShowProgressBar          int  `json:"show_progress_bar"`
			DraftProgressBar         int  `json:"draft_progress_bar"`
			AllowDuet                bool `json:"allow_duet"`
			AllowReact               bool `json:"allow_react"`
			PreventDownloadType      int  `json:"prevent_download_type"`
			AllowDynamicWallpaper    bool `json:"allow_dynamic_wallpaper"`
			TimerStatus              int  `json:"timer_status"`
			AllowMusic               bool `json:"allow_music"`
			AllowStitch              bool `json:"allow_stitch"`
			AllowDouplus             bool `json:"allow_douplus"`
			AllowShare               bool `json:"allow_share"`
			ShareGrayed              bool `json:"share_grayed"`
			DownloadIgnoreVisibility bool `json:"download_ignore_visibility"`
			DuetIgnoreVisibility     bool `json:"duet_ignore_visibility"`
			ShareIgnoreVisibility    bool `json:"share_ignore_visibility"`
			DownloadInfo             struct {
				Level    int `json:"level"`
				FailInfo struct {
					Code   int    `json:"code"`
					Reason string `json:"reason"`
					Msg    string `json:"msg"`
				} `json:"fail_info,omitempty"`
			} `json:"download_info"`
			DuetInfo struct {
				Level    int `json:"level"`
				FailInfo struct {
					Code   int    `json:"code"`
					Reason string `json:"reason"`
				} `json:"fail_info,omitempty"`
			} `json:"duet_info"`
			AllowRecord         bool   `json:"allow_record"`
			DisableRecordReason string `json:"disable_record_reason"`
			TimerInfo           struct {
				PublicTime  int `json:"public_time,omitempty"`
				TimerStatus int `json:"timer_status,omitempty"`
			} `json:"timer_info"`
		} `json:"video_control"`
		AwemeControl struct {
			CanForward     bool `json:"can_forward"`
			CanShare       bool `json:"can_share"`
			CanComment     bool `json:"can_comment"`
			CanShowComment bool `json:"can_show_comment"`
		} `json:"aweme_control"`
		FollowShotAssets   interface{} `json:"follow_shot_assets"`
		ItemAigcFollowShot int         `json:"item_aigc_follow_shot"`
		AwemeListenStruct  struct {
			TraceInfo string `json:"trace_info"`
		} `json:"aweme_listen_struct"`
		IsNewTextMode     int         `json:"is_new_text_mode"`
		Anchors           interface{} `json:"anchors"`
		HybridLabel       interface{} `json:"hybrid_label"`
		GeofencingRegions interface{} `json:"geofencing_regions"`
		AiFollowImages    interface{} `json:"ai_follow_images"`
		GameTagInfo       struct {
			IsGame bool `json:"is_game"`
		} `json:"game_tag_info"`
		IsStory                   int `json:"is_story"`
		EntertainmentVideoPaidWay struct {
			PaidWays            []interface{} `json:"paid_ways"`
			PaidType            int           `json:"paid_type"`
			EnableUseNewEntData bool          `json:"enable_use_new_ent_data"`
		} `json:"entertainment_video_paid_way"`
		CanCacheToLocal         bool        `json:"can_cache_to_local"`
		PackUsageSceneByReqPath string      `json:"pack_usage_scene_by_req_path"`
		CoverLabels             interface{} `json:"cover_labels"`
		FollowShootClipInfo     struct {
			ClipVideoAll     int64 `json:"clip_video_all"`
			ClipFromUser     int64 `json:"clip_from_user,omitempty"`
			ClipFromPlatform int64 `json:"clip_from_platform,omitempty"`
			OriginClipId     int64 `json:"origin_clip_id,omitempty"`
		} `json:"follow_shoot_clip_info"`
		TrendsEventTrack string `json:"trends_event_track"`
		GuideBtnType     int    `json:"guide_btn_type"`
		DouplusUserType  int    `json:"douplus_user_type"`
		ComponentControl struct {
			DataSourceUrl string `json:"data_source_url"`
		} `json:"component_control"`
		Images []struct {
			Uri                          string      `json:"uri"`
			UrlList                      []string    `json:"url_list"`
			DownloadUrlList              []string    `json:"download_url_list"`
			Height                       int         `json:"height"`
			Width                        int         `json:"width"`
			MaskUrlList                  interface{} `json:"mask_url_list"`
			InteractionStickers          interface{} `json:"interaction_stickers"`
			IsNewTextMode                int         `json:"is_new_text_mode"`
			WatermarkFreeDownloadUrlList interface{} `json:"watermark_free_download_url_list"`
		} `json:"images"`
		RelationLabels interface{} `json:"relation_labels"`
		HorizontalType int         `json:"horizontal_type,omitempty"`
		InterestPoints interface{} `json:"interest_points"`
		ImpressionData struct {
			GroupIdListA   []int64     `json:"group_id_list_a"`
			GroupIdListB   []int64     `json:"group_id_list_b"`
			SimilarIdListA interface{} `json:"similar_id_list_a"`
			SimilarIdListB interface{} `json:"similar_id_list_b"`
			GroupIdListC   []int64     `json:"group_id_list_c"`
			GroupIdListD   []int64     `json:"group_id_list_d"`
		} `json:"impression_data"`
		EntertainmentRecommendInfo string      `json:"entertainment_recommend_info"`
		OriginDuetResourceUri      string      `json:"origin_duet_resource_uri"`
		LibfinsertTaskId           string      `json:"libfinsert_task_id"`
		SocialTagList              interface{} `json:"social_tag_list"`
		SuggestWords               struct {
			SuggestWords []struct {
				Words []struct {
					Word   string `json:"word"`
					WordId string `json:"word_id"`
					Info   string `json:"info"`
				} `json:"words"`
				Scene     string `json:"scene"`
				IconUrl   string `json:"icon_url"`
				HintText  string `json:"hint_text"`
				ExtraInfo string `json:"extra_info"`
			} `json:"suggest_words"`
		} `json:"suggest_words"`
		ShowFollowButton struct {
		} `json:"show_follow_button"`
		DuetAggregateInMusicTab bool `json:"duet_aggregate_in_music_tab"`
		IsDuetSing              bool `json:"is_duet_sing"`
		CommentPermissionInfo   struct {
			CommentPermissionStatus int  `json:"comment_permission_status"`
			CanComment              bool `json:"can_comment"`
			ItemDetailEntry         bool `json:"item_detail_entry"`
			PressEntry              bool `json:"press_entry"`
			ToastGuide              bool `json:"toast_guide"`
		} `json:"comment_permission_info"`
		OriginalImages interface{} `json:"original_images"`
		SeriesPaidInfo struct {
			SeriesPaidStatus int `json:"series_paid_status"`
			ItemPrice        int `json:"item_price"`
		} `json:"series_paid_info"`
		CategoryDa          int           `json:"category_da,omitempty"`
		ImgBitrate          []interface{} `json:"img_bitrate"`
		CommentGid          int64         `json:"comment_gid"`
		ImageAlbumMusicInfo struct {
			BeginTime int `json:"begin_time"`
			EndTime   int `json:"end_time"`
			Volume    int `json:"volume"`
		} `json:"image_album_music_info"`
		VideoTag []struct {
			TagId   int    `json:"tag_id"`
			TagName string `json:"tag_name"`
			Level   int    `json:"level"`
		} `json:"video_tag"`
		IsCollectsSelected int         `json:"is_collects_selected"`
		ChapterList        interface{} `json:"chapter_list"`
		FeedCommentConfig  struct {
			InputConfigText         string `json:"input_config_text"`
			AuthorAuditStatus       int    `json:"author_audit_status"`
			CommonFlags             string `json:"common_flags"`
			AudioCommentPermission  int    `json:"audio_comment_permission"`
			CommonCommentPermission int    `json:"common_comment_permission"`
		} `json:"feed_comment_config"`
		IsImageBeat          bool        `json:"is_image_beat"`
		DislikeDimensionList interface{} `json:"dislike_dimension_list"`
		StandardBarInfoList  interface{} `json:"standard_bar_info_list"`
		PhotoSearchEntrance  struct {
			EcomType int `json:"ecom_type"`
		} `json:"photo_search_entrance"`
		DanmakuControl struct {
			EnableDanmaku      bool   `json:"enable_danmaku"`
			PostPrivilegeLevel int    `json:"post_privilege_level"`
			IsPostDenied       bool   `json:"is_post_denied"`
			PostDeniedReason   string `json:"post_denied_reason"`
			SkipDanmaku        bool   `json:"skip_danmaku"`
			DanmakuCnt         int    `json:"danmaku_cnt"`
			Activities         []struct {
				Id   int `json:"id"`
				Type int `json:"type"`
			} `json:"activities"`
			PassThroughParams  string `json:"pass_through_params"`
			SmartModeDecision  int    `json:"smart_mode_decision"`
			FirstDanmakuOffset int    `json:"first_danmaku_offset"`
			LastDanmakuOffset  int    `json:"last_danmaku_offset"`
		} `json:"danmaku_control,omitempty"`
		IsLifeItem           bool        `json:"is_life_item"`
		ImageList            interface{} `json:"image_list"`
		ComponentInfoV2      string      `json:"component_info_v2"`
		ItemWarnNotification struct {
			Type    int    `json:"type"`
			Show    bool   `json:"show"`
			Content string `json:"content"`
		} `json:"item_warn_notification"`
		OriginTextExtra    interface{} `json:"origin_text_extra"`
		DisableRelationBar int         `json:"disable_relation_bar"`
		PackedClips        interface{} `json:"packed_clips"`
		VtagSearch         struct {
			VtagEnable  bool `json:"vtag_enable"`
			VtagDelayTs int  `json:"vtag_delay_ts"`
		} `json:"vtag_search,omitempty"`
		AuthorMaskTag        int  `json:"author_mask_tag"`
		UserRecommendStatus  int  `json:"user_recommend_status"`
		CollectionCornerMark int  `json:"collection_corner_mark"`
		IsSharePost          bool `json:"is_share_post"`
		ImageComment         struct {
			CommentHighlightText string `json:"comment_highlight_text"`
		} `json:"image_comment"`
		VisualSearchInfo struct {
			VisualSearchLongpress int  `json:"visual_search_longpress,omitempty"`
			IsShowImgEntrance     bool `json:"is_show_img_entrance"`
			IsEcomImg             bool `json:"is_ecom_img"`
			IsHighAccuracyEcom    bool `json:"is_high_accuracy_ecom"`
			IsHighRecallEcom      bool `json:"is_high_recall_ecom"`
		} `json:"visual_search_info"`
		TtsIdList                  interface{} `json:"tts_id_list"`
		RefTtsIdList               interface{} `json:"ref_tts_id_list"`
		VoiceModifyIdList          interface{} `json:"voice_modify_id_list"`
		RefVoiceModifyIdList       interface{} `json:"ref_voice_modify_id_list"`
		AuthenticationToken        string      `json:"authentication_token"`
		VideoGameDataChannelConfig struct {
		} `json:"video_game_data_channel_config"`
		DislikeDimensionListV2 interface{} `json:"dislike_dimension_list_v2"`
		DistributeCircle       struct {
			DistributeType         int  `json:"distribute_type"`
			CampusBlockInteraction bool `json:"campus_block_interaction"`
			IsCampus               bool `json:"is_campus"`
		} `json:"distribute_circle"`
		ImageCropCtrl    int         `json:"image_crop_ctrl"`
		YummeRecreason   interface{} `json:"yumme_recreason"`
		SlidesMusicBeats interface{} `json:"slides_music_beats"`
		JumpTabInfoList  interface{} `json:"jump_tab_info_list"`
		MediaType        int         `json:"media_type"`
		PlayProgress     struct {
			PlayProgress     int `json:"play_progress"`
			LastModifiedTime int `json:"last_modified_time"`
		} `json:"play_progress"`
		ReplySmartEmojis         interface{} `json:"reply_smart_emojis"`
		ActivityVideoType        int         `json:"activity_video_type"`
		BoostStatus              int         `json:"boost_status"`
		CreateScaleType          interface{} `json:"create_scale_type"`
		EntertainmentProductInfo struct {
			SubTitle   interface{} `json:"sub_title"`
			MarketInfo struct {
				LimitFree struct {
					InFree bool `json:"in_free"`
				} `json:"limit_free"`
				MarketingTag interface{} `json:"marketing_tag"`
			} `json:"market_info"`
		} `json:"entertainment_product_info"`
		Caption       string `json:"caption"`
		ItemTitle     string `json:"item_title"`
		IsUseMusic    bool   `json:"is_use_music"`
		Original      int    `json:"original"`
		XiguaBaseInfo struct {
			Status           int `json:"status"`
			StarAltarOrderId int `json:"star_altar_order_id"`
			StarAltarType    int `json:"star_altar_type"`
			ItemId           int `json:"item_id"`
		} `json:"xigua_base_info"`
		MarkLargelyFollowing bool `json:"mark_largely_following"`
		FriendRecommendInfo  struct {
			FriendRecommendSource            int         `json:"friend_recommend_source"`
			LabelUserList                    interface{} `json:"label_user_list"`
			DisableFriendRecommendGuideLabel bool        `json:"disable_friend_recommend_guide_label"`
		} `json:"friend_recommend_info"`
		EnableCommentStickerRec bool `json:"enable_comment_sticker_rec"`
		VideoShareEditStatus    int  `json:"video_share_edit_status"`
		Is24Story               int  `json:"is_24_story"`
		TrendsInfos             []struct {
			TrendsMaterials         interface{} `json:"trends_materials"`
			TrendsMusicInfo         interface{} `json:"trends_music_info"`
			TrendsUnifiedMusicGroup interface{} `json:"trends_unified_music_group"`
			InsertMusicIds          interface{} `json:"insert_music_ids"`
			TrackPassThrough        string      `json:"track_pass_through"`
		} `json:"trends_infos"`
		ChapterBarColor interface{} `json:"chapter_bar_color"`
		DuetExtraInfo   struct {
			DuetReason int `json:"duet_reason"`
		} `json:"duet_extra_info,omitempty"`
		LunaVideoCandidateStatus string `json:"luna_video_candidate_status,omitempty"`
		RelatedMusicAnchor       struct {
			Type      string `json:"type"`
			SchemaUrl string `json:"schema_url"`
			ImageUrl  struct {
				Uri     string   `json:"uri"`
				UrlList []string `json:"url_list"`
			} `json:"image_url"`
			Priority int    `json:"priority"`
			Extra    string `json:"extra"`
		} `json:"related_music_anchor,omitempty"`
		GalileoPadTextcrop struct {
			IpadDHCutRatio    []int `json:"ipad_d_h_cut_ratio"`
			IpadDVCutRatio    []int `json:"ipad_d_v_cut_ratio"`
			AndroidDHCutRatio []int `json:"android_d_h_cut_ratio"`
			AndroidDVCutRatio []int `json:"android_d_v_cut_ratio"`
			Version           int   `json:"version"`
		} `json:"galileo_pad_textcrop,omitempty"`
		IsMultiContent        int    `json:"is_multi_content,omitempty"`
		ImageItemQualityLevel int    `json:"image_item_quality_level,omitempty"`
		MainArchCommon        string `json:"main_arch_common,omitempty"`
		RecommendChapterInfo  struct {
			RecommendChapterList []struct {
				Desc      string `json:"desc"`
				Timestamp int    `json:"timestamp"`
				Detail    string `json:"detail"`
				Points    []struct {
					Desc   string `json:"desc"`
					Detail string `json:"detail"`
					Type   int    `json:"type"`
				} `json:"points"`
			} `json:"recommend_chapter_list"`
			ChapterRecommendType   int           `json:"chapter_recommend_type"`
			ChapterAbstract        string        `json:"chapter_abstract"`
			ChapterRecommendSource int           `json:"chapter_recommend_source"`
			PushScene              []interface{} `json:"push_scene"`
			ChapterBarColor        interface{}   `json:"chapter_bar_color"`
		} `json:"recommend_chapter_info,omitempty"`
		FollowMaterialInfo string `json:"follow_material_info,omitempty"`
		AigcInfo           struct {
			IsStickerAigc        int    `json:"is_sticker_aigc"`
			OriginalStickerId    string `json:"original_sticker_id"`
			AigcStickerId        string `json:"aigc_sticker_id"`
			AigcType             int    `json:"aigc_type"`
			Source               string `json:"source"`
			Extra                string `json:"extra"`
			AigcCocreationStruct struct {
				CocreationUserList []interface{} `json:"cocreation_user_list"`
			} `json:"aigc_cocreation_struct"`
		} `json:"aigc_info,omitempty"`
	} `json:"aweme_list"`
	TimeList []string `json:"time_list"`
	LogPb    struct {
		ImprId string `json:"impr_id"`
	} `json:"log_pb"`
	RequestItemCursor  int `json:"request_item_cursor"`
	PostSerial         int `json:"post_serial"`
	ReplaceSeriesCover int `json:"replace_series_cover"`
}

type AwemeItem struct {
}
