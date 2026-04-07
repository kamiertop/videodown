package api

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"time"
)

// WBI 打乱表（来自 B站官方文档）
var mixinKeyEncTab = []int{
	46, 47, 18, 2, 53, 8, 23, 32, 15, 50, 10, 31, 58, 3, 45, 35, 27, 43, 5, 49,
	33, 9, 42, 19, 29, 28, 14, 39, 12, 38, 41, 13, 37, 48, 7, 16, 24, 55, 40,
	61, 26, 17, 0, 1, 60, 51, 30, 4, 22, 25, 54, 21, 56, 59, 6, 63, 57, 62, 11,
	36, 20, 34, 44, 52,
}

// wbiKeys 存储从 nav 接口获取的 wbi 密钥
type wbiKeys struct {
	ImgKey string
	SubKey string
}

// genMixinKey 根据 img_key + sub_key 生成 mixin_key
func genMixinKey(imgKey, subKey string) string {
	rawWbiKey := imgKey + subKey
	bytes := []byte(rawWbiKey)
	var mixin strings.Builder
	mixin.Grow(32)

	for _, idx := range mixinKeyEncTab {
		if idx < len(bytes) {
			mixin.WriteByte(bytes[idx])
		}
	}

	// 截取前32位
	result := mixin.String()
	if len(result) > 32 {
		return result[:32]
	}
	return result
}

// encWbiParams 为参数做 WBI 签名：添加 wts，并计算 w_rid
func (b *BiliBili) encWbiParams(params map[string]string) (map[string]string, error) {
	wbiKey, err := b.getWbiKeys()
	if err != nil {
		return nil, err
	}
	// 1. 生成 mixin_key
	mixinKey := genMixinKey(wbiKey.ImgKey, wbiKey.SubKey)

	// 2. 添加 wts（当前秒级时间戳）
	now := time.Now().Unix()
	params["wts"] = fmt.Sprintf("%d", now)

	// 3. 排序 key
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// 4. 过滤 value 中的 "!'()*" 并按 encodeURIComponent 编码
	var encodedPairs []string
	for _, k := range keys {
		v := params[k]
		// 过滤特殊字符
		filtered := strings.Map(func(r rune) rune {
			if r == '!' || r == '\'' || r == '(' || r == ')' || r == '*' {
				return -1 // 删除这些字符
			}
			return r
		}, v)

		// URL 编码
		keyEnc := url.QueryEscape(k)
		valEnc := url.QueryEscape(filtered)
		encodedPairs = append(encodedPairs, fmt.Sprintf("%s=%s", keyEnc, valEnc))
	}

	query := strings.Join(encodedPairs, "&")

	// 5. 计算 w_rid = MD5(query + mixin_key)
	signStr := query + mixinKey
	hash := md5.Sum([]byte(signStr))
	wRid := hex.EncodeToString(hash[:])

	params["w_rid"] = wRid

	return params, nil
}

// getWbiKeys 从 nav 接口获取 img_key 和 sub_key
func (b *BiliBili) getWbiKeys() (*wbiKeys, error) {

	type NavData struct {
		WbiImg struct {
			ImgURL string `json:"img_url"`
			SubURL string `json:"sub_url"`
		} `json:"wbi_img"`
	}

	var resp struct {
		Code    int     `json:"code"`
		Message string  `json:"message"`
		Data    NavData `json:"data"`
	}
	cookie, err := b.getCookies()
	if err != nil {
		return nil, err
	}
	err = b.client.
		Get("https://api.bilibili.com/x/web-interface/nav").
		SetHeader("Cookie", cookie).
		SetHeaders(publicHeaders()).
		Do().
		Into(&resp)

	if err != nil {
		return nil, fmt.Errorf("请求nav接口失败: %w", err)
	}

	if resp.Code != 0 {
		return nil, fmt.Errorf("nav接口返回错误: %s", resp.Message)
	}

	// 从 URL 中提取 img_key 和 sub_key（文件名去掉后缀）
	imgKey := extractKeyFromURL(resp.Data.WbiImg.ImgURL)
	subKey := extractKeyFromURL(resp.Data.WbiImg.SubURL)

	if imgKey == "" || subKey == "" {
		return nil, fmt.Errorf("提取wbi密钥失败")
	}

	return &wbiKeys{
		ImgKey: imgKey,
		SubKey: subKey,
	}, nil
}

// extractKeyFromURL 从 wbi 图片 URL 中提取 key（去掉路径和后缀）
func extractKeyFromURL(imgURL string) string {
	// 例如: https://i0.hdslb.com/bfs/wbi/7cd084941338484aae1ad9425b84077c.png
	// 提取: 7cd084941338484aae1ad9425b84077c

	// 找到最后一个 /
	lastSlash := strings.LastIndex(imgURL, "/")
	if lastSlash == -1 {
		return ""
	}

	filename := imgURL[lastSlash+1:]

	// 去掉文件扩展名
	dotIndex := strings.LastIndex(filename, ".")
	if dotIndex == -1 {
		return filename
	}

	return filename[:dotIndex]
}
