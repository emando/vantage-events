// Copyright © 2019 Emando B.V.

package events

import (
	"context"
	"time"
)

// Source is a source for competition events.
type Source interface {
	CompetitionActivations(ctx context.Context, history time.Duration) (<-chan *Competition, error)
	DistanceActivations(ctx context.Context, competitionID string) (<-chan *Distance, error)
	HeatActivations(ctx context.Context, competitionID, distanceID string) (<-chan *Heat, error)
}
