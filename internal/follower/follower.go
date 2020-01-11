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
				close(ch)
				return
			case competition := <-activations:
				if cancel, ok := competitions[competition.ID]; ok {
					cancel()
				}
				ctx, cancel := context.WithCancel(ctx)
				defer cancel()
				competitions[competition.ID] = cancel
				ev := &CompetitionEvents{
					source: f.Source,
					logger: f.Logger.With(
						zap.String("competition_id", competition.ID),
						zap.String("competition_name", competition.Name),
					),
					Competition: competition,
					Update:      make(chan *entities.Competition),
					Distance:    make(chan *DistanceEvents),
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

	Competition *entities.Competition
	Update      chan *entities.Competition
	Distance    chan *DistanceEvents
}

func (c *CompetitionEvents) follow(ctx context.Context) error {
	activations, err := c.source.DistanceActivations(ctx, c.Competition.ID)
	if err != nil {
		return err
	}
	for {
		select {
		case <-ctx.Done():
			close(c.Distance)
			return ctx.Err()
		case distance := <-activations:
			ev := &DistanceEvents{
				source: c.source,
				logger: c.logger.With(
					zap.String("distance_id", distance.ID),
					zap.String("distance_name", distance.Name),
				),
				Competition: c.Competition,
				Distance:    distance,
				Heats:       make(chan *HeatEvents),
			}
			c.Distance <- ev
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
	Heats       chan *HeatEvents
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
			close(d.Heats)
			return ctx.Err()
		case heat := <-activations:
			ev := &HeatEvents{
				source: d.source,
				logger: d.logger.With(
					zap.Int("heat_round", heat.Key.Round),
					zap.Int("heat_number", heat.Key.Number),
				),
				Competition: d.Competition,
				Distance:    d.Distance,
				Heat:        heat,
			}
			d.Heats <- ev
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
}
