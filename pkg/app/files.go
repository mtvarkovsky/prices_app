package app

import (
	"context"
	"prices/pkg/config"
	"prices/pkg/files"
	"prices/pkg/files/processor"
	"prices/pkg/files/scanner"
	"prices/pkg/files/splitter"
	"prices/pkg/migrations"
	"prices/pkg/repository"
	"sync"
)

func RunFiles(ctx context.Context, config *config.FileProcessor) error {
	logger := getLogger("FilesApp")

	logger.Sugar().Infof("start FilesApp")

	logger.Sugar().Info("run migrations")
	err := migrations.MigrateDB(config.Storage.DSN)
	if err != nil {
		logger.Sugar().Errorf("unable to run migrations: (%s)", err.Error())
		return err
	}

	logger.Sugar().Infof("init prices repo for storage=%s", config.Storage.Type)
	pricesRepo, err := repository.NewPrices(config.Storage)
	if err != nil {
		logger.Sugar().Errorf("unable to init prices repo for storage=%s: (%s)", config.Storage.Type, err.Error())
		return err
	}

	wg := &sync.WaitGroup{}

	stopScanner := make(chan bool)
	stopSplitter := make(chan bool)
	stopProcessor := make(chan bool)

	filesQueue := files.NewFileQueueInMem(config.FilesQueueSize)
	filesSplitQueue := files.NewFileQueueInMem(config.FilesSplitQueueSize)

	filesCache := files.NewFileCacheInMem()

	scnnr := scanner.NewScanner(wg, logger, config, filesQueue, filesSplitQueue, filesCache, stopScanner)
	go scnnr.Scan()

	splttr := splitter.NewSplitter(wg, logger, config, filesSplitQueue, stopSplitter)
	go splttr.Split()

	prcssr := processor.NewProcessor(ctx, wg, config, filesQueue, pricesRepo, logger, stopProcessor)
	go prcssr.Process()

	<-ctx.Done()
	logger.Sugar().Infof("stopping FilesApp")

	_ = filesQueue.Close()

	stopScanner <- true
	stopSplitter <- true
	stopProcessor <- true

	wg.Wait()
	logger.Sugar().Infof("FilesApp stopped. Bye!")

	return nil
}
