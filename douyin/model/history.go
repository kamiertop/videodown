package model

// HistoryResponse 观看历史接口的响应
type HistoryResponse struct {
	HasMore    uint8         `json:"has_more"`    // 是否有更多数据，1表示有，0表示没有
	MaxCursor  int64         `json:"max_cursor"`  // 最大游标，下一页请求时传入
	StatusCode int           `json:"status_code"` // 状态码，成功时为0
	StatusMsg  string        `json:"status_msg"`  // 状态信息，成功时为""
	AwemeList  []HistoryItem `json:"aweme_list"`  // 观看历史列表
}
type HistoryItem struct {
	Author                          Author          `json:"author"`
	AuthorMaskTag                   int             `json:"author_mask_tag"`
	AuthorUserId                    int64           `json:"author_user_id"`
	AwemeId                         string          `json:"aweme_id"`
	AwemeType                       int             `json:"aweme_type"`
	AwemeTypeTags                   string          `json:"aweme_type_tags"`
	Caption                         string          `json:"caption"`
	CreateTime                      int             `json:"create_time"`
	Desc                            string          `json:"desc"`
	GroupId                         string          `json:"group_id"`
	HaveDashboard                   bool            `json:"have_dashboard"`
	ImageCropCtrl                   int             `json:"image_crop_ctrl"`
	Is24Story                       int             `json:"is_24_story"`
	Is25Story                       int             `json:"is_25_story"`
	IsCollectsSelected              int             `json:"is_collects_selected"`
	IsDuetSing                      bool            `json:"is_duet_sing"`
	IsFirstVideo                    bool            `json:"is_first_video"`
	IsImageBeat                     bool            `json:"is_image_beat"`
	IsKaraoke                       bool            `json:"is_karaoke"`
	IsLifeItem                      bool            `json:"is_life_item"`
	IsMomentHistory                 int             `json:"is_moment_history"`
	IsMomentStory                   int             `json:"is_moment_story"`
	IsNewTextMode                   int             `json:"is_new_text_mode"`
	IsPreview                       int             `json:"is_preview"`
	IsSharePost                     bool            `json:"is_share_post"`
	IsStory                         int             `json:"is_story"`
	IsTop                           int             `json:"is_top"`
	IsUseMusic                      bool            `json:"is_use_music"`
	ItemAigcFollowShot              int             `json:"item_aigc_follow_shot"`
	ItemStitch                      int             `json:"item_stitch"`
	ItemTitle                       string          `json:"item_title"`
	MarkLargelyFollowing            bool            `json:"mark_largely_following"`
	MediaType                       int             `json:"media_type"`
	OriginDuetResourceUri           string          `json:"origin_duet_resource_uri"`
	Original                        int             `json:"original"`
	PackUsageSceneByReqPath         string          `json:"pack_usage_scene_by_req_path"`
	PersonalPageBottonDiagnoseStyle int             `json:"personal_page_botton_diagnose_style"`
	PreventDownload                 bool            `json:"prevent_download"`
	PreviewTitle                    string          `json:"preview_title"`
	PreviewVideoStatus              int             `json:"preview_video_status"`
	Region                          string          `json:"region"`
	ReportAction                    bool            `json:"report_action"`
	SelectAnchorExpandedContent     int             `json:"select_anchor_expanded_content"`
	Statistics                      VideoStatistics `json:"statistics"` // 视频统计信息
	Video                           Video           `json:"video"`
	VideoShareEditStatus            int             `json:"video_share_edit_status"`
}

// SearchHistoryResponse 搜索观看历史接口的响应
type SearchHistoryResponse struct {
	StatusCode int           `json:"status_code"`
	Cursor     int           `json:"cursor"`
	HasMore    int           `json:"has_more"`
	AwemeList  []HistoryItem `json:"aweme_list"`
}
