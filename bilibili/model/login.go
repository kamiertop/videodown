package model

type QRCodeResponse struct {
	Code    int32      `json:"code"`
	Message string     `json:"message"`
	Ttl     int32      `json:"ttl"`
	Data    QRCodeData `json:"data"`
}

type QRCodeData struct {
	Url       string `json:"url"`
	QRCodeKey string `json:"qrcode_key"`
}

type PollQRCodeResponse struct {
	Code    int32          `json:"code"`
	Message string         `json:"message"`
	Ttl     int32          `json:"ttl"`
	Data    PollQRCodeData `json:"data"`
}

type PollQRCodeData struct {
	Url          string `json:"url"`
	RefreshToken string `json:"refresh_token"`
	Timestamp    uint64 `json:"timestamp"`
	Code         uint32 `json:"code"`
	Message      string `json:"message"`
}

type RefreshResponse struct {
	Code    int32       `json:"code"`
	Message string      `json:"message"`
	Ttl     int32       `json:"ttl"`
	Data    RefreshData `json:"data"`
}

type RefreshData struct {
	Refresh   bool   `json:"refresh"`
	Timestamp uint64 `json:"timestamp"`
}

type LogOut struct {
	Code   int32      `json:"code"` // 0：成功，2202：csrf请求非法
	Status bool       `json:"status"`
	Ts     uint64     `json:"ts"` // 时间戳
	Data   LogOutData `json:"data"`
}
type LogOutData struct {
	RedirectUrl string `json:"redirectUrl"` // 退出登录后跳转的URL
}
