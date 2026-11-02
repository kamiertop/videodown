package youtube

import (
	"fmt"
	"mime"
	"sort"
	"strings"

	"github.com/kamiertop/videodown/youtube/internal/extractor"
)

func usableFormats(v *extractor.Video) bool {
	description := describe(v)
	return len(description.Formats)+len(description.AudioFormats) > 0
}

// Keep ordinary itags compatible with saved tasks. DRC is a distinct rendition
// of the same itag, so it needs its own stable application selection ID.
const drcAudioIDOffset = 1_000_000

func audioID(f *extractor.Format) int {
	if f.IsDRC {
		return f.ItagNo + drcAudioIDOffset
	}
	return f.ItagNo
}

func audioBitrate(f *extractor.Format) int {
	if f.AverageBitrate > 0 {
		return f.AverageBitrate
	}
	return f.Bitrate
}

// This application's audio picker supports the default language. Collapse
// repeated copies of each default-track rendition, retaining DRC as a separate
// choice. Both the picker and media download use this same canonical pool.
func selectableAudioFormats(formats extractor.FormatList) []*extractor.Format {
	var result []*extractor.Format
	byID := map[int]int{}
	for i := range formats {
		f := &formats[i]
		if f.ItagNo <= 0 || !strings.HasPrefix(f.MimeType, "audio/") || (f.URL == "" && f.Cipher == "") || (f.AudioTrack != nil && !f.AudioTrack.AudioIsDefault) {
			continue
		}
		id := audioID(f)
		if index, ok := byID[id]; ok {
			old := result[index]
			// Prefer an explicitly identified default track to an unlabelled
			// duplicate, then the more complete bitrate rendition.
			if (old.AudioTrack == nil && f.AudioTrack != nil) || ((old.AudioTrack == nil) == (f.AudioTrack == nil) && audioBitrate(f) > audioBitrate(old)) {
				result[index] = f
			}
			continue
		}
		byID[id] = len(result)
		result = append(result, f)
	}
	sort.SliceStable(result, func(i, j int) bool {
		if result[i].IsDRC != result[j].IsDRC {
			return !result[i].IsDRC
		}
		return audioBitrate(result[i]) > audioBitrate(result[j])
	})
	return result
}

func audioFormat(f extractor.Format) Format {
	media, params, _ := mime.ParseMediaType(f.MimeType)
	codec := params["codecs"]
	if strings.HasPrefix(codec, "mp4a") {
		codec = "AAC"
	}
	if strings.HasPrefix(codec, "opus") {
		codec = "Opus"
	}
	if codec == "" {
		codec = "音频"
	}
	container := strings.ToUpper(strings.TrimPrefix(media, "audio/"))
	if container == "MP4" {
		container = "M4A"
	}
	bitrate := f.AverageBitrate
	if bitrate <= 0 {
		bitrate = f.Bitrate
	}
	summary := fmt.Sprintf("%s · %d kbps", codec, bitrate/1000)
	variant := "标准音频"
	if f.IsDRC {
		variant = "动态压缩(DRC)"
		summary += " · DRC"
	}
	heading := fmt.Sprintf("%s · %d kbps · %s", codec, bitrate/1000, container)
	return Format{ID: audioID(&f), Label: fmt.Sprintf("%s · %d声道 · %s · #%d", heading, f.AudioChannels, variant, f.ItagNo), Summary: summary, Audio: true, Mime: f.MimeType}
}
