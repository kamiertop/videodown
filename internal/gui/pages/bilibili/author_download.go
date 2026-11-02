package bilibili

import (
	"github.com/kamiertop/videodown/bilibili/model"
	"strings"
)

func bestAuthorStreams(play *model.VideoURLData) (videoURL, audioURL string) {
	videoBandwidth, videoPixels := -1, -1
	for _, stream := range play.Dash.Video {
		url := strings.TrimSpace(stream.BaseUrl)
		if url == "" {
			url = strings.TrimSpace(stream.BaseUrl1)
		}
		pixels := stream.Width * stream.Height
		if url != "" && (stream.Bandwidth > videoBandwidth || stream.Bandwidth == videoBandwidth && pixels > videoPixels) {
			videoURL, videoBandwidth, videoPixels = url, stream.Bandwidth, pixels
		}
	}
	if videoURL == "" && len(play.Durl) > 0 {
		videoURL = strings.TrimSpace(play.Durl[0].Url)
	}
	audioBandwidth := -1
	for _, stream := range play.Dash.Audio {
		url := strings.TrimSpace(stream.BaseUrl)
		if url == "" {
			url = strings.TrimSpace(stream.BaseUrl1)
		}
		if url != "" && stream.Bandwidth > audioBandwidth {
			audioURL, audioBandwidth = url, stream.Bandwidth
		}
	}
	return videoURL, audioURL
}
