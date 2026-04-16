package model

type MyInfo struct {
	OwnerSecUid  string       `json:"owner_sec_uid"`
	StatusCode   int          `json:"status_code"`
	NextReqCount int          `json:"next_req_count"`
	Data         []MyInfoData `json:"data"`
}

type MyInfoData struct {
	Nickname    string `json:"nickname"`
	SecUid      string `json:"sec_uid"`
	ShortId     string `json:"short_id"`
	Signature   string `json:"signature"`
	Uid         string `json:"uid"`
	UniqueId    string `json:"unique_id"`
	AvatarThumb Avatar `json:"avatar_thumb"`
}
