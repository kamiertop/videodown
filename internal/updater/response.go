package updater

// releaseResponse 请求API返回的响应，这里只取关心的字段，下载链接单独拼接
type releaseResponse struct {
	TagName string `json:"tag_name"` // 版本号
	Body    string `json:"body"`     // 发布说明
}
