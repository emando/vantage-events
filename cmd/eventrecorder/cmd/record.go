// Copyright © 2020 Emando B.V.

package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/emando/vantage-events/internal/hub"
	"github.com/emando/vantage-events/pkg/events"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// recordCmd represents the record command.
var recordCmd = &cobra.Command{
	Use:   "record",
	Short: "Record events.",
	Run: func(cmd *cobra.Command, args []string) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		url := fmt.Sprintf("wss://%s/v1/competitions", viper.GetString("host"))
		if id := viper.GetString("competition"); id != "" {
			url += "/" + id
		}
		client, err := hub.Connect(ctx, url)
		if err != nil {
			logger.Fatal("failed to connect client", zap.Error(err))
		}

		ch := make(chan *events.Raw, 16)
		go func() {
			defer close(ch)
			if err := client.Read(ctx, ch); err != nil {
				if err != context.Canceled {
					logger.Fatal("failed to read messages", zap.Error(err))
				}
				return
			}
		}()
		go func() {
			file, err := os.OpenFile(viper.GetString("file"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				logger.Fatal("failed to open file for writing",
					zap.String("file", viper.GetString("file")),
					zap.Error(err),
				)
			}
			defer file.Close()
			for {
				select {
				case <-ctx.Done():
					return
				case event, ok := <-ch:
					if !ok {
						return
					}
					now := time.Now()
					m := make(map[string]interface{})
					if err := json.Unmarshal(event.Bytes, &m); err != nil {
						logger.Fatal("failed to unmarshal event", zap.Error(err))
					}
					m["_time"] = now
					buf, err := json.Marshal(m)
					if err != nil {
						logger.Fatal("failed to marshal event", zap.Error(err))
					}
					buf = append(buf, '\n')
					if _, err := file.Write(buf); err != nil {
						logger.Fatal("failed to write to file", zap.Error(err))
					}
					logger.Info("wrote to file", zap.Time("time", now), zap.String("type", event.TypeName()))
				}
			}
		}()

		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt, os.Kill, syscall.SIGTERM)
		<-sigCh
	},
}

func init() {
	rootCmd.AddCommand(recordCmd)
	recordCmd.Flags().String("competition", "", "competition ID to record")
	viper.BindPFlags(recordCmd.Flags())
}
