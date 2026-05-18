package api

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/bytedance/sonic"
	"github.com/kamiertop/videodown/douyin/model"
	"golang.org/x/net/html"
)

// parseJingXuanAwemeID 解析精选页链接，提取awemeID
func (d *Douyin) parseJingXuanAwemeID(urlValue string) (string, error) {
	var htmlResp = d.client.
		Get(urlValue).
		SetHeaders(map[string]string{
			AcceptEncoding:  "identity;q=1, *;q=0",
			Referer:         "https://www.douyin.com/",
			UserAgent:       userAgent(),
			SecCHUAPlatform: fmt.Sprintf(`"%s"`, osName()),
			SecCHUAMobile:   "?0",
			SecCHUA:         `"Chromium";v="148", "Google Chrome";v="148", "Not/A)Brand";v="99"`,
		}).
		Do()
	if htmlResp.IsErrorState() {
		d.logger.Errorf("request jing xuan failed, status_code=%d", htmlResp.StatusCode)
		return "", fmt.Errorf("请求精选页失败, status_code=%d", htmlResp.StatusCode)
	}

	doc, err := html.Parse(htmlResp.Body)
	if err != nil {
		d.logger.Errorf("request jing xuan html failed: %v", err)
		return "", fmt.Errorf("解析精选页HTML失败: %w", err)
	}
	var scriptContent string
	var findScript func(*html.Node)

	// 递归遍历 DOM 节点
	findScript = func(n *html.Node) {
		// 找到 <script> 标签
		if n.Type == html.ElementNode && n.Data == "script" {
			// 遍历属性，匹配 id
			for _, attr := range n.Attr {
				if attr.Key == "id" && attr.Val == "RENDER_DATA" {
					// 提取 script 内部文本
					if n.FirstChild != nil {
						scriptContent = n.FirstChild.Data
					}
					return
				}
			}
		}

		// 递归子节点
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			findScript(child)
		}
	}

	findScript(doc)
	var errRetry = errors.New("解析失败，请尝试使用视频的分享链接进行解析")

	if scriptContent == "" {
		d.logger.Error("find script with id RENDER_DATA failed")
		return "", errRetry
	}

	unescape, err := url.QueryUnescape(scriptContent)
	if err != nil {
		d.logger.Errorf("url decode script content failed: %v", err)
		return "", errRetry
	}

	var renderData struct {
		App struct {
			VideoDetail struct {
				AwemeID string `json:"awemeId"`
			} `json:"videoDetail"`
		} `json:"app"`
	}
	if err = sonic.Unmarshal([]byte(unescape), &renderData); err != nil {
		d.logger.Errorf("unmarshal script content failed: %v", err)
		return "", errRetry
	}
	if renderData.App.VideoDetail.AwemeID == "" {
		d.logger.Error("awemeId is empty in render data")
		return "", errRetry
	}

	return renderData.App.VideoDetail.AwemeID, nil
}

// parseJingXuanVideo 解析精选页链接
// Deprecated: 解析HTML中的内容，返回值结构和之前的不同，所以不再使用
func (d *Douyin) parseJingXuanVideo(link string) (model.JingXuan, error) {
	if !strings.Contains(link, "douyin.com/jingxuan?modal_id=") {
		return model.JingXuan{}, errors.New("链接格式不正确")
	}
	var resp struct {
		App struct {
			VideoDetail model.JingXuan `json:"videoDetail"`
		} `json:"app"`
	}
	return resp.App.VideoDetail, nil
}
