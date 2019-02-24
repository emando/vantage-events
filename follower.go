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
		competitions := make(map[string]context.CancelFunc)
		for {
			select {
			case <-ctx.Done():
				return
			case competition := <-activations:
				if cancel, ok := competitions[competition.ID]; ok {
					cancel()
				}
				ctx, cancel := context.WithCancel(ctx)
				competitions[competition.ID] = cancel
				ev := &CompetitionEvents{
					Follower:    f,
					Competition: competition,
					Update:      make(chan *Competition),
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
				Follower:    c.Follower,
				Competition: c.Competition,
				Distance:    distance,
				Heats:       make(chan *HeatEvents),
			}
			c.Distance <- ev
			go func() {
				if err := ev.follow(ctx); err != nil && err != context.Canceled {
					c.Logger.Error("failed to follow distance", zap.Error(err))
				}
			}()
		}
	}
}

type DistanceEvents struct {
	Follower
	Competition *Competition
	Distance    *Distance
	Heats       chan *HeatEvents
}

func (d *DistanceEvents) follow(ctx context.Context) error {
	activations, err := d.Source.HeatActivations(ctx, d.Competition.ID, d.Distance.ID)
	if err != nil {
		return err
	}
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case heat := <-activations:
			ev := &HeatEvents{
				Follower:    d.Follower,
				Competition: d.Competition,
				Distance:    d.Distance,
				Heat:        heat,
			}
			d.Heats <- ev
		}
	}
}

type HeatEvents struct {
	Follower
	Competition *Competition
	Distance    *Distance
	Heat        *Heat
}
