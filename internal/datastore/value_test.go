package datastore

import (
	"encoding/json"
	"reflect"
	"testing"

	clouddatastore "cloud.google.com/go/datastore"
	"cloud.google.com/go/datastore/apiv1/datastorepb"
	structpb "google.golang.org/protobuf/types/known/structpb"
)

func TestValueFromDatastoreProtoValueNull(t *testing.T) {
	t.Parallel()

	src := &datastorepb.Value{
		ValueType: &datastorepb.Value_NullValue{NullValue: structpb.NullValue_NULL_VALUE},
	}
	var v Value
	v.fromDatastoreProtoValue(src)

	if v.Type != NullType {
		t.Errorf("Type = %q, want %q", v.Type, NullType)
	}
	if v.Value != nil {
		t.Errorf("Value = %v, want nil", v.Value)
	}

	got, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("failed to marshal value: %v", err)
	}
	if want := `{"type":"null","value":null}`; string(got) != want {
		t.Errorf("JSON = %s, want %s", got, want)
	}

	tests := []struct {
		name  string
		src   *datastorepb.Value
		type_ Type
		value any
	}{
		{
			name:  "integer",
			src:   &datastorepb.Value{ValueType: &datastorepb.Value_IntegerValue{IntegerValue: 42}},
			type_: IntType,
			value: int64(42),
		},
		{
			name:  "double",
			src:   &datastorepb.Value{ValueType: &datastorepb.Value_DoubleValue{DoubleValue: 1.5}},
			type_: FloatType,
			value: 1.5,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var v Value
			v.fromDatastoreProtoValue(tt.src)
			if v.Type != tt.type_ || v.Value != tt.value {
				t.Errorf("Value = {%q, %v}, want {%q, %v}", v.Type, v.Value, tt.type_, tt.value)
			}
		})
	}
}

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
