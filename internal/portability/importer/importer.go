package importer

import (
	"encoding/json"
	"io"
)

// FromJSON reads JSON data from a reader and returns a slice of the given type.
func FromJSON[T any](r io.Reader) (T, error) {
	var data T

	if err := json.NewDecoder(r).Decode(&data); err != nil {
		return data, err
	}

	return data, nil
}
