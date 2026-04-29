package model

type MyInfoResponse struct {
	StatusCode int  `json:"status_code"`
	User       User `json:"user"`
}
