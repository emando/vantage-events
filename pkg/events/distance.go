// Copyright © 2019 Emando B.V.

package events

import (
	"github.com/emando/vantage-events/pkg/entities"
)

// Distance is a Vantage competition distance event.
type Distance struct {
	Competition
	DistanceID string `json:"distanceId"`
}

const (
	// DistanceActivatedType is the event name of a Vantage competition distance activated event.
	DistanceActivatedType = "DistanceActivatedEvent"
)

// DistanceActivated is the event data of a Vantage competition distance activation.
type DistanceActivated struct {
	Distance
	Value entities.Distance `json:"distance"`
}
