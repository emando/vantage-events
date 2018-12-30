// Copyright © 2018 Emando B.V.

package eventmodels

import events "github.com/johanstokking/vantage-events"

type Competition struct {
	Base
	CompetitionID string `json:"competitionId"`
}

const (
	CompetitionActivatedType = "CompetitionActivatedEvent"
)

type CompetitionActivated struct {
	Competition
	Value events.Competition `json:"competition"`
}
