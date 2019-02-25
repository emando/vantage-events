// Copyright © 2019 Emando B.V.

package nats

import (
	"context"
	"fmt"
	"time"

	events "github.com/emando/vantage-events"
	"github.com/emando/vantage-events/pkg/eventmodels"
	stan "github.com/nats-io/go-nats-streaming"
	"go.uber.org/zap"
)

// Source is a NATS seeker for competition events.
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
	heatActivations        = "competition.%v.distances.%v.heats.activations.*"
)

// CompetitionActivations returns the competition activations.
func (s *Source) CompetitionActivations(ctx context.Context, history time.Duration) (<-chan *events.Competition, error) {
	ch := make(chan *events.Competition)
	cb := func(msg *stan.Msg) {
		event := new(eventmodels.CompetitionActivated)
		if err := eventmodels.Unmarshal(msg.Data, eventmodels.CompetitionActivatedType, event); err != nil {
			s.logger.Warn("failed to unmarshal data", zap.Error(err))
			return
		}
		s.logger.Debug("received competition activation", zap.String("id", event.CompetitionID))
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
func (s *Source) DistanceActivations(ctx context.Context, competitionID string) (<-chan *events.Distance, error) {
	ch := make(chan *events.Distance)
	cb := func(msg *stan.Msg) {
		event := new(eventmodels.DistanceActivated)
		if err := eventmodels.Unmarshal(msg.Data, eventmodels.DistanceActivatedType, event); err != nil {
			s.logger.Warn("failed to unmarshal data", zap.Error(err))
			return
		}
		s.logger.Debug("received distance activation",
			zap.String("competition_id", competitionID),
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
		s.logger.Debug("unsubscribe from distance activations", zap.String("competition_id", competitionID))
		sub.Close()
	}()
	return ch, nil
}

// HeatActivations returns the heat activations. The last activated heat is always returned.
func (s *Source) HeatActivations(ctx context.Context, competitionID, distanceID string) (<-chan *events.Heat, error) {
	ch := make(chan *events.Heat)
	cb := func(msg *stan.Msg) {
		event := new(eventmodels.HeatActivated)
		if err := eventmodels.Unmarshal(msg.Data, eventmodels.HeatActivatedType, event); err != nil {
			s.logger.Warn("failed to unmarshal data", zap.Error(err))
			return
		}
		s.logger.Debug("received heat activation",
			zap.String("competition_id", competitionID),
			zap.String("distance_id", distanceID),
			zap.Int("round", event.Key.Round),
			zap.Int("heat", event.Key.Number),
		)
		ch <- &event.Heat.Heat
	}
	subject := fmt.Sprintf(heatActivations, competitionID, distanceID)
	sub, err := s.conn.stan.Subscribe(subject, cb, stan.StartWithLastReceived())
	if err != nil {
		return nil, err
	}
	go func() {
		<-ctx.Done()
		s.logger.Debug("unsubscribe from heat activations",
			zap.String("competition_id", competitionID),
			zap.String("distance_id", distanceID),
		)
		sub.Close()
	}()
	return ch, nil
}
