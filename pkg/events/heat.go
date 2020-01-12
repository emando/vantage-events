// Copyright © 2019 Emando B.V.

package events

import (
	"time"

	"github.com/emando/vantage-events/pkg/entities"
)

// Heat is a Vantage competition distance heat event.
type Heat struct {
	Distance
	entities.Heat
}

const (
	// HeatActivatedType is the event name of a Vantage competition distance heat activation.
	HeatActivatedType = "HeatActivatedEvent"
	// HeatDeactivatedType is the event name of a Vantage competition distance heat deactivation.
	HeatDeactivatedType = "HeatDeactivatedEvent"
)

// HeatActivated is the event data of a Vantage competition activation.
type HeatActivated struct {
	Heat
	Time time.Time `json:"-"`
	Raw  []byte    `json:"-"`
}
