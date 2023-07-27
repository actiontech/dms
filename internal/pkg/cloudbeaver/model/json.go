package model

import (
	"encoding/json"
	"fmt"
	"io"
)

type JSON map[string]interface{}

// UnmarshalGQL implements the graphql.Unmarshaler interface
func (b *JSON) UnmarshalGQL(v interface{}) error {
	*b = make(map[string]interface{})
	byteData, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("marshal scheme err: %v", err)
	}
	tmp := make(map[string]interface{})
	err = json.Unmarshal(byteData, &tmp)
	if err != nil {
		return fmt.Errorf("unmarshal scheme err: %v", err)
	}
	*b = tmp
	return nil
}

// MarshalGQL implements the graphql.Marshaler interface
func (b JSON) MarshalGQL(w io.Writer) {
	// :todo handle err
	byteData, err := json.Marshal(b)
	if err != nil {
		panic("FAIL WHILE MARSHAL SCHEME")
	}
	_, _ = w.Write(byteData)
}
