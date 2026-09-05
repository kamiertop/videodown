package util

import "strconv"

func DownloadCacheKey(cid int64) string {
	if cid == 0 {
		return ""
	}
	if cid < 0 {
		return CachePrefix + "cover:" + strconv.FormatInt(-cid, 10)
	}

	return CachePrefix + strconv.FormatInt(cid, 10)
}
