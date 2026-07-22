package model

import (
	"strconv"
	"strings"
)

const (
	DynamicTypeAV      = "DYNAMIC_TYPE_AV"
	DynamicTypeForward = "DYNAMIC_TYPE_FORWARD"
)

const (
	MajorTypeArchive   = "MAJOR_TYPE_ARCHIVE"
	MajorTypeUgcSeason = "MAJOR_TYPE_UGC_SEASON"
)

const (
	AdditionalTypeUGC = "ADDITIONAL_TYPE_UGC"
)

// DynamicData is the data payload returned by /x/polymer/web-dynamic/v1/feed/all.
type DynamicData struct {
	HasMore        bool          `json:"has_more"`
	Items          []DynamicItem `json:"items"`
	Offset         string        `json:"offset"`
	UpdateBaseline string        `json:"update_baseline"`
	UpdateNum      DynamicInt    `json:"update_num"`
}

type DynamicItem struct {
	Basic   DynamicBasic     `json:"basic"`
	IDStr   string           `json:"id_str"`
	Modules DynamicModules   `json:"modules"`
	Orig    *DynamicOrigItem `json:"orig,omitempty"`
	Type    string           `json:"type"`
	Visible bool             `json:"visible"`
}

type DynamicOrigItem struct {
	Basic   DynamicBasic   `json:"basic"`
	IDStr   string         `json:"id_str"`
	Modules DynamicModules `json:"modules"`
	Type    string         `json:"type"`
	Visible bool           `json:"visible"`
}

type DynamicBasic struct {
	CommentIDStr string          `json:"comment_id_str"`
	CommentType  int             `json:"comment_type"`
	JumpURL      string          `json:"jump_url"`
	LikeIcon     DynamicLikeIcon `json:"like_icon"`
	RidStr       string          `json:"rid_str"`
}

type DynamicLikeIcon struct {
	ActionURL string     `json:"action_url"`
	EndURL    string     `json:"end_url"`
	ID        DynamicInt `json:"id"`
	StartURL  string     `json:"start_url"`
}

type DynamicModules struct {
	ModuleAuthor      DynamicAuthor `json:"module_author"`
	ModuleDispute     any           `json:"module_dispute"`
	ModuleDynamic     DynamicModule `json:"module_dynamic"`
	ModuleFold        any           `json:"module_fold"`
	ModuleInteraction any           `json:"module_interaction"`
	ModuleMore        any           `json:"module_more"`
	ModuleShareInfo   any           `json:"module_share_info"`
	ModuleStat        DynamicStat   `json:"module_stat"`
	ModuleTag         any           `json:"module_tag"`
}

type DynamicAuthor struct {
	Face            string     `json:"face"`
	Following       any        `json:"following"`
	IsTop           bool       `json:"is_top"`
	JumpURL         string     `json:"jump_url"`
	Mid             int64      `json:"mid"`
	Name            string     `json:"name"`
	PubAction       string     `json:"pub_action"`
	PubLocationText string     `json:"pub_location_text"`
	PubTime         string     `json:"pub_time"`
	PubTs           DynamicInt `json:"pub_ts"`
	Type            string     `json:"type"`
}

type DynamicModule struct {
	Additional *DynamicAdditional `json:"additional"`
	Desc       *DynamicDesc       `json:"desc"`
	Major      *DynamicMajor      `json:"major"`
	Topic      any                `json:"topic"`
}

type DynamicDesc struct {
	Text          string        `json:"text"`
	RichTextNodes []DynamicText `json:"rich_text_nodes"`
}

type DynamicText struct {
	OrigText string `json:"orig_text"`
	Text     string `json:"text"`
	Type     string `json:"type"`
}

type DynamicAdditional struct {
	Type string                `json:"type"`
	UGC  *DynamicAdditionalUGC `json:"ugc"`
}

type DynamicAdditionalUGC struct {
	Cover      string `json:"cover"`
	DescSecond string `json:"desc_second"`
	Duration   string `json:"duration"`
	HeadText   string `json:"head_text"`
	IDStr      string `json:"id_str"`
	JumpURL    string `json:"jump_url"`
	MultiLine  bool   `json:"multi_line"`
	Title      string `json:"title"`
}

type DynamicMajor struct {
	Type      string          `json:"type"`
	Archive   *DynamicArchive `json:"archive"`
	UgcSeason *DynamicArchive `json:"ugc_season"`
}

type DynamicArchive struct {
	Aid            string              `json:"aid"`
	Badge          DynamicArchiveBadge `json:"badge"`
	Bvid           string              `json:"bvid"`
	Cover          string              `json:"cover"`
	Desc           string              `json:"desc"`
	DisablePreview int                 `json:"disable_preview"`
	DurationText   string              `json:"duration_text"`
	JumpURL        string              `json:"jump_url"`
	Stat           DynamicArchiveStat  `json:"stat"`
	Title          string              `json:"title"`
	Type           int                 `json:"type"`
}

