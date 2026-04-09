package model

// VideoURLData 视频播放地址数据（来自 /x/player/playurl 接口）
type VideoURLData struct {
	From              string          `json:"from"`                // 请求来源，通常为 "local"
	Result            string          `json:"result"`              // 结果标识，通常为 "suee"
	Message           string          `json:"message"`             // 错误信息，成功时为 "0"
	Quality           int             `json:"quality"`             // 当前视频清晰度代码（qn值）
	Format            string          `json:"format"`              // 视频格式，如 "flv", "mp4", "dash"
	Timelength        int             `json:"timelength"`          // 视频总时长，单位毫秒
	AcceptFormat      string          `json:"accept_format"`       // 支持的格式列表
	AcceptDescription []string        `json:"accept_description"`  // 支持的清晰度描述列表
	AcceptQuality     []int           `json:"accept_quality"`      // 支持的清晰度代码列表
	VideoCodecid      int             `json:"video_codecid"`       // 视频编码 ID：7=AVC, 12=HEVC, 13=AV1
	SeekParam         string          `json:"seek_param"`          // 拖拽参数类型
	SeekType          string          `json:"seek_type"`           // 拖拽类型
	Dash              Dash            `json:"dash"`                // DASH 流信息（音视频分离）
	SupportFormats    []SupportFormat `json:"support_formats"`     // 支持的格式详细信息
	HighFormat        any             `json:"high_format"`         // 高码率格式信息（通常为 null）
	Volume            Volume          `json:"volume"`              // 音量响度标准化信息
	LastPlayTime      int             `json:"last_play_time"`      // 上次播放位置，单位秒
	LastPlayCid       int64           `json:"last_play_cid"`       // 上次播放的分P CID
	ViewInfo          any             `json:"view_info"`           // 观看信息（通常为 null）
	PlayConf          PlayConf        `json:"play_conf"`           // 播放配置
	CurLanguage       string          `json:"cur_language"`        // 当前语言
	CurProductionType int             `json:"cur_production_type"` // 当前制作类型
	AutoQnResp        AutoQnResp      `json:"auto_qn_resp"`        // 自动清晰度响应
}

// AutoQnResp 自动清晰度响应
type AutoQnResp struct {
	Dyeid string `json:"dyeid"` // 染色ID，用于追踪
}

// PlayConf 播放配置
type PlayConf struct {
	IsNewDescription bool `json:"is_new_description"` // 是否使用新的描述格式
}

// Volume 音量响度标准化信息（用于统一不同视频的音量）
type Volume struct {
	MeasuredI         float64              `json:"measured_i"`         // 测量响度（Integrated Loudness）
	MeasuredLra       float64              `json:"measured_lra"`       // 测量响度范围（Loudness Range）
	MeasuredTp        float64              `json:"measured_tp"`        // 测量真峰值（True Peak）
	MeasuredThreshold float64              `json:"measured_threshold"` // 测量阈值
	TargetOffset      float64              `json:"target_offset"`      // 目标偏移量
	TargetI           int                  `json:"target_i"`           // 目标响度
	TargetTp          int                  `json:"target_tp"`          // 目标真峰值
	MultiSceneArgs    VolumeMultiSceneArgs `json:"multi_scene_args"`   // 多场景参数
}

// VolumeMultiSceneArgs 音量多场景参数
type VolumeMultiSceneArgs struct {
	HighDynamicTargetI string `json:"high_dynamic_target_i"` // 高动态范围目标响度
	NormalTargetI      string `json:"normal_target_i"`       // 普通目标响度
	UndersizedTargetI  string `json:"undersized_target_i"`   // 小尺寸目标响度
}

// SupportFormat 支持的格式信息
type SupportFormat struct {
	Quality          int            `json:"quality"`             // 清晰度代码
	Format           string         `json:"format"`              // 格式名称
	NewDescription   string         `json:"new_description"`     // 新描述
	DisplayDesc      string         `json:"display_desc"`        // 显示描述（如 "1080P 高清"）
	Superscript      string         `json:"superscript"`         // 上标标签（如 "VIP", "HDR"）
	Codecs           []string       `json:"codecs"`              // 编码格式列表
	CanWatchQnReason int            `json:"can_watch_qn_reason"` // 可观看原因代码
	LimitWatchReason int            `json:"limit_watch_reason"`  // 限制观看原因代码
	Report           map[string]any `json:"report"`              // 上报信息
}

// Dash 流信息（Dynamic Adaptive Streaming over HTTP，音视频分离）
type Dash struct {
	Duration       int         `json:"duration"`        // 视频总时长，单位秒
	MinBufferTime  float64     `json:"minBufferTime"`   // 最小缓冲时间（驼峰命名）
	MinBufferTime1 float64     `json:"min_buffer_time"` // 最小缓冲时间（下划线命名）
	Video          []VideoItem `json:"video"`           // 视频流列表（不同清晰度）
	Audio          []AudioItem `json:"audio"`           // 音频流列表（不同音质）
	Dolby          Dolby       `json:"dolby"`           // 杜比音效信息
	Flac           *Flac       `json:"flac"`            // Hi-Res 无损音轨信息（为 null 时表示不支持）
}

