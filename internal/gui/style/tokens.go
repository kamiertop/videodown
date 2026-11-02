// Package style 集中管理原生界面的字号、字重、圆角、间距和布局尺寸。
// 几何尺寸使用 DIP（设备无关像素），由界面框架处理屏幕缩放；
// 字号是辅助功能文字缩放前的基准值，使用时应通过 Font 或 Body 获取。
// 宽高比、透明度、行高倍数和数量上限没有 DIP 单位。
// 颜色由当前 MujicaUI 主题提供，不在此处固定。
package style

import (
	"github.com/ELY-Open-Source/MujicaUI/core"
	"github.com/egoist/mygo/ui"
)

const (
	// 抖音下载页：同一行中的画质选择和进度区域。
	DouyinDownloadProgressWidth float32 = 180 // 抖音下载卡片中进度条或休眠状态区域的宽度。
	DouyinDownloadQualityWidth  float32 = 240 // 抖音下载画质选择框宽度，给较长的画质名称预留空间。

	// 抖音竖屏视频卡片。
	DouyinMediaAspectRatio   float32 = 3.0 / 5.0       // 抖音封面宽高比，3:5 竖屏；与横屏视频卡片分开调整。
	DouyinMediaTextGap       float32 = 3               // 抖音标题与作者信息区域之间的间距。
	DouyinMediaMetaGap       float32 = 2               // 抖音用户名与发布时间之间的间距。
	DouyinMediaTitleHeight   float32 = 36              // 抖音标题区域最小高度，为两行标题预留空间。
	DouyinMediaRowGap        float32 = SpaceMedium - 1 // 抖音视频网格的行间距，比通用中间距少 1 DIP。
	DouyinMediaBadgeSize     float32 = 24              // 抖音图文、Live Photo 标识的外框边长。
	DouyinMediaBadgeIconSize float32 = 16              // 媒体类型标识内部 SVG 图标的边长。

	// 全局字号：修改后需通过 Font/Body 应用系统文字缩放。
	FontBody    float32 = 14 // 正文与常用页面文本。
	FontButton  float32 = 13 // 按钮文字。
	FontMeta    float32 = 12 // 作者、日期等辅助信息。
	FontCompact float32 = 10 // 侧栏等紧凑区域的小字。
	FontSection float32 = 16 // 卡片或页面分区标题。
	FontHeading float32 = 18 // 较醒目的标题。
	FontTitle   float32 = 20 // 账号昵称等主要标题。
	FontHero    float32 = 24 // 首页、关于页等突出展示的文字。

	// 全局字重与行高。
	WeightNormal           = 400 // 常规字重。
	WeightSemibold         = 600 // 半粗字重，用于按钮、标题等。
	WeightBold             = 700 // 粗体字重。
	LineHeight     float32 = 1.5 // 全局默认行高倍数，相对于当前字号计算。

	// 圆角：常规控件保持矩形轮廓，仅轻微圆角。
	RadiusWindow  float32 = 12 // 窗口外框圆角；窗口最大化时由窗口代码设为 0。
	RadiusPanel   float32 = 12 // 面板、卡片等容器的圆角。
	RadiusControl float32 = 6  // 普通按钮、输入框、选择框的统一圆角。
	RadiusMenu    float32 = 6  // 菜单项及相关操作按钮的圆角。
	RadiusMedia   float32 = 8  // 视频卡片与封面的圆角。
	RadiusChrome  float32 = 2  // 标题栏控件、时长标签等小元素的圆角。
	RadiusChoice  float32 = 6  // 设置页选项卡片的圆角。

	// 边框与通用间距。
	BorderWidth    float32 = 1  // 常规边框宽度。
	BorderEmphasis float32 = 2  // 选中或强调描边的宽度。
	SpaceSmall     float32 = 8  // 小间距，适合相关控件及卡片内边距。
	SpaceMedium    float32 = 12 // 中间距，适合卡片、网格及内容分组。
	SpaceLarge     float32 = 16 // 大间距。
	SpaceXLarge    float32 = 24 // 较大留白，适合登录提示、账号面板等。

	// 下载页：封面、控件、卡片留白与响应式断点。
	BatchCoverWidth           float32 = 96   // 紧凑下载卡片封面宽度。
	BatchCoverHeight          float32 = 54   // 紧凑下载卡片封面高度，与宽度组成 16:9。
	DownloadStreamWidth       float32 = 112  // 通用视频画质选择框宽度，抖音使用独立常量。
	DownloadAudioWidth        float32 = 136  // 音质选择框宽度。
	DownloadControlHeight     float32 = 30   // 下载卡片内按钮和选择框的紧凑高度。
	DownloadCardPadding       float32 = 10   // 下载卡片内边距。
	DownloadCardGap           float32 = 6    // 下载卡片内部信息行之间的间距。
	DownloadCoverWidth        float32 = 160  // 普通下载卡片及历史记录使用的封面宽度。
	DownloadCoverHeight       float32 = 90   // 普通下载卡片及历史记录使用的封面高度。
	DownloadInlineBreakpoint  float32 = 1000 // 下载页将控件放在同一行的窗口宽度阈值。
	DownloadCompactBreakpoint float32 = 700  // 下载页进一步收紧布局的窗口宽度阈值。
	// 应用外框、导航、输入框与滚动区域。
	PagePadding           float32 = 12  // 页面内容与外边界之间的留白。
	ScrollbarWidth        float32 = 4   // 全局滚动条宽度。
	ScrollbarGutter       float32 = 12  // 滚动内容侧边预留空间，避免滚动条覆盖卡片。
	TitlebarHeight        float32 = 38  // 自绘标题栏高度。
	PlatformNavWidth      float32 = 88  // B 站、抖音左侧导航栏宽度。
	PlatformNavItemHeight float32 = 70  // 平台侧栏单个导航按钮的高度。
	SettingsNavWidth      float32 = 176 // 设置页左侧分类导航宽度。
	CompactBreakpoint     float32 = 768 // 通用紧凑布局切换阈值，设置导航等使用。
	InputHeight           float32 = 36  // 常规单行输入框高度。
	// 账号页：头像、信息卡片和二维码。
	AvatarSize           float32 = 64  // 通用头像尺寸。
	ProfileAvatarSize    float32 = 80  // 已登录账号信息卡片中的头像尺寸。
	ProfileLoginMaxWidth float32 = 360 // 登录卡片及未登录提示卡片的最大宽度。
	ProfileCardMaxWidth  float32 = 640 // 账号信息卡片的最大宽度。
	ProfileFactHeight    float32 = 112 // 账号签名、会员状态等信息块的最小高度。
	QRCodeSize           float32 = 240 // 登录二维码的展示尺寸。
	// 关注用户卡片与作者详情页。
	AuthorCardMinWidth               float32 = 160  // 关注用户网格中卡片的最小参考宽度。
	AuthorAvatarSize                 float32 = 48   // 关注列表中作者头像的尺寸。
	AuthorCardPadding                float32 = 12   // 作者卡片内边距。
	AuthorCardGap                    float32 = 6    // 作者卡片内部元素间距。
	AuthorHintHeight                 float32 = 22   // 作者卡片悬浮提示区域的预留高度。
	AuthorDetailAvatarSize           float32 = 36   // 作者详情顶部资料栏的头像尺寸。
	AuthorDetailPaddingY             float32 = 6    // 作者详情资料栏上下内边距。
	AuthorDetailPaddingX             float32 = 10   // 作者详情资料栏左右内边距。
	AuthorDetailGap                  float32 = 8    // 作者详情资料栏内部元素间距。
	AuthorGridMaxColumns                     = 5    // 关注用户网格一行最多展示的卡片数量。
	AuthorListWidth                  float32 = 236  // 作者详情中合集、系列侧栏的宽度。
	AuthorListItemHeight             float32 = 64   // 作者合集、系列列表单项高度。
	AuthorListIconSize               float32 = 32   // 作者合集、系列列表图标尺寸。
	AuthorControlsDownloadBreakpoint float32 = 1040 // 作者及收藏工具栏切换紧凑布局的窗口宽度阈值。
	AuthorListViewportOffsetWide     float32 = 168  // 宽屏时估算作者列表可用高度需扣除的顶部区域高度。
	AuthorListViewportOffsetCompact  float32 = 212  // 紧凑布局下需扣除的顶部区域高度，包含额外换行空间。
	// 收藏夹、合集等资源列表。
	LibraryListWidth      float32 = 224 // 收藏夹、订阅合集等来源列表的侧栏宽度。
	LibraryListItemHeight float32 = 48  // 收藏来源列表单项高度。
	LibraryListIconSize   float32 = 30  // 收藏来源列表图标尺寸。
	LibraryHeaderHeight   float32 = 52  // 收藏来源列表与内容工具栏的对齐高度。
	// 动态列表与自动分页。
	DynamicColumnsHeight     float32 = 36  // 动态列表表头高度。
	DynamicRowHeight         float32 = 86  // 动态列表每条视频记录的高度。
	DynamicThumbWidth        float32 = 112 // 动态列表视频缩略图宽度。
	DynamicThumbHeight       float32 = 63  // 动态列表视频缩略图高度。
	DynamicColumnGap         float32 = 16  // 动态列表各列之间的间距。
	DynamicAuthorWidth       float32 = 128 // 动态列表作者列宽度。
	DynamicTimeWidth         float32 = 132 // 动态列表时间列宽度。
	DynamicActionWidth       float32 = 64  // 动态列表操作列宽度。
	DynamicAutoLoadThreshold float32 = 120 // 距滚动底部小于此值时触发下一页加载，其他列表也会复用。
	DynamicAutoFillLimit             = 6   // 内容不足一屏时，连续自动补充的最大页数。
	// 横屏视频卡片及封面信息层，部分文字和标签样式也供抖音复用。
	MediaAspectRatio          float32 = 16.0 / 9 // 通用横屏视频封面宽高比，16:9；抖音使用独立比例。
	MediaTitleHeight          float32 = 42       // 通用视频卡片标题区域最小高度。
	MediaAuthorUnderlineWidth float32 = 2        // 作者链接悬浮下划线的粗细。
	MediaAuthorUnderlineGap   float32 = 4        // 作者名称与悬浮下划线之间的留白。
	MediaStatFontSize         float32 = 12       // 封面上播放量、弹幕等统计文字的基准字号。
	MediaStatIconSize         float32 = 13       // 封面统计图标尺寸。
	MediaStatGap              float32 = 2        // 统计图标与对应数字之间的间距。
	MediaDurationFontSize     float32 = 12       // 视频时长标签文字的基准字号。
	MediaDurationPaddingX     float32 = 4        // 时长标签左右内边距。
	MediaDurationPaddingY     float32 = 1        // 时长标签上下内边距。
	MediaDurationOpacity      float32 = .65      // 时长标签黑色背景的不透明度，取值 0～1。
	MediaOverlayPaddingX      float32 = 7        // 封面底部信息层的左右内边距。
	MediaOverlayPaddingY      float32 = 4        // 封面底部信息层的上下内边距。
)

// Font 将基准字号转换为遵循系统辅助功能文字缩放的实际字号。
func Font(c *ui.Context, size float32) float32 { return core.FontSize(c, size) }

// Body 返回应用默认正文字号，已经应用系统文字缩放。
func Body(c *ui.Context) float32 { return Font(c, FontBody) }

// Apply 在 core.Use 设置主题后调用，为原生控件设置默认字号、滚动条宽度和行高。
// 先复制主题再修改，避免直接修改共用主题对象。
func Apply(c *ui.Context) {
	theme := *c.Theme()
	theme.FontSize = Body(c)
	theme.ScrollbarWidth = ScrollbarWidth
	c.SetTheme(&theme)
	c.Root().FontSize(theme.FontSize).FontWeight(WeightNormal).LineHeight(LineHeight)
}
