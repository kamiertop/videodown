package utils

import (
	"strconv"

	"github.com/kamiertop/videodown/internal/constant"
)

const (
	// bilibiliDefaultQNKey 默认视频清晰度 ID。
	bilibiliDefaultQNKey = "bilibili_default_qn"
)

// GetBilibiliDefaultQN returns the preferred Bilibili quality ID, when set.
func (s *Settings) GetBilibiliDefaultQN() (string, error) {
	return s.store.Get(bilibiliDefaultQNKey)
}

// GetParsePlayURLNum 获取同时BiliBili解析播放链接的数量，默认为3
func (s *Settings) GetParsePlayURLNum() (int, error) {
	return s.store.ParsePlayURLNum()
}

func (s *Settings) SetParsePlayURLNum(num int) error {
	s.logger.Infof("Setting parse play url num to %d", num)

	return s.store.Set(constant.ParsePlayURLNumKey, strconv.Itoa(num))
}

// GetParsePlayURLNumSafe 获取并发解析数，出错时返回默认值 3
func (s *Settings) GetParsePlayURLNumSafe() int {
	num, err := s.GetParsePlayURLNum()
	if err != nil || num <= 0 {
		return 3
	}

	return num
}

func (s *Settings) GetParsePlayURLSleep() (int, error) {
	return s.store.ParsePlayURLSleep()
}

func (s *Settings) SetParsePlayURLSleep(seconds int) error {
	return s.store.Set(constant.ParsePlayURLSleepKey, strconv.Itoa(seconds))
}

// GetParsePlayURLSleepSafe 获取解析后随机休眠上限（秒），默认 5
func (s *Settings) GetParsePlayURLSleepSafe() int {
	n, err := s.GetParsePlayURLSleep()
	if err != nil || n < 0 {
		return 5
	}
	return n
}
