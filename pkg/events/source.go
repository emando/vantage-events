// Copyright © 2019 Emando B.V.

package events

import (
	"context"
	"time"

	"github.com/emando/vantage-events/pkg/entities"
)

// Source is a source for competition events.
type Source interface {
	CompetitionActivations(ctx context.Context, history time.Duration) (<-chan *entities.Competition, error)
	DistanceActivations(ctx context.Context, competitionID string) (<-chan *entities.Distance, error)
	HeatActivations(ctx context.Context, competitionID, distanceID string) (<-chan *entities.Heat, error)
}
