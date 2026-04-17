package utils

import (
	"errors"
	"net/url"
	"strings"
)

func ParseMid(spaceURLOrMid string) (string, error) {
	spaceURLOrMid = strings.TrimSpace(spaceURLOrMid)
	if spaceURLOrMid == "" {
		return "", errors.New("解析到mid为空，请输入正确的空间链接或mid")
	}

	if strings.Contains(spaceURLOrMid, "space.bilibili.com") {
		// 兼容不带协议的输入：space.bilibili.com/xxxxx
		if !strings.HasPrefix(spaceURLOrMid, "http://") && !strings.HasPrefix(spaceURLOrMid, "https://") {
			spaceURLOrMid = "https://" + spaceURLOrMid
		}
		parseUrl, err := url.Parse(spaceURLOrMid)
		if err != nil {
			return "", err
		}
		s := strings.Split(parseUrl.Path, "/")
		if len(s) >= 1 {
			spaceURLOrMid = s[1]
		}

	}

	if len(spaceURLOrMid) == 0 {
		return "", errors.New("解析到mid为空，请输入正确的空间链接或mid")
	}

	for _, c := range spaceURLOrMid {
		if c < '0' || c > '9' {
			return "", errors.New("解析到mid包含非数字字符，请输入正确的空间链接或mid")
		}
	}

	return spaceURLOrMid, nil
}
