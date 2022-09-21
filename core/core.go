package core

import (
	"github.com/broothie/slink.chat/async"
	"github.com/broothie/slink.chat/config"
	"github.com/broothie/slink.chat/db"
	"github.com/broothie/slink.chat/search"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Core struct {
	Config *config.Config
	Logger *zap.Logger
	DB     *db.DB
	Search search.Search
	Async  *async.Async
}

func New(cfg *config.Config) (Core, error) {
	loggerBuilder := zap.NewDevelopment
	if cfg.IsHosted() {
		encoderConfig := zap.NewProductionEncoderConfig()
		encoderConfig.MessageKey = "message"
		encoderConfig.LevelKey = "severity"
		encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

		cfg := zap.NewProductionConfig()
		cfg.EncoderConfig = encoderConfig

		loggerBuilder = cfg.Build
	}

	logger, err := loggerBuilder()
	if err != nil {
		return Core{}, errors.Wrap(err, "failed to create logger")
	}

	db, err := db.New(cfg)
	if err != nil {
		return Core{}, errors.Wrap(err, "failed to create new db")
	}

	var src search.Search = search.NewAlgolia(cfg)
	if cfg.IsLocal() {
		src = search.NewDB(db)
	}

	async, err := async.New(cfg)
	if err != nil {
		return Core{}, errors.Wrap(err, "failed to create async")
	}

	return Core{
		Config: cfg,
		Logger: logger,
		DB:     db,
		Search: src,
		Async:  async,
	}, err
}
