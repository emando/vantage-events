// Copyright © 2020 Emando B.V.

package hub

import (
	"context"
	"net/http"
	"time"

	"github.com/emando/vantage-events/internal/follower"
	"github.com/emando/vantage-events/pkg/events"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

var upgrader = websocket.Upgrader{}

// Hub is a websocket hub to distribute events to subscribers.
type Hub struct {
	logger *zap.Logger
	source events.Source
	address,
	certFile,
	keyFile string
}

// New instantiates a new Hub.
func New(logger *zap.Logger, source events.Source, address, certFile, keyFile string) *Hub {
	return &Hub{
		logger:   logger,
		source:   source,
		address:  address,
		certFile: certFile,
		keyFile:  keyFile,
	}
}

func (h *Hub) getCompetitions(w http.ResponseWriter, r *http.Request) {
	// TODO: Authenticate via Vantage API.

	ctx := r.Context()
	logger := h.logger.With(zap.String("remote_address", r.RemoteAddr))
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Debug("failed to upgrade websocket", zap.Error(err))
		return
	}
	defer c.Close()

	ch, err := h.source.CompetitionActivations(ctx, 24*time.Hour)
	if err != nil {
		logger.Debug("failed to follow competition activations", zap.Error(err))
		return
	}

	sent := make(map[string]struct{})
	for {
		select {
		case <-ctx.Done():
			return
		case activation := <-ch:
			if _, ok := sent[activation.CompetitionID]; ok {
				continue
			}
			if err := c.WriteMessage(websocket.TextMessage, activation.Raw); err != nil {
				logger.Debug("failed to write message", zap.Error(err))
				return
			}
			sent[activation.CompetitionID] = struct{}{}
		}
	}
}

func (h *Hub) getCompetition(w http.ResponseWriter, r *http.Request) {
	// TODO: Authenticate via Vantage API.

	ctx := r.Context()
	logger := h.logger.With(zap.String("remote_address", r.RemoteAddr))
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Debug("failed to upgrade websocket", zap.Error(err))
		return
	}
	defer c.Close()

	follower := &follower.Follower{
		Logger: logger,
		Source: h.source,
	}
	eventsCh, err := follower.Run(ctx, 24*time.Hour, mux.Vars(r)["id"])
	if err != nil {
		logger.Debug("failed to run follower", zap.Error(err))
		return
	}

	outCh := make(chan []byte)
	go followRawEvents(ctx, eventsCh, outCh)

	for {
		select {
		case <-ctx.Done():
			return
		case buf := <-outCh:
			if err := c.WriteMessage(websocket.TextMessage, buf); err != nil {
				logger.Debug("failed to write message", zap.Error(err))
				return
			}
		}
	}
}

func followRawEvents(ctx context.Context, inCh <-chan *follower.CompetitionEvents, outCh chan<- []byte) {
	for {
		select {
		case <-ctx.Done():
			return
		case competition, ok := <-inCh:
			if !ok {
				return
			}
			outCh <- competition.RawActivation
			go func() {
				for {
					select {
					case <-ctx.Done():
						return
					case event, ok := <-competition.RawEvents:
						if !ok {
							return
						}
						outCh <- event
					case distance, ok := <-competition.DistanceEvents:
						if !ok {
							return
						}
						outCh <- distance.RawActivation
						go func() {
							for {
								select {
								case <-ctx.Done():
									return
								case event, ok := <-competition.RawEvents:
									if !ok {
										return
									}
									outCh <- event
								case heat, ok := <-distance.HeatEvents:
									if !ok {
										return
									}
									outCh <- heat.RawActivation
									go func() {
										for {
											select {
											case <-ctx.Done():
												return
											case event, ok := <-heat.RawEvents:
												if !ok {
													return
												}
												outCh <- event
											}
										}
									}()
								}
							}
						}()
					}
				}
			}()
		}
	}
}

// ListenAndServeTLS starts the websocket hub.
func (h *Hub) ListenAndServeTLS() error {
	r := mux.NewRouter()
	r.HandleFunc("/v1/competitions", h.getCompetitions)
	r.HandleFunc("/v1/competitions/{id}", h.getCompetition)
	return http.ListenAndServeTLS(h.address, h.certFile, h.keyFile, r)
}
