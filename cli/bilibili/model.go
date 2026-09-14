package bilibili

import bilibiliapi "github.com/kamiertop/videodown/bilibili/api"

type Core struct {
	bilibiliapi.BiliBili
}

func NewCore() *Core {
	return &Core{}
}
