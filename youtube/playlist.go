package youtube

import (
	"context"
	"github.com/kamiertop/videodown/youtube/internal/extractor"
)

type Playlist struct {
	ID, Title, Continuation string
	Videos                  []Video
}

func PlaylistID(input string) (string, error) { return extractor.PlaylistID(input) }

func (s *Service) Playlist(ctx context.Context, input, continuation string) (Playlist, error) {
	page, err := s.client().GetPlaylistPage(ctx, input, continuation)
	if err != nil {
		return Playlist{}, err
	}
	result := Playlist{ID: page.ID, Title: page.Title, Continuation: page.Continuation}
	for _, entry := range page.Videos {
		result.Videos = append(result.Videos, Video{ID: entry.ID, Title: entry.Title, Author: entry.Author, Duration: entry.Duration, Cover: "https://i.ytimg.com/vi/" + entry.ID + "/hqdefault.jpg"})
	}
	return result, nil
}
