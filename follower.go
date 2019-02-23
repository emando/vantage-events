// Copyright © 2019 Emando B.V.

package events

import (
	"context"
	"time"

	"go.uber.org/zap"
)

type CompetitionEvents struct {
	Competition *Competition
	Update      chan *Competition
	Distance    chan *DistanceEvents
}

type DistanceEvents struct {
	Distance *Distance
}

type Follower struct {
	Logger *zap.Logger
	Source Source
}

func (f *Follower) Run(ctx context.Context, history time.Duration) (<-chan *CompetitionEvents, error) {
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
					Competition: competition,
					Update:      make(chan *Competition),
					Distance:    make(chan *DistanceEvents),
				}
				events[competition.ID] = ev
				ch <- ev
			}
		}
	}()
	return ch, nil
}
