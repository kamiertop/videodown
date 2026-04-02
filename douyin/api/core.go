package api

import (
	"github.com/imroc/req/v3"
	"github.com/kamiertop/videodown/logger"
)

type Douyin struct {
	Logger *logger.Logger
	Client *req.Client
}

func New(logger *logger.Logger) *Douyin {
	return &Douyin{
		Logger: logger.WithName("Douyin"),
		Client: req.C(),
	}
}
