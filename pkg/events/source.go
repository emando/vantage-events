// Copyright © 2019 Emando B.V.

package events

import (
	"context"
	"time"
)

// Source is a source for competition events.
type Source interface {
	CompetitionActivations(ctx context.Context, history time.Duration) (<-chan *CompetitionActivated, error)
	CompetitionEvents(ctx context.Context, since *CompetitionActivated) (<-chan *Raw, error)
	DistanceActivations(ctx context.Context, competitionID string) (<-chan *DistanceActivated, error)
	DistanceEvents(ctx context.Context, since *DistanceActivated) (<-chan *Raw, error)
	HeatActivations(ctx context.Context, competitionID, distanceID string, groups ...int) (<-chan *HeatActivated, error)
	HeatEvents(ctx context.Context, since *HeatActivated) (<-chan *Raw, error)
}
