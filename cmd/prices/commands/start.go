package commands

import (
	"context"
	"github.com/spf13/cobra"
	"os"
	"os/signal"
	"prices/pkg/app"
	"prices/pkg/config"
	"syscall"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the application",
	RunE: func(c *cobra.Command, args []string) error {
		ctx, closer := context.WithCancel(context.Background())
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)

		go func() {
			select {
			case <-ch:
				closer()
			case <-ctx.Done():
			}
		}()

		cfg := &config.APIServer{}
		cfg, err := cfg.LoadConfig("prices_app.yaml")
		if err != nil {
			return err
		}

		return app.RunPrices(ctx, cfg)
	},
}
