// Copyright © 2019 Emando B.V.

package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/emando/vantage-events/internal/follower"
	"github.com/emando/vantage-events/internal/hub"
	"github.com/emando/vantage-events/internal/nats"
	"github.com/emando/vantage-events/pkg/events"
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
			source = nats.NewSource(logger, conn)
		default:
			logger.Fatal("invalid driver")
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		hub := hub.NewServer(logger, source,
			viper.GetString("hub-address"),
			viper.GetString("cert-file"),
			viper.GetString("key-file"),
		)
		go func() {
			if err := hub.ListenAndServeTLS(); err != nil {
				logger.With(zap.Error(err)).Fatal("failed to listen and serve hub")
			}
		}()

		follower := &follower.Follower{
			Logger: logger,
			Source: source,
		}
		competitionCh, err := follower.Run(ctx, viper.GetDuration("history"), viper.GetStringSlice("filter")...)
		if err != nil {
			logger.Fatal("failed to run follower", zap.Error(err))
		}
		go followCompetitions(ctx, competitionCh)

		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt, os.Kill, syscall.SIGTERM)
		<-sigCh
	},
}

func followCompetitions(ctx context.Context, ch <-chan *follower.CompetitionEvents) {
	for {
		select {
		case <-ctx.Done():
			return
		case c, ok := <-ch:
			if !ok {
				return
			}
			logger := logger.With(zap.String("competition_name", c.Competition.Name))
			logger.Info("competition activated")
			go followCompetition(ctx, c)
		}
	}
}

func followCompetition(ctx context.Context, competition *follower.CompetitionEvents) {
	for {
		select {
		case <-ctx.Done():
			return
		case _, ok := <-competition.RawEvents:
			if !ok {
				return
			}
			logger.Debug("received competition event")
		case d, ok := <-competition.DistanceEvents:
			if !ok {
				return
			}
			logger := logger.With(zap.String("distance_name", d.Distance.Name))
			logger.Info("distance activated")
			go followDistance(ctx, d)
		}
	}
}

func followDistance(ctx context.Context, distance *follower.DistanceEvents) {
	for {
		select {
		case <-ctx.Done():
			return
		case _, ok := <-distance.RawEvents:
			if !ok {
				return
			}
			logger.Debug("received distance event")
		case h, ok := <-distance.HeatEvents:
			if !ok {
				return
			}
			logger := logger.With(
				zap.Int("heat_round", h.Heat.Key.Round),
				zap.Int("heat_number", h.Heat.Key.Number),
			)
			logger.Info("heat activated")
			go followHeat(ctx, h)
		}
	}
}

func followHeat(ctx context.Context, heat *follower.HeatEvents) {
	for {
		select {
		case <-ctx.Done():
			return
		case _, ok := <-heat.RawEvents:
			if !ok {
				return
			}
			logger.Debug("received heat event")
		}
	}
}

func init() {
	rootCmd.AddCommand(startCmd)
	startCmd.Flags().Duration("history", 24*time.Hour, "time to seek competition activations")
	startCmd.Flags().StringSlice("filter", nil, "filter competitions by ID")
	startCmd.Flags().String("hub-address", ":443", "hub listen address")
	startCmd.Flags().String("cert-file", "cert.pem", "TLS certificate file")
	startCmd.Flags().String("key-file", "key.pem", "TLS key file")
	viper.BindPFlags(startCmd.Flags())
}
