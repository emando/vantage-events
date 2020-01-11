// Copyright © 2019 Emando B.V.

package follower

import (
	"context"
	"strings"
	"time"

	"github.com/emando/vantage-events/pkg/entities"
	"github.com/emando/vantage-events/pkg/events"
	"go.uber.org/zap"
)

// Follower follows Vantage events.
type Follower struct {
	Logger *zap.Logger
	Source events.Source
}

// Run starts the follower.
func (f Follower) Run(ctx context.Context, history time.Duration) (<-chan *CompetitionEvents, error) {
	activations, err := f.Source.CompetitionActivations(ctx, history)
	if err != nil {
		return nil, err
	}
	ch := make(chan *CompetitionEvents)
	go func() {
		competitions := make(map[string]context.CancelFunc)
		for {
			select {
			case <-ctx.Done():
				return
			case activation := <-activations:
				if cancel, ok := competitions[activation.CompetitionID]; ok {
					cancel()
				}
				ctx, cancel := context.WithCancel(ctx)
				defer cancel()
				competitions[activation.CompetitionID] = cancel
				ev := &CompetitionEvents{
					source: f.Source,
					logger: f.Logger.With(
						zap.String("competition_id", activation.CompetitionID),
						zap.String("competition_name", activation.Value.Name),
					),
					Competition:    &activation.Value,
					DistanceEvents: make(chan *DistanceEvents),
					RawActivation:  activation.Raw,
					RawEvents:      make(chan []byte),
				}
				ch <- ev
				go func() {
					if err := ev.follow(ctx); err != nil && err != context.Canceled {
						f.Logger.Error("failed to follow competition", zap.Error(err))
					}
				}()
			}
		}
	}()
	return ch, nil
}

// CompetitionEvents provides Vantage events from a competition.
type CompetitionEvents struct {
	source events.Source
	logger *zap.Logger

	Competition    *entities.Competition
	DistanceEvents chan *DistanceEvents

	RawActivation []byte
	RawEvents     chan []byte
}

func (c *CompetitionEvents) follow(ctx context.Context) error {
	activations, err := c.source.DistanceActivations(ctx, c.Competition.ID)
	if err != nil {
		return err
	}
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case activation := <-activations:
			ev := &DistanceEvents{
				source: c.source,
				logger: c.logger.With(
					zap.String("distance_id", activation.DistanceID),
					zap.String("distance_name", activation.Value.Name),
				),
				Competition:   c.Competition,
				Distance:      &activation.Value,
				HeatEvents:    make(chan *HeatEvents),
				RawActivation: activation.Raw,
				RawEvents:     make(chan []byte),
			}
			c.DistanceEvents <- ev
			go func() {
				if err := ev.follow(ctx); err != nil && err != context.Canceled {
					c.logger.Error("failed to follow distance", zap.Error(err))
				}
			}()
		}
	}
}

// DistanceEvents provides Vantage competition distance events.
type DistanceEvents struct {
	source events.Source
	logger *zap.Logger

	Competition *entities.Competition
	Distance    *entities.Distance
	HeatEvents  chan *HeatEvents

	RawActivation []byte
	RawEvents     chan []byte
}

func (d *DistanceEvents) follow(ctx context.Context) error {
	groups := make([]int, 1, 2)
	switch {
	case strings.HasPrefix(d.Distance.Discipline, "SpeedSkating.LongTrack.PairsDistance."):
		// Subscribe to both pairs if the start mode is not SingleHeat.
		if d.Distance.StartMode != 0 {
			groups = append(groups, 1)
		}
	}
	activations, err := d.source.HeatActivations(ctx, d.Competition.ID, d.Distance.ID, groups...)
	if err != nil {
		return err
	}
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case activation := <-activations:
			ev := &HeatEvents{
				source: d.source,
				logger: d.logger.With(
					zap.Int("heat_round", activation.Key.Round),
					zap.Int("heat_number", activation.Key.Number),
				),
				Competition:   d.Competition,
				Distance:      d.Distance,
				Heat:          &activation.Heat.Heat,
				RawActivation: activation.Raw,
				RawEvents:     make(chan []byte),
			}
			d.HeatEvents <- ev
		}
	}
}

// HeatEvents provides Vantage competition distance heat events.
type HeatEvents struct {
	source events.Source
	logger *zap.Logger

	Competition *entities.Competition
	Distance    *entities.Distance
	Heat        *entities.Heat

	RawActivation []byte
	RawEvents     chan []byte
}
