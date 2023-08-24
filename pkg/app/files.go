package app

import (
	"context"
	"prices/pkg/config"
	"prices/pkg/files"
	"prices/pkg/migrations"
	"prices/pkg/repository"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func RunFiles(ctx context.Context, config *config.FileProcessor) error {
	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.TimeKey = "timestamp"
	encoderCfg.EncodeTime = zapcore.RFC3339TimeEncoder

	zapCfg := zap.NewProductionConfig()
	zapCfg.EncoderConfig = encoderCfg

	logger := zap.Must(zapCfg.Build())

	log := logger.Named("FilesApp")

	log.Sugar().Infof("start FilesApp")

	log.Sugar().Info("run migrations")
	err := migrations.MigrateDB(config.Storage.DSN)
	if err != nil {
		log.Sugar().Errorf("unable to run migrations: (%s)", err.Error())
		return err
	}

	log.Sugar().Infof("init prices repo for storage=%s", config.Storage.Type)
	pricesRepo, err := repository.NewPrices(config.Storage)
	if err != nil {
		log.Sugar().Errorf("unable to init prices repo for storage=%s: (%s)", config.Storage.Type, err.Error())
		return err
	}

	wg := &sync.WaitGroup{}

	stopScanner := make(chan bool)
	stopSplitter := make(chan bool)
	stopProcessor := make(chan bool)

	filesQueue := files.NewFileQueueInMem(config.FilesQueueSize)
	filesSplitQueue := files.NewFileQueueInMem(config.FilesSplitQueueSize)

	filesCache := files.NewFileCacheInMem()

	scanner := files.NewScanner(wg, log, config, filesQueue, filesSplitQueue, filesCache, stopScanner)
	go scanner.Scan()

	splitter := files.NewSplitter(wg, log, config, filesSplitQueue, stopSplitter)
	go splitter.Split()

	processor := files.NewProcessor(ctx, wg, config, filesQueue, pricesRepo, logger, stopProcessor)
	go processor.Process()

	<-ctx.Done()
	log.Sugar().Infof("stopping FilesApp")

	_ = filesQueue.Close()

	stopScanner <- true
	stopSplitter <- true
	stopProcessor <- true

	wg.Wait()
	log.Sugar().Infof("FilesApp stopped. Bye!")

	return nil
}
