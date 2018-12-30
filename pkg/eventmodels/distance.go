// Copyright © 2018 Emando B.V.

package eventmodels

import (
	events "github.com/johanstokking/vantage-events"
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
