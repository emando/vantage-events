// Copyright © 2019 Emando B.V.

package events

import (
	"context"
	"time"
)

// Source is a source for competition events.
type Source interface {
	CompetitionActivations(ctx context.Context, history time.Duration) (<-chan *CompetitionActivated, error)
	DistanceActivations(ctx context.Context, competitionID string) (<-chan *DistanceActivated, error)
	HeatActivations(ctx context.Context, competitionID, distanceID string, groups ...int) (<-chan *HeatActivated, error)
}
