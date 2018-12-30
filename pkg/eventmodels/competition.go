// Copyright © 2018 Emando B.V.

package eventmodels

import events "github.com/johanstokking/vantage-events"

type Base struct {
	Type string `json:"typeName"`
}

func (b Base) TypeName() string {
	return b.Type
}

type Competition struct {
	Base
	CompetitionID string `json:"competitionId"`
}

const CompetitionActivatedType = "CompetitionActivatedEvent"

type CompetitionActivated struct {
	Competition
	Value events.Competition `json:"competition"`
}
