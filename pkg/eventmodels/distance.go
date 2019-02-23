// Copyright © 2019 Emando B.V.

package eventmodels

import (
	events "github.com/emando/vantage-events"
)

type Distance struct {
	Competition
	DistanceID string `json:"distanceId"`
}

const (
	DistanceActivatedType = "DistanceActivatedEvent"
)

type DistanceActivated struct {
	Distance
	Value events.Distance `json:"distance"`
}
