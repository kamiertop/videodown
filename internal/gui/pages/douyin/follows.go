package douyin

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/ELY-Open-Source/MujicaUI/core"
	"github.com/ELY-Open-Source/MujicaUI/input"
	"github.com/ELY-Open-Source/MujicaUI/layout"
	"github.com/egoist/mygo/ui"
	"github.com/kamiertop/videodown/douyin/model"
	"github.com/kamiertop/videodown/internal/gui/style"
)

const followSearchDelay = 450 * time.Millisecond

func (s *State) syncFollowSearch() {
	query := strings.TrimSpace(s.followSearchInput)
	if query == s.followQuery {
		return
	}
	if s.cancelFollowSearch != nil {
		s.cancelFollowSearch()
		s.cancelFollowSearch = nil
	}
	s.followQuery = query
	s.followResults.seq++
	s.followResults = resource[[]model.SearchFollowResponse]{seq: s.followResults.seq}
	s.followGeneration++
	s.followAutoPending = false
	s.scroll("authors").Y = 0
	s.followSearchPending = query != ""
	if query == "" {
		return
	}
	epoch, generation, path := s.epoch, s.followGeneration, s.activePath
	run := func() {
		if s.stopped || s.epoch != epoch || s.followGeneration != generation || s.activePath != path || strings.TrimSpace(s.followSearchInput) != query {
			return
		}
		s.followSearchPending = false
		s.searchFollows()
	}
	if s.runtime.After != nil {
		s.cancelFollowSearch = s.runtime.After(followSearchDelay, run)
	} else {
		run()
	}
}
func (s *State) searchFollows() {
	query := strings.TrimSpace(s.followSearchInput)
	if query == "" || s.api == nil {
		return
	}
	if s.cancelFollowSearch != nil {
		s.cancelFollowSearch()
		s.cancelFollowSearch = nil
	}
	s.followSearchPending = false
	api := s.api
	load(s, &s.followResults, func() ([]model.SearchFollowResponse, error) { return api.SearchFollow(query) }, nil)
}
func (s *State) maybeMoreFollows() {
	if s.runtime.After == nil || s.followAutoPending || !s.follows.Loaded || s.follows.Loading || s.follows.Error != "" || !s.follows.Data.HasMore || strings.TrimSpace(s.followSearchInput) != "" {
		return
	}
	scroll := s.scroll("authors")
	if scroll.MaxY > 0 && scroll.Y+style.DynamicAutoLoadThreshold < scroll.MaxY {
		return
	}
	s.followAutoPending = true
	epoch, generation, path := s.epoch, s.followGeneration, s.activePath
	s.runtime.After(0, func() {
		if s.stopped || s.epoch != epoch || s.followGeneration != generation || s.activePath != path {
			return
		}
		s.followAutoPending = false
		if strings.TrimSpace(s.followSearchInput) != "" || scroll.MaxY > 0 && scroll.Y+style.DynamicAutoLoadThreshold < scroll.MaxY {
			return
		}
		s.loadFollows(true)
	})
}
func (s *State) authorsPage(c *ui.Context, navigate func(string)) {
	panel(c, "抖音用户工具栏", func() {
		ui.Row(c).FillWidth().Wrap().Padding(style.SpaceSmall).Gap(style.SpaceSmall).AlignItems(ui.Center).Children(func() {
			ui.Text(c, fmt.Sprintf("我的关注 · %d", s.follows.Data.Total)).Shrink(0).FontWeight(style.WeightSemibold)
			button(c, "刷新", input.ButtonOptions{Variant: input.Outline, Loading: s.follows.Loading}).OnClick(func() {
				if strings.TrimSpace(s.followSearchInput) != "" {
					s.searchFollows()
				} else {
					s.loadFollows(false)
				}
			})
			field(c.Key("follow-search"), &s.followSearchInput, "搜索已关注用户", "搜索已关注的昵称 / 抖音号", nil).Grow(1).MinWidth(180)
			field(c.Key("author"), &s.authorInput, "抖音用户主页", "输入用户主页链接或 sec_uid", func() { s.findAuthor(navigate) }).Grow(1).MinWidth(200)
			button(c, "解析", input.ButtonOptions{Variant: input.Primary}).OnClick(func() { s.findAuthor(navigate) })
		})
	})
	s.syncFollowSearch()
	searching := strings.TrimSpace(s.followSearchInput) != ""
	if searching {
		s.errorRow(c, s.followResults.Error, s.searchFollows)
	} else {
		s.errorRow(c, s.follows.Error, func() { s.loadFollows(s.follows.Loaded) })
	}
	layout.ScrollArea(c.Key("authors"), s.scroll("authors"), layout.ScrollAreaOptions{}, func() {
		ui.Column(c).FillWidth().Padding(0, style.ScrollbarGutter, style.SpaceMedium, 0).Gap(style.SpaceMedium).Children(func() {
			width, _ := c.Size()
			columns := max(1, min(style.AuthorGridMaxColumns, int((width-style.PlatformNavWidth-2*style.PagePadding)/style.AuthorCardMinWidth)))
			ui.Grid(c).Columns(columns).FillWidth().Gap(style.SpaceMedium).Children(func() {
				if searching {
					for _, user := range s.followResults.Data {
						avatar := user.AvatarURI
						s.authorCard(c.Key(user.SecUid), user.SecUid, user.NickName, user.ShortID, avatar, navigate)
					}
				} else {
					for _, user := range s.follows.Data.Followings {
						s.authorCard(c.Key(user.SecUid), user.SecUid, user.Nickname, user.Signature, first(user.AvatarThumb.UrlList), navigate)
					}
				}
			})
			switch {
			case searching && (s.followSearchPending || s.followResults.Loading):
				core.Spinner(c, core.SpinnerOptions{Size: 16, Label: "正在搜索关注", ShowLabel: true})
			case searching && s.followResults.Loaded && len(s.followResults.Data) == 0:
				empty(c, "未找到匹配的关注用户")
			case !searching && s.follows.Loading:
				core.Spinner(c, core.SpinnerOptions{Size: 16, Label: "正在加载关注用户", ShowLabel: true})
			case !searching && s.follows.Loaded && len(s.follows.Data.Followings) == 0:
				empty(c, "暂无关注用户")
			}
		})
	}).Grow(1).MinHeight(0)
	if !searching {
		s.maybeMoreFollows()
	}
}
func (s *State) authorCard(c *ui.Context, uid, name, signature, avatar string, navigate func(string)) {
	if uid == "" {
		return
	}
	k := core.Tokens(c)
	card := ui.ButtonBase(c).Label("查看抖音用户："+name).MinWidth(0).Padding(style.AuthorCardPadding).Gap(style.AuthorCardGap).Column().AlignItems(ui.Center).Radius(style.RadiusPanel).Border(style.BorderWidth, k.Border).Background(k.Surface).FocusRing(true)
	if card.Hovered() {
		card.Background(k.Selection).Border(style.BorderWidth, k.AccentText)
	}
	card.Children(func() {
		s.image(c, avatar, style.AuthorAvatarSize, style.AuthorAvatarSize, true)
		ui.Text(c, name).FillWidth().TextAlign(ui.Center).MaxLines(1).FontWeight(style.WeightSemibold)
		caption(c, signature)
	}).OnClick(func() { navigate("/douyin/user/" + url.PathEscape(uid)) })
}