// VideoItem DASH 视频流项
type VideoItem struct {
	Id            int          `json:"id"`             // 视频流 ID（与清晰度对应）
	BaseUrl       string       `json:"baseUrl"`        // 主播放地址（驼峰命名）
	BaseUrl1      string       `json:"base_url"`       // 主播放地址（下划线命名）
	BackupUrl     []string     `json:"backupUrl"`      // 备用播放地址列表（驼峰命名）
	BackupUrl1    []string     `json:"backup_url"`     // 备用播放地址列表（下划线命名）
	Bandwidth     int          `json:"bandwidth"`      // 码率，单位 bps
	MimeType      string       `json:"mimeType"`       // MIME 类型（驼峰命名）
	MimeType1     string       `json:"mime_type"`      // MIME 类型（下划线命名）
	Codecs        string       `json:"codecs"`         // 编码格式（如 "avc1.640028"）
	Width         int          `json:"width"`          // 视频宽度
	Height        int          `json:"height"`         // 视频高度
	FrameRate     string       `json:"frameRate"`      // 帧率（驼峰命名）
	FrameRate1    string       `json:"frame_rate"`     // 帧率（下划线命名）
	Sar           string       `json:"sar"`            // 像素宽高比（Sample Aspect Ratio）
	StartWithSap  int          `json:"startWithSap"`   // SAP 类型（驼峰命名）
	StartWithSap1 int          `json:"start_with_sap"` // SAP 类型（下划线命名）
	SegmentBase   SegmentBase  `json:"SegmentBase"`    // 分段信息（驼峰命名）
	SegmentBase1  SegmentBase1 `json:"segment_base"`   // 分段信息（下划线命名）
	Codecid       int          `json:"codecid"`        // 编码器 ID：7=AVC, 12=HEVC, 13=AV1
}

// SegmentBase1 分段基础信息（下划线命名）
type SegmentBase1 struct {
	Initialization string `json:"initialization"` // 初始化段范围
	IndexRange     string `json:"index_range"`    // 索引段范围
}

// SegmentBase 分段基础信息（驼峰命名）
type SegmentBase struct {
	Initialization string `json:"Initialization"` // 初始化段范围
	IndexRange     string `json:"indexRange"`     // 索引段范围
}

// Flac Hi-Res 无损音频信息
type Flac struct {
	Display bool      `json:"display"` // 是否在播放器显示切换 Hi-Res 无损音轨按钮
	Audio   AudioItem `json:"audio"`   // 无损音频流信息
}

// Dolby 杜比音效信息
type Dolby struct {
	Type  int         `json:"type"`  // 杜比类型：1=杜比全景声, 2=杜比视界
	Audio interface{} `json:"audio"` // 杜比音频流列表（通常为数组或 null）
}

// AudioItem DASH 音频流项
type AudioItem struct {
	Id            int          `json:"id"`             // 音频流 ID（与音质对应）
	BaseUrl       string       `json:"baseUrl"`        // 主播放地址（驼峰命名）
	BaseUrl1      string       `json:"base_url"`       // 主播放地址（下划线命名）
	BackupUrl     []string     `json:"backupUrl"`      // 备用播放地址列表（驼峰命名）
	BackupUrl1    []string     `json:"backup_url"`     // 备用播放地址列表（下划线命名）
	Bandwidth     int          `json:"bandwidth"`      // 码率，单位 bps
	MimeType      string       `json:"mimeType"`       // MIME 类型（驼峰命名）
	MimeType1     string       `json:"mime_type"`      // MIME 类型（下划线命名）
	Codecs        string       `json:"codecs"`         // 编码格式（如 "mp4a.40.2"）
	Width         int          `json:"width"`          // 宽度（音频通常为 0）
	Height        int          `json:"height"`         // 高度（音频通常为 0）
	FrameRate     string       `json:"frameRate"`      // 帧率（音频通常为空）
	FrameRate1    string       `json:"frame_rate"`     // 帧率（音频通常为空）
	Sar           string       `json:"sar"`            // 像素宽高比（音频通常为空）
	StartWithSap  int          `json:"startWithSap"`   // SAP 类型（驼峰命名）
	StartWithSap1 int          `json:"start_with_sap"` // SAP 类型（下划线命名）
	SegmentBase   SegmentBase  `json:"SegmentBase"`    // 分段信息（驼峰命名）
	SegmentBase1  SegmentBase1 `json:"segment_base"`   // 分段信息（下划线命名）
	Codecid       int          `json:"codecid"`        // 音频编码器 ID
}
