package utils

import "strconv"

const (
	// parsePlayURLNumKey 同时BiliBili解析播放链接的数量，默认为5
	parsePlayURLNumKey = "parse_play_url_num"
	// parsePlayURLSleepKey 解析后随机休眠上限（秒），默认 5，单位秒
	parsePlayURLSleepKey = "parse_play_url_sleep"
)

// GetParsePlayURLNum 获取同时BiliBili解析播放链接的数量，默认为3
func (s *Settings) GetParsePlayURLNum() (int, error) {
	num, err := s.GetKey(parsePlayURLNumKey)
	if err != nil {
		// 如果没有设置，默认返回5
		return 3, err
	}
	val, err := strconv.Atoi(num)
	if err != nil {
		return 3, err
	}

	return val, nil
}

func (s *Settings) SetParsePlayURLNum(num int) error {
	s.logger.Infof("Setting parse play url num to %d", num)

	return s.SetKey(parsePlayURLNumKey, strconv.Itoa(num))
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
	val, err := s.GetKey(parsePlayURLSleepKey)
	if err != nil {
		return 0, err
	}
	sleep, err := strconv.Atoi(val)
	if err != nil {
		return 0, err
	}
	return sleep, nil
}

func (s *Settings) SetParsePlayURLSleep(seconds int) error {
	return s.SetKey(parsePlayURLSleepKey, strconv.Itoa(seconds))
}

// GetParsePlayURLSleepSafe 获取解析后随机休眠上限（秒），默认 5
func (s *Settings) GetParsePlayURLSleepSafe() int {
	n, err := s.GetParsePlayURLSleep()
	if err != nil || n < 0 {
		return 5
	}
	return n
}
