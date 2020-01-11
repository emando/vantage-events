// Copyright © 2019 Emando B.V.

package events

import "github.com/emando/vantage-events/pkg/entities"

// Competition is a Vantage competition event.
type Competition struct {
	Base
	CompetitionID string `json:"competitionId"`
}

const (
	// CompetitionActivatedType is the event name of a Vantage competition activation.
	CompetitionActivatedType = "CompetitionActivatedEvent"
)

// CompetitionActivated is the event data of a Vantage competition activation.
type CompetitionActivated struct {
	Competition
	Value entities.Competition `json:"competition"`
}
