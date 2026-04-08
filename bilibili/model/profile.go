package model

// MyInfoData 包含了用户的基本信息和一些统计数据
type MyInfoData struct {
	Coins float32 `json:"coins"`
	//Following uint32        `json:"following"` //数值不准
	Follower uint32         `json:"follower"` //可能是准的
	Profile  MyInfoProfile  `json:"profile"`
	LevelExp MyInfoLevelExp `json:"level_exp"`
}

type MyInfoLevelExp struct {
	CurrentExp   uint32 `json:"current_exp"`
	CurrentLevel uint8  `json:"current_level"`
	CurrentMin   uint32 `json:"current_min"`
	NextExp      int32  `json:"next_exp"`
}

type MyInfoProfile struct {
	Mid            uint64           `json:"mid"`
	Name           string           `json:"name"`
	Sex            string           `json:"sex"`
	Face           string           `json:"face"` // Face 头像链接
	Sign           string           `json:"sign"` // 个性签名
	Rank           int32            `json:"rank"`
	Level          uint8            `json:"level"` // 等级
	Birthday       uint64           `json:"birthday"`
	JoinTime       int64            `json:"jointime"`       // 应该没用，目前是0
	Moral          int32            `json:"moral"`          // 节操
	Silence        uint8            `json:"silence"`        // 封禁状态， 0：未封禁， 1：封禁
	EmailStatus    uint8            `json:"email_status"`   // 邮箱状态， 0：未绑定， 1：绑定
	TelStatus      uint8            `json:"tel_status"`     // 手机状态， 0：未绑定， 1：绑定
	Identification int8             `json:"identification"` // 作用不明确
	IsFakeAccount  int8             `json:"is_fake_account"`
	IsTourist      int8             `json:"is_tourist"`
	PinPrompting   int8             `json:"pin_prompting"` // 作用不明确
	Official       MyInfoOfficial   `json:"official"`
	NamePlate      MyInfoNamePlate  `json:"nameplate"`
	Vip            MyInfoProfileVip `json:"vip"`
	IsRipUser      bool             `json:"is_rip_user"` // 是否是风纪委用户
	IsRegAudit     int8             `json:"is_reg_audit"`
	CountryCode    string           `json:"country_code"`
	ExpertInfo     MyInfoExpertInfo `json:"expert_info"`
	Profession     MyInfoProfession `json:"profession"`
}

type MyInfoOfficial struct {
	Role  int32  `json:"role"`  // 认证类型
	Title string `json:"title"` // 认证信息
	Desc  string `json:"desc"`  // 认证备注
	Type  int8   `json:"type"`  // 是否认证，-1：未认证，0：个人认证
}

type MyInfoNamePlate struct {
	Nid        uint64 `json:"nid"`         // 勋章id
	Name       string `json:"name"`        // 勋章名称
	Image      string `json:"image"`       // 挂件图片url
	ImageSmall string `json:"image_small"` // 挂件图片url (小图)
	Level      string `json:"level"`       // 勋章等级
	Condition  string `json:"condition"`   // 勋章条件
}

type MyInfoExpertInfo struct {
	Title string `json:"title"`
	State int32  `json:"state"`
	Type  int32  `json:"type"`
	Desc  string `json:"desc"`
}

type MyInfoProfession struct {
	Id              int32  `json:"id"`
	Name            string `json:"name"`
	IsShow          int32  `json:"is_show"`
	CategoryOne     string `json:"category_one"`
	RealName        string `json:"realname"`
	Title           string `json:"title"`
	Department      string `json:"department"`
	CertificateNo   string `json:"certificate_no"`
	CertificateShow bool   `json:"certificate_show"`
}

type MyInfoProfileVip struct {
	Type            uint8             `json:"type"`     // 会员类型, 0：无会员，1：月度大会员，2：年度及以上大会员
	Status          uint8             `json:"status"`   // 会员状态, 0：无效，1：有效
	DueDate         uint64            `json:"due_date"` // 会员过期时间, 单位为毫秒时间戳
	ThemeType       uint32            `json:"theme_type"`
	Label           MyInfoVipLabel    `json:"label"`            // 会员标签
	AvatarSubscript uint8             `json:"avatar_subscript"` // 是否显示会员图标， 0：不显示，1：显示
	NicknameColor   string            `json:"nickname_color"`   //会员昵称颜色
	OttInfo         MyInfoVipOttInfo  `json:"OttInfo"`
	SuperVip        MyInfoVipSuperVip `json:"super_vip"`
}

type MyInfoVipLabel struct {
	Path                  string `json:"path"`
	Text                  string `json:"text"`
	LabelTheme            string `json:"label_theme"`
	TextColor             string `json:"text_color"`
	BgColor               string `json:"bg_color"`
	BorderColor           string `json:"border_color"`
	ImgLabelUriHantStatic string `json:"img_label_uri_hant_static"`
	ImgLabelUriHansStatic string `json:"img_label_uri_hans_static"`
	LabelId               int32  `json:"label_id"`
	BgStyle               int32  `json:"bg_style"`
	UseImgLabel           bool   `json:"use_img_label"`
}

type MyInfoVipOttInfo struct {
	VipType      int8   `json:"vip_type"`
	PayType      int8   `json:"pay_type"`
	PayChannelId string `json:"pay_channel_id"`
	Status       int8   `json:"status"`
	OverdueTime  uint64 `json:"overdue_time"`
}

type MyInfoVipSuperVip struct {
	IsSuperVip bool `json:"is_super_vip"`
}
