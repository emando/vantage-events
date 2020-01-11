// Copyright © 2019 Emando B.V.

package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/emando/vantage-events/internal/follower"
	"github.com/emando/vantage-events/internal/nats"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// startCmd represents the start command.
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the Event Aggregator.",
	Run: func(cmd *cobra.Command, args []string) {
		follower := &follower.Follower{
			Logger: logger,
		}
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
			logger.With(
				zap.String("url", opts.URL),
				zap.String("username", opts.Username),
				zap.Bool("tls", opts.UseTLS),
				zap.String("cluster_id", opts.ClusterID),
				zap.String("client_id", opts.ClientID),
			).Info("connecting NATS...")
			conn, err := nats.Connect(opts)
			if err != nil {
				logger.Fatal("failed to connect to NATS", zap.Error(err))
			}
			defer conn.Close()
			follower.Source = nats.NewSource(logger, conn)
		default:
			logger.Fatal("invalid driver")
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		competitionCh, err := follower.Run(ctx, viper.GetDuration("history"))
		if err != nil {
			logger.Fatal("failed to get competitions", zap.Error(err))
		}
		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				case c := <-competitionCh:
					logger := logger.With(zap.String("competition_name", c.Competition.Name))
					logger.Info("competition activated")
					go func() {
						for {
							select {
							case <-ctx.Done():
								return
							case <-c.Update:
							case d := <-c.Distance:
								logger := logger.With(zap.String("distance_name", d.Distance.Name))
								logger.Info("distance activated")
								go func() {
									for {
										select {
										case <-ctx.Done():
											return
										case h := <-d.Heats:
											logger := logger.With(
												zap.Int("heat_round", h.Heat.Key.Round),
												zap.Int("heat_number", h.Heat.Key.Number),
											)
											logger.Info("heat activated")
										}
									}
								}()
							}
						}
					}()
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
