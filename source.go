// Copyright © 2018 Emando B.V.

package events

import (
	"context"
	"time"
)

// Source is a source for competition events.
type Source interface {
	Competitions(ctx context.Context, history time.Duration) (<-chan *Competition, error)
}
