// Copyright © 2019 Emando B.V.

package nats

import (
	"context"
	"fmt"
	"time"

	"github.com/emando/vantage-events/pkg/entities"
	"github.com/emando/vantage-events/pkg/events"
	stan "github.com/nats-io/stan.go"
	"go.uber.org/zap"
)

// Source is a NATS seeker for competition entities.
type Source struct {
	logger *zap.Logger
	conn   *Conn
}

// NewSource returns a new NATS Streaming Server seeker.
func NewSource(logger *zap.Logger, conn *Conn) *Source {
	return &Source{
		logger: logger,
		conn:   conn,
	}
}

const (
	competitionActivations = "competition.activations"
	distanceActivations    = "competition.%v.distances.activations"
	heatActivations        = "competition.%v.distances.%v.heats.activations.%d"
)

// CompetitionActivations returns the competition activations.
func (s *Source) CompetitionActivations(ctx context.Context, history time.Duration) (<-chan *entities.Competition, error) {
	ch := make(chan *entities.Competition)
	cb := func(msg *stan.Msg) {
		event := new(events.CompetitionActivated)
		if err := events.Unmarshal(msg.Data, events.CompetitionActivatedType, event); err != nil {
			s.logger.Warn("failed to unmarshal data", zap.Error(err))
			return
		}
		s.logger.Debug("received competition activation",
			zap.String("competition_id", event.CompetitionID),
		)
		ch <- &event.Value
	}
	sub, err := s.conn.stan.Subscribe(competitionActivations, cb, stan.StartAtTimeDelta(history))
	if err != nil {
		return nil, err
	}
	go func() {
		<-ctx.Done()
		s.logger.Debug("unsubscribe from competition activations")
		sub.Close()
	}()
	return ch, nil
}

// DistanceActivations returns the distance activations. The last activated distance is always returned.
func (s *Source) DistanceActivations(ctx context.Context, competitionID string) (<-chan *entities.Distance, error) {
	logger := s.logger.With(
		zap.String("competition_id", competitionID),
	)
	ch := make(chan *entities.Distance)
	cb := func(msg *stan.Msg) {
		event := new(events.DistanceActivated)
		if err := events.Unmarshal(msg.Data, events.DistanceActivatedType, event); err != nil {
			logger.Warn("failed to unmarshal data", zap.Error(err))
			return
		}
		logger.Debug("received distance activation",
			zap.String("distance_id", event.DistanceID),
		)
		ch <- &event.Value
	}
	subject := fmt.Sprintf(distanceActivations, competitionID)
	sub, err := s.conn.stan.Subscribe(subject, cb, stan.StartWithLastReceived())
	if err != nil {
		return nil, err
	}
	go func() {
		<-ctx.Done()
		logger.Debug("unsubscribe from distance activations")
		sub.Close()
	}()
	return ch, nil
}

// HeatActivations returns the heat activations. The last activated heat is always returned.
func (s *Source) HeatActivations(ctx context.Context, competitionID, distanceID string, groups ...int) (<-chan *entities.Heat, error) {
	logger := s.logger.With(
		zap.String("competition_id", competitionID),
		zap.String("distance_id", distanceID),
	)
	ch := make(chan *entities.Heat)
	cb := func(msg *stan.Msg) {
		event := new(events.HeatActivated)
		if err := events.Unmarshal(msg.Data, events.HeatActivatedType, event); err != nil {
			logger.Warn("failed to unmarshal data", zap.Error(err))
			return
		}
		logger.Debug("received heat activation",
			zap.Int("heat_round", event.Key.Round),
			zap.Int("heat_number", event.Key.Number),
		)
		ch <- &event.Heat.Heat
	}
	for _, group := range groups {
		subject := fmt.Sprintf(heatActivations, competitionID, distanceID, group)
		sub, err := s.conn.stan.Subscribe(subject, cb, stan.StartWithLastReceived())
		if err != nil {
			return nil, err
		}
		go func() {
			<-ctx.Done()
			logger.Debug("unsubscribe from heat activations")
			sub.Close()
		}()
	}
	return ch, nil
}
