// Copyright © 2018 Emando B.V.

package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	events "github.com/johanstokking/vantage-events"
	"github.com/johanstokking/vantage-events/pkg/nats"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// startCmd represents the start command.
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the Event Aggregator.",
	Run: func(cmd *cobra.Command, args []string) {
		var source events.Source
		switch viper.GetString("driver") {
		case "nats":
			opts := nats.Options{
				URL:       viper.GetString("nats-url"),
				Username:  viper.GetString("nats-username"),
				Password:  viper.GetString("nats-password"),
				UseTLS:    viper.GetBool("nats-tls"),
				ClusterID: viper.GetString("nats-cluster-id"),
				ClientID:  viper.GetString("nats-client-id"),
			}
			conn, err := nats.Connect(opts)
			if err != nil {
				logger.Fatal("failed to connect to NATS", zap.Error(err))
			}
			defer conn.Close()
			source = nats.NewSource(logger, conn)
		default:
			logger.Fatal("invalid driver")
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		competitionCh, err := source.Competitions(ctx, viper.GetDuration("history"))
		if err != nil {
			logger.Fatal("failed to get competitions", zap.Error(err))
		}
		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				case c := <-competitionCh:
					logger.Info("competition activated", zap.String("name", c.Name))
				}
			}
		}()

		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt, os.Kill, syscall.SIGTERM)
		<-sigCh
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
	startCmd.Flags().Duration("history", 24*time.Hour, "time to seek competition activations")
	viper.BindPFlags(startCmd.Flags())
}
