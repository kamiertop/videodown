package api

import (
	"github.com/dgraph-io/badger/v4"
	"github.com/imroc/req/v3"
	"github.com/kamiertop/videodown/logger"
)

type Douyin struct {
	logger *logger.Logger
	client *req.Client
	db     *badger.DB
}

func New(logger *logger.Logger, db *badger.DB) *Douyin {
	return &Douyin{
		logger: logger.WithName("Douyin"),
		client: req.C().EnableAutoDecompress(),
		db:     db,
	}
}

func NewTest() *Douyin {
	testLogger := logger.New().WithName("DouyinTest")
	return &Douyin{
		logger: testLogger,
		client: req.C().SetLogger(testLogger).EnableDebugLog(),
	}
}
