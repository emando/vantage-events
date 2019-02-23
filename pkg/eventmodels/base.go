// Copyright © 2019 Emando B.V.

package eventmodels

type Base struct {
	Type string `json:"typeName"`
}

func (b Base) TypeName() string {
	return b.Type
}
