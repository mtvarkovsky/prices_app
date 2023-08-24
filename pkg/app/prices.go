package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"prices/pkg/api"
	"prices/pkg/config"
	"prices/pkg/migrations"
	"prices/pkg/repository"
	"time"

	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func RunPrices(ctx context.Context, config *config.APIServer) error {
	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.TimeKey = "timestamp"
	encoderCfg.EncodeTime = zapcore.RFC3339TimeEncoder

	zapCfg := zap.NewProductionConfig()
	zapCfg.EncoderConfig = encoderCfg

	logger := zap.Must(zapCfg.Build())

	log := logger.Named("PricesApp")

	log.Sugar().Infof("start PricesApp")

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

	r := gin.New()

	r.Use(
		ginzap.Ginzap(logger, time.RFC3339, true),
		gin.Recovery(),
	)

	httpSrv := &http.Server{
		Addr:    fmt.Sprintf(":%v", config.Port),
		Handler: r,
	}

	restAPI := api.NewAPI(config, log, pricesRepo)
	restAPI.RegisterHandlers(r)

	go func() {
		log.Sugar().Infof("start listening on port=%d", config.Port)
		if err = httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Sugar().Fatalf("can't start api server at port=%d: (%s)", config.Port, err.Error())
		}
	}()

	<-ctx.Done()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := httpSrv.Shutdown(ctx); err != nil {
		log.Sugar().Fatalf("server shutdown failed: %v\n", err)
	}

	log.Sugar().Infof("PricesApp stopped. Bye!")

	return nil
}
