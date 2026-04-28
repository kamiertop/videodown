package model

type VideoDetailResponse struct {
	StatusCode  int       `json:"status_code"`
	AwemeDetail AwemeItem `json:"aweme_detail"`
}
