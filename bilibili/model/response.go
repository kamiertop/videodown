package model

// ApiResponse 公共响应，用于拼接完整的响应
type ApiResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	TTL     int    `json:"ttl"`
}
