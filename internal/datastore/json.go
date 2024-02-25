package datastore

import "encoding/json"

func unmarshalJSON[T any](b []byte) (T, error) {
	var v T
	if err := json.Unmarshal(b, &v); err != nil {
		return v, err
	}
	return v, nil
}
