// Copyright © 2019 Emando B.V.

package events

import (
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
)

// HeatActivated is the event data of a Vantage competition activation.
type HeatActivated struct {
	Heat
	Raw []byte `json:"-"`
}
