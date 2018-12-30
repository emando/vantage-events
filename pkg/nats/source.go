// Copyright © 2018 Emando B.V.

package nats

import (
	"context"
	"time"

	events "github.com/johanstokking/vantage-events"
	"github.com/johanstokking/vantage-events/pkg/eventmodels"
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
)

// Competitions returns the competition activations.
func (s *Source) Competitions(ctx context.Context, history time.Duration) (<-chan *events.Competition, error) {
	ch := make(chan *events.Competition)
	cb := func(msg *stan.Msg) {
		s.logger.Debug("received message", zap.String("subject", competitionActivations))
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
		sub.Close()
		close(ch)
	}()
	return ch, nil
}
