// Copyright © 2019 Emando B.V.

package nats

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

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
	competitionEvents      = "competition.%v"
	distanceActivations    = "competition.%v.distances.activations"
	distanceEvents         = "competition.%v.distances.%v"
	heatActivations        = "competition.%v.distances.%v.heats.activations.%d"
	heatEvents             = "competition.%v.distances.%v.heats.%d.%d"
)

// CompetitionActivations returns the competition activations.
func (s *Source) CompetitionActivations(ctx context.Context, history time.Duration) (<-chan *events.CompetitionActivated, error) {
	ch := make(chan *events.CompetitionActivated)
	cb := func(msg *stan.Msg) {
		event := &events.CompetitionActivated{
			Time: time.Unix(0, msg.Timestamp),
			Raw:  append(msg.Data[:0:0], msg.Data...),
		}
		if err := events.Unmarshal(msg.Data, events.CompetitionActivatedType, event); err != nil {
			s.logger.Warn("failed to unmarshal data", zap.Error(err))
			return
		}
		s.logger.Debug("received competition activation",
			zap.String("competition_id", event.CompetitionID),
		)
		ch <- event
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

// CompetitionEvents returns the competition events.
func (s *Source) CompetitionEvents(ctx context.Context, since *events.CompetitionActivated) (<-chan *events.Raw, error) {
	logger := s.logger.With(zap.String("competition_id", since.CompetitionID))
	ch := make(chan *events.Raw)
	cb := func(msg *stan.Msg) {
		logger.Debug("received competition event")
		event := &events.Raw{
			Bytes: append(msg.Data[:0:0], msg.Data...),
		}
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			logger.Warn("failed to unmarshal data", zap.Error(err))
			return
		}
		ch <- event
	}
	subject := fmt.Sprintf(competitionEvents, since.CompetitionID)
	sub, err := s.conn.stan.Subscribe(subject, cb, stan.StartAtTime(since.Time))
	if err != nil {
		return nil, err
	}
	go func() {
		<-ctx.Done()
		logger.Debug("unsubscribe from competition events")
		sub.Close()
	}()
	return ch, nil
}

// DistanceActivations returns the distance activations. The last activated distance is always returned.
func (s *Source) DistanceActivations(ctx context.Context, competitionID string) (<-chan *events.DistanceActivated, error) {
	logger := s.logger.With(
		zap.String("competition_id", competitionID),
	)
	ch := make(chan *events.DistanceActivated)
	cb := func(msg *stan.Msg) {
		event := &events.DistanceActivated{
			Time: time.Unix(0, msg.Timestamp),
			Raw:  append(msg.Data[:0:0], msg.Data...),
		}
		if err := events.Unmarshal(msg.Data, events.DistanceActivatedType, event); err != nil {
			logger.Warn("failed to unmarshal data", zap.Error(err))
			return
		}
		logger.Debug("received distance activation",
			zap.String("distance_id", event.DistanceID),
		)
		ch <- event
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

// DistanceEvents returns the competition distance events.
func (s *Source) DistanceEvents(ctx context.Context, since *events.DistanceActivated) (<-chan *events.Raw, error) {
	logger := s.logger.With(
		zap.String("competition_id", since.CompetitionID),
		zap.String("distance_id", since.DistanceID),
	)
	ch := make(chan *events.Raw)
	cb := func(msg *stan.Msg) {
		logger.Debug("received distance event")
		event := &events.Raw{
			Bytes: append(msg.Data[:0:0], msg.Data...),
		}
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			logger.Warn("failed to unmarshal data", zap.Error(err))
			return
		}
		ch <- event
	}
	subject := fmt.Sprintf(distanceEvents, since.CompetitionID, since.DistanceID)
	sub, err := s.conn.stan.Subscribe(subject, cb, stan.StartAtTime(since.Time))
	if err != nil {
		return nil, err
	}
	go func() {
		<-ctx.Done()
		logger.Debug("unsubscribe from distance events")
		sub.Close()
	}()
	return ch, nil
}

// HeatActivations returns the heat activations. The last activated heat is always returned.
func (s *Source) HeatActivations(ctx context.Context, competitionID, distanceID string, groups ...int) (<-chan *events.HeatActivated, error) {
	logger := s.logger.With(
		zap.String("competition_id", competitionID),
		zap.String("distance_id", distanceID),
	)
	ch := make(chan *events.HeatActivated)
	cb := func(msg *stan.Msg) {
		event := &events.HeatActivated{
			Time: time.Unix(0, msg.Timestamp),
			Raw:  append(msg.Data[:0:0], msg.Data...),
		}
		if err := events.Unmarshal(msg.Data, events.HeatActivatedType, event); err != nil {
			logger.Warn("failed to unmarshal data", zap.Error(err))
			return
		}
		logger.Debug("received heat activation",
			zap.Int("heat_round", event.Key.Round),
			zap.Int("heat_number", event.Key.Number),
		)
		ch <- event
	}
	for _, group := range groups {
		logger := logger.With(zap.Int("group", group))
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

// HeatEvents returns the competititon distance heat events.
func (s *Source) HeatEvents(ctx context.Context, since *events.HeatActivated) (<-chan *events.Raw, error) {
	logger := s.logger.With(
		zap.String("competition_id", since.CompetitionID),
		zap.String("distance_id", since.DistanceID),
		zap.Int("heat_round", since.Key.Round),
		zap.Int("heat_number", since.Key.Number),
	)
	ch := make(chan *events.Raw)
	cb := func(msg *stan.Msg) {
		logger.Debug("received heat event")
		event := &events.Raw{
			Bytes: append(msg.Data[:0:0], msg.Data...),
		}
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			logger.Warn("failed to unmarshal data", zap.Error(err))
			return
		}
		ch <- event
	}
	subject := fmt.Sprintf(heatEvents, since.CompetitionID, since.DistanceID, since.Key.Round, since.Key.Number)
	sub, err := s.conn.stan.Subscribe(subject, cb, stan.StartAtTime(since.Time))
	if err != nil {
		return nil, err
	}
	go func() {
		<-ctx.Done()
		logger.Debug("unsubscribe from heat events")
		sub.Close()
	}()
	return ch, nil
}
