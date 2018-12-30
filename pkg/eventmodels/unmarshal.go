// Copyright © 2018 Emando B.V.

package eventmodels

import (
	"encoding/json"
	"fmt"
)

type typeNamer interface {
	TypeName() string
}

// Unmarshal parses the JSON encoded data.
// This function returns an error if the type name does not equal the given type name.
func Unmarshal(data []byte, typeName string, v interface{}) error {
	if err := json.Unmarshal(data, v); err != nil {
		return err
	}
	if e, ok := v.(typeNamer); ok && e.TypeName() != typeName {
		return fmt.Errorf("events: invalid type name (%v)", e.TypeName())
	}
	return nil
}
