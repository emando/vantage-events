// Copyright © 2019 Emando B.V.

package events

import (
	"context"
	"time"

	"go.uber.org/zap"
)

type Follower struct {
	Logger *zap.Logger
	Source Source
}

func (f Follower) Run(ctx context.Context, history time.Duration) (<-chan *CompetitionEvents, error) {
	activations, err := f.Source.CompetitionActivations(ctx, history)
	if err != nil {
		return nil, err
	}
	ch := make(chan *CompetitionEvents)
	go func() {
		events := make(map[string]*CompetitionEvents)
		for {
			select {
			case <-ctx.Done():
				return
			case competition := <-activations:
				if ev, ok := events[competition.ID]; ok {
					ev.Competition = competition
					ev.Update <- competition
					continue
				}
				ev := &CompetitionEvents{
					Follower:    f,
					Competition: competition,
					Update:      make(chan *Competition),
					Distance:    make(chan *DistanceEvents),
				}
				events[competition.ID] = ev
				ch <- ev
				go func() {
					if err := ev.follow(ctx); err != nil {
						f.Logger.Error("failed to follow competition", zap.Error(err))
					}
				}()
			}
		}
	}()
	return ch, nil
}

type CompetitionEvents struct {
	Follower
	Competition *Competition
	Update      chan *Competition
	Distance    chan *DistanceEvents
}

func (c *CompetitionEvents) follow(ctx context.Context) error {
	activations, err := c.Source.DistanceActivations(ctx, c.Competition.ID)
	if err != nil {
		return err
	}
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case distance := <-activations:
			ev := &DistanceEvents{
				Follower: c.Follower,
				Distance: distance,
			}
			c.Distance <- ev
		}
	}
}

type DistanceEvents struct {
	Follower
	Distance *Distance
	Heats    chan *HeatEvents
}

type HeatEvents struct {
	Follower
}
