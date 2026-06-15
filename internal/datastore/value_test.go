package datastore

import (
	"encoding/json"
	"reflect"
	"testing"

	clouddatastore "cloud.google.com/go/datastore"
)

func TestValueEntityRoundTripPreservesKey(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		key  *clouddatastore.Key
	}{
		{name: "without key"},
		{name: "with complete key", key: clouddatastore.NameKey("Embedded", "example", nil)},
		{name: "with incomplete key", key: clouddatastore.IncompleteKey("Embedded", nil)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			src := &clouddatastore.Entity{
				Key: tt.key,
				Properties: []clouddatastore.Property{
					{Name: "name", Value: "example"},
				},
			}

			var value Value
			value.fromDatastoreValue(src)

			got := value.toDatastoreValue().(*clouddatastore.Entity)
			if tt.key == nil {
				if got.Key != nil {
					t.Fatalf("unexpected key after round trip: %v", got.Key)
				}
				return
			}
			if got.Key == nil {
				t.Fatal("entity key was lost during round trip")
			}
			if !got.Key.Equal(tt.key) {
				t.Fatalf("entity key changed during round trip: got %v, want %v", got.Key, tt.key)
			}
		})
	}
}

func TestValueEntityJSONCompatibility(t *testing.T) {
	t.Parallel()

	legacyJSON := []byte(`{"type":"entity","value":[{"type":"string","value":"example","name":"name"}]}`)
	var legacy Value
	if err := json.Unmarshal(legacyJSON, &legacy); err != nil {
		t.Fatalf("failed to unmarshal legacy entity value: %v", err)
	}
	if _, ok := legacy.Value.([]Property); !ok {
		t.Fatalf("legacy entity value has unexpected type: %T", legacy.Value)
	}
	gotLegacy, err := json.Marshal(&legacy)
	if err != nil {
		t.Fatalf("failed to marshal legacy entity value: %v", err)
	}
	if !reflect.DeepEqual(gotLegacy, legacyJSON) {
		t.Fatalf("keyless entity JSON changed: got %s, want %s", gotLegacy, legacyJSON)
	}

	keyedJSON := []byte(`{"type":"entity","value":{"key":{"kind":"Embedded","name":"example"},"properties":[{"type":"string","value":"example","name":"name"}]}}`)
	var keyed Value
	if err := json.Unmarshal(keyedJSON, &keyed); err != nil {
		t.Fatalf("failed to unmarshal keyed entity value: %v", err)
	}
	entity := keyed.toDatastoreValue().(*clouddatastore.Entity)
	wantKey := clouddatastore.NameKey("Embedded", "example", nil)
	if entity.Key == nil || !entity.Key.Equal(wantKey) {
		t.Fatalf("unexpected embedded entity key: got %v, want %v", entity.Key, wantKey)
	}
	gotKeyed, err := json.Marshal(&keyed)
	if err != nil {
		t.Fatalf("failed to marshal keyed entity value: %v", err)
	}
	if !reflect.DeepEqual(gotKeyed, keyedJSON) {
		t.Fatalf("keyed entity JSON changed: got %s, want %s", gotKeyed, keyedJSON)
	}
}
