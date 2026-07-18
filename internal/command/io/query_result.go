package io

import (
	"encoding/json"
	"io"

	clouddatastore "cloud.google.com/go/datastore"

	"github.com/karupanerura/dutil/internal/datastore"
)

func writeQueryResult(stdout io.Writer, encoder *json.Encoder, keyFormatter datastore.KeyFormatter, key *clouddatastore.Key, entity datastore.Entity, keysOnly bool) error {
	if keysOnly {
		key := keyFormatter.FormatKey(datastore.FromDatastoreKey(key))
		if s, ok := key.(string); ok {
			_, _ = io.WriteString(stdout, s)
			_, _ = io.WriteString(stdout, "\n")
			return nil
		}
		return encoder.Encode(key)
	}
	return encoder.Encode(entity)
}
