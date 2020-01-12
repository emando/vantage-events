// Copyright © 2020 Emando B.V.

package cmd

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/websocket"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// replayCmd represents the replay command.
var replayCmd = &cobra.Command{
	Use:   "replay",
	Short: "Replay events.",
	Run: func(cmd *cobra.Command, args []string) {
		logger.Info("starting server", zap.String("address", viper.GetString("address")))
		http.HandleFunc("/", handle)
		if err := http.ListenAndServe(viper.GetString("address"), nil); err != nil {
			logger.Fatal("failed to listen", zap.Error(err))
		}
	},
}

var upgrader = websocket.Upgrader{}

func handle(w http.ResponseWriter, r *http.Request) {
	logger := logger.With(zap.String("remote_address", r.RemoteAddr))
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Debug("failed to upgrade websocket", zap.Error(err))
		return
	}
	defer conn.Close()

	logger.Info("client connected")
	defer logger.Info("client disconnected")

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	go func() {
		defer cancel()
		file, err := os.Open(viper.GetString("file"))
		if err != nil {
			logger.Fatal("failed to open file for reading",
				zap.String("file", viper.GetString("file")),
				zap.Error(err),
			)
		}
		reader := bufio.NewReader(file)
		var last *time.Time
		for {
			line, err := reader.ReadString('\n')
			if err == io.EOF {
				return
			} else if err != nil {
				logger.Error("failed to read from file", zap.Error(err))
				return
			}
			line = line[:len(line)-1]
			event := struct {
				Time     time.Time `json:"_time"`
				TypeName string    `json:"typeName"`
			}{}
			if err := json.Unmarshal([]byte(line), &event); err != nil {
				logger.Error("failed to unmarshal event", zap.Error(err))
				return
			}
			var wait time.Duration
			if last != nil {
				wait = event.Time.Sub(*last)
				wait = wait / time.Duration(viper.GetInt("speed"))
				logger.Info("waiting to send next message", zap.Duration("time", wait))
			}
			last = &event.Time
			select {
			case <-ctx.Done():
				return
			case <-time.After(wait):
			}
			logger.Info("writing message", zap.String("type", event.TypeName))
			if err := conn.WriteMessage(websocket.TextMessage, []byte(line)); err != nil {
				logger.Error("failed to write message", zap.Error(err))
				return
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		_, _, err := conn.ReadMessage()
		if err != nil {
			logger.Debug("failed to read message", zap.Error(err))
			return
		}
	}
}

func init() {
	rootCmd.AddCommand(replayCmd)
	replayCmd.Flags().String("address", ":3000", "listen address")
	replayCmd.Flags().Int("speed", 1, "reduce wait time between events by this factor")
	viper.BindPFlags(replayCmd.Flags())
}
