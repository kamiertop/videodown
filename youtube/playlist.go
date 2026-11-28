package youtube

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/kamiertop/videodown/youtube/internal/extractor"
)

type Playlist struct {
	ID, Title, Continuation string
	Videos                  []Video
}

func PlaylistID(input string) (string, error) { return extractor.PlaylistID(input) }

// accountContinuationPrefix keeps pagination on the same authentication mode as the first page.
const accountContinuationPrefix = "account:"

func (s *Service) Playlist(ctx context.Context, input, continuation string) (Playlist, error) {
	account := s.client()
	client := &extractor.Client{HTTPClient: account.HTTPClient}
	token, useAccount := strings.CutPrefix(continuation, accountContinuationPrefix)
	if useAccount {
		if account.Cookie == "" {
			return Playlist{}, errors.New("请先在账号页设置 Cookie，再重新加载播放列表")
		}
		client = account
	}
	page, err := client.GetPlaylistPage(ctx, input, token)
	if err != nil && continuation == "" && account.Cookie != "" && ctx.Err() == nil &&
		(requiresAccount(err) || errors.Is(err, extractor.ErrPlaylistNotFound)) {
		useAccount = true
		page, err = account.GetPlaylistPage(ctx, input, "")
	}
	if err != nil {
		if errors.Is(err, extractor.ErrPlaylistNotFound) {
			if useAccount {
				return Playlist{}, fmt.Errorf("%w。请确认当前账号能打开此播放列表", err)
			}
			if continuation == "" && account.Cookie == "" {
				return Playlist{}, fmt.Errorf("%w。如果是私人列表，请在账号页登录有权限的账号", err)
			}
		}
		if requiresAccount(err) {
			return Playlist{}, s.parseError(err)
		}
		return Playlist{}, err
	}
	if useAccount && page.Continuation != "" {
		page.Continuation = accountContinuationPrefix + page.Continuation
	}
	result := Playlist{ID: page.ID, Title: page.Title, Continuation: page.Continuation}
	for _, entry := range page.Videos {
		result.Videos = append(result.Videos, Video{ID: entry.ID, Title: entry.Title, Author: entry.Author, Duration: entry.Duration, Cover: "https://i.ytimg.com/vi/" + entry.ID + "/hqdefault.jpg"})
	}

	return result, nil
}
