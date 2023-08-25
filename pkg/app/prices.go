package app

import (
	"context"
	"errors"
	"fmt"
	"github.com/Depado/ginprom"
	"net/http"
	"prices/pkg/api"
	"prices/pkg/config"
	"prices/pkg/migrations"
	"prices/pkg/repository"
	"time"

	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
)

func RunPrices(ctx context.Context, config *config.APIServer) error {
	logger := getLogger("PricesApp")

	logger.Sugar().Infof("start PricesApp")

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

	r := gin.New()
	p := ginprom.New(
		ginprom.Engine(r),
		ginprom.Subsystem("gin"),
		ginprom.Path("/metrics"),
	)
	r.Use(
		ginzap.Ginzap(logger, time.RFC3339, true),
		gin.Recovery(),
		p.Instrument(),
	)

	httpSrv := &http.Server{
		Addr:    fmt.Sprintf(":%v", config.Port),
		Handler: r,
	}

	restAPI := api.NewAPI(config, logger, pricesRepo)
	restAPI.RegisterHandlers(r)

	go func() {
		logger.Sugar().Infof("start listening on port=%d", config.Port)
		if err = httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Sugar().Fatalf("can't start api server at port=%d: (%s)", config.Port, err.Error())
		}
	}()

	<-ctx.Done()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := httpSrv.Shutdown(ctx); err != nil {
		logger.Sugar().Fatalf("server shutdown failed: %v\n", err)
	}

	logger.Sugar().Infof("PricesApp stopped. Bye!")

	return nil
}
