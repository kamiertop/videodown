package util

import "strconv"

// DownloadKindAudio 单独下载音频的记录类型；这类记录使用独立的 audio: 缓存键，
// 由下载包写入、历史接口按类型反查删除。
const DownloadKindAudio = "audio"

func DownloadCacheKey(cid int64) string {
	if cid == 0 {
		return ""
	}
	if cid < 0 {
		return CachePrefix + "cover:" + strconv.FormatInt(-cid, 10)
	}

	return CachePrefix + strconv.FormatInt(cid, 10)
}

// AudioDownloadCacheKey 返回单独下载音频的缓存键；与视频/封面键隔离，
// 避免音频记录被视频已下载判断命中。
func AudioDownloadCacheKey(cid int64) string {
	if cid == 0 {
		return ""
	}
	return CachePrefix + "audio:" + strconv.FormatInt(cid, 10)
}
