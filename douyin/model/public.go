package model

type Extra struct {
	FatalItemIds []any  `json:"fatal_item_ids"` //观测了几个接口，一直都是空的
	LogId        string `json:"log_id"`
	Now          int64  `json:"now"`
}

type LogPb struct {
	ImprId string `json:"impr_id"`
}
