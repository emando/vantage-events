// Copyright © 2019 Emando B.V.

package events

// Base contains fields of all events.
type Base struct {
	Type string `json:"typeName"`
}

// TypeName returns the event type.
func (b Base) TypeName() string {
	return b.Type
}

// Raw is a raw event.
type Raw struct {
	Base
	Bytes []byte `json:"-"`
}
