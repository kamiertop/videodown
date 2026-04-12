package model

type UserInfoData struct {
	Mid        uint64  `json:"mid"`
	Name       string  `json:"name"`
	Sex        string  `json:"sex"`
	Face       string  `json:"face"`
	Sign       string  `json:"sign"`
	Rank       int     `json:"rank"`
	Level      int     `json:"level"`
	Birthday   string  `json:"birthday"`
	IsFollowed bool    `json:"is_followed"`
	IsRisk     bool    `json:"is_risk"`
	Vip        UserVip `json:"vip"`
}
