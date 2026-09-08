package service

import (
	"bytes"
	"encoding/json"
)

// Field distinguishes PATCH omission from explicit null, including zero values.
type Field[T any] struct {
	Set   bool
	Null  bool
	Value T
}

func (f *Field[T]) UnmarshalJSON(data []byte) error {
	f.Set = true
	f.Null = bytes.Equal(bytes.TrimSpace(data), []byte("null"))
	if f.Null {
		return nil
	}
	return json.Unmarshal(data, &f.Value)
}
