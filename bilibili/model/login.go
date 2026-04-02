package model

const SuccessCode = 0

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
