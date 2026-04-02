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
