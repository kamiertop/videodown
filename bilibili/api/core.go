package api

import (
	"github.com/dgraph-io/badger/v4"
	"github.com/imroc/req/v3"
	"github.com/kamiertop/videodown/logger"
)

type BiliBili struct {
	logger *logger.Logger
	client *req.Client
	db     *badger.DB
}

func New(logger *logger.Logger) *BiliBili {
	logger = logger.WithName("BiliBili")
	return &BiliBili{
		logger: logger,
		client: req.C().SetLogger(logger).EnableDebugLog(),
	}
}

func NewTest() *BiliBili {
	testLogger := logger.New().WithName("BiliBiliTest")
	return &BiliBili{
		logger: testLogger,
		client: req.C().SetLogger(testLogger).EnableDebugLog(),
	}
}