type DynamicArchiveBadge struct {
	BgColor string `json:"bg_color"`
	Color   string `json:"color"`
	IconURL string `json:"icon_url"`
	Text    string `json:"text"`
}

type DynamicArchiveStat struct {
	Danmaku string `json:"danmaku"`
	Play    string `json:"play"`
}

type DynamicStat struct {
	Comment DynamicStatItem `json:"comment"`
	Forward DynamicStatItem `json:"forward"`
	Like    DynamicStatItem `json:"like"`
}

type DynamicStatItem struct {
	Count     int  `json:"count"`
	Forbidden bool `json:"forbidden"`
	Hidden    bool `json:"hidden"`
	Status    bool `json:"status"`
}

type DynamicInt int64

func (d *DynamicInt) UnmarshalJSON(data []byte) error {
	raw := strings.Trim(string(data), `"`)
	if raw == "" || raw == "null" {
		*d = 0
		return nil
	}

	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return err
	}
	*d = DynamicInt(value)
	return nil
}

// DynamicArchivePage is a Wails-friendly single return value for one dynamic page.
type DynamicArchivePage struct {
	Items   []DynamicArchiveItem `json:"items"`
	Offset  string               `json:"offset"`
	HasMore bool                 `json:"has_more"`
}

// DynamicArchiveItem is the flattened video payload used by downstream download flows.
type DynamicArchiveItem struct {
	Bvid         string `json:"bvid"`
	Title        string `json:"title"`
	Cover        string `json:"cover"`
	DurationText string `json:"duration_text"`
	AuthorName   string `json:"author_name"`
	PubTime      string `json:"pub_time"`
	PubTs        int64  `json:"pub_ts"`
}

func (d DynamicData) GetArchiveItems() []DynamicArchiveItem {
	items := make([]DynamicArchiveItem, 0, len(d.Items))
	for _, item := range d.Items {
		items = appendArchiveItem(items, item.Type, item.Modules)
		if item.Type == DynamicTypeForward && item.Orig != nil {
			items = appendArchiveItem(items, item.Orig.Type, item.Orig.Modules)
		}
	}
	return items
}

func appendArchiveItem(items []DynamicArchiveItem, dynamicType string, modules DynamicModules) []DynamicArchiveItem {
	if dynamicType == DynamicTypeAV {
		if item, ok := archiveItemFromMajor(modules); ok {
			return append(items, item)
		}
	}

	if item, ok := archiveItemFromAdditional(modules); ok {
		return append(items, item)
	}

	return items
}

func archiveItemFromMajor(modules DynamicModules) (DynamicArchiveItem, bool) {
	if modules.ModuleDynamic.Major == nil {
		return DynamicArchiveItem{}, false
	}

	major := modules.ModuleDynamic.Major
	archive := major.Archive
	if major.Type == MajorTypeUgcSeason {
		archive = major.UgcSeason
	}
	if archive == nil || (major.Type != MajorTypeArchive && major.Type != MajorTypeUgcSeason) {
		return DynamicArchiveItem{}, false
	}

	return DynamicArchiveItem{
		Bvid:         archive.Bvid,
		Title:        archive.Title,
		Cover:        archive.Cover,
		DurationText: archive.DurationText,
		AuthorName:   modules.ModuleAuthor.Name,
		PubTime:      modules.ModuleAuthor.PubTime,
		PubTs:        int64(modules.ModuleAuthor.PubTs),
	}, true
}

func archiveItemFromAdditional(modules DynamicModules) (DynamicArchiveItem, bool) {
	additional := modules.ModuleDynamic.Additional
	if additional == nil || additional.Type != AdditionalTypeUGC || additional.UGC == nil {
		return DynamicArchiveItem{}, false
	}

	ugc := additional.UGC
	return DynamicArchiveItem{
		Bvid:         bvidFromJumpURL(ugc.JumpURL),
		Title:        ugc.Title,
		Cover:        ugc.Cover,
		DurationText: ugc.Duration,
		AuthorName:   modules.ModuleAuthor.Name,
		PubTime:      modules.ModuleAuthor.PubTime,
		PubTs:        int64(modules.ModuleAuthor.PubTs),
	}, true
}

func bvidFromJumpURL(jumpURL string) string {
	start := strings.Index(jumpURL, "/video/")
	if start < 0 {
		return ""
	}

	bvid := strings.TrimPrefix(jumpURL[start:], "/video/")
	if slash := strings.IndexByte(bvid, '/'); slash >= 0 {
		bvid = bvid[:slash]
	}
	if question := strings.IndexByte(bvid, '?'); question >= 0 {
		bvid = bvid[:question]
	}
	return bvid
}
