// Copyright © 2019 Emando B.V.

package eventmodels

import (
	events "github.com/emando/vantage-events"
)

type Heat struct {
	Distance
	events.Heat
}

const (
	HeatActivatedType = "HeatActivatedEvent"
)

type HeatActivated struct {
	Heat
}
