package datastore

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"cloud.google.com/go/datastore"
	"cloud.google.com/go/datastore/apiv1/datastorepb"
)

type Value struct {
	Type  Type `json:"type"`
	Value any  `json:"value"`
}

type EmbeddedEntity struct {
	Key        *Key       `json:"key,omitempty"`
	Properties []Property `json:"properties"`
}

func (v *Value) fromDatastoreProtoValue(src *datastorepb.Value) {
	// NOTE: not fully types supported because actual types are not used (this is just used by the aggregation query)
	switch value := src.ValueType.(type) {
	case *datastorepb.Value_IntegerValue:
		v.Type = IntType
		v.Value = value.IntegerValue
	case *datastorepb.Value_DoubleValue:
		v.Type = FloatType
		v.Value = value.DoubleValue
	case *datastorepb.Value_NullValue:
		v.Type = NullType
		v.Value = nil
	default:
		panic(fmt.Sprintf("unexpected value type: %T", src.ValueType))
	}
}

func (v *Value) fromDatastoreValue(src any) {
	v.Type = getType(src)
	switch v.Type {
	case ArrayType:
		src := src.([]any)
		dest := make([]Value, len(src))
		for i, v := range src {
			dest[i].fromDatastoreValue(v)
		}
		v.Value = dest

	case EntityType:
		src := src.(*datastore.Entity)
		dest := make([]Property, len(src.Properties))
		for i, v := range src.Properties {
			dest[i].fromDatastoreProperty(v)
		}
		if src.Key == nil {
			v.Value = dest
		} else {
			v.Value = EmbeddedEntity{
				Key:        FromDatastoreKey(src.Key),
				Properties: dest,
			}
		}

	case GeoPointType:
		src := src.(datastore.GeoPoint)
		v.Value = GeoPoint(src)

	case KeyType:
		src := src.(*datastore.Key)
		v.Value = FromDatastoreKey(src)

	default:
		v.Value = src
	}
}

func (v *Value) toDatastoreValue() any {
	switch v.Type {
	case ArrayType:
		src := v.Value.([]Value)
		dest := make([]any, len(src))
		for i, v := range src {
			dest[i] = v.toDatastoreValue()
		}
		return dest

	case EntityType:
		var key *datastore.Key
		var properties []Property
		switch src := v.Value.(type) {
		case []Property:
			properties = src
		case EmbeddedEntity:
			properties = src.Properties
			if src.Key != nil {
				key = src.Key.ToDatastore()
			}
		default:
			panic(fmt.Sprintf("unexpected entity value type: %T", v.Value))
		}
		dest := &datastore.Entity{Key: key, Properties: make([]datastore.Property, len(properties))}
		for i, v := range properties {
			dest.Properties[i] = v.toDatastoreProperty()
		}
		return dest

	case GeoPointType:
		src := v.Value.(GeoPoint)
		return datastore.GeoPoint(src)

	case KeyType:
		src := v.Value.(*Key)
		return src.ToDatastore()

	default:
		return v.Value
	}
}

func (v *Value) UnmarshalJSON(b []byte) error {
	var box struct {
		Type  Type            `json:"type"`
		Value json.RawMessage `json:"value"`
	}
	if err := json.Unmarshal(b, &box); err != nil {
		return err
	}
	v.Type = box.Type

	switch box.Type {
	case ArrayType:
		value, err := unmarshalJSON[[]Value](box.Value)
		if err != nil {
			return err
		}
		v.Value = value
	case BlobType:
		value, err := unmarshalJSON[[]byte](box.Value)
		if err != nil {
			return err
		}
		v.Value = value
	case BoolType:
		value, err := unmarshalJSON[bool](box.Value)
		if err != nil {
			return err
		}
		v.Value = value
	case TimestampType:
		value, err := unmarshalJSON[time.Time](box.Value)
		if err != nil {
			return err
		}
		v.Value = value
	case EntityType:
		valueJSON := bytes.TrimSpace(box.Value)
		switch {
		case bytes.HasPrefix(valueJSON, []byte("[")):
			value, err := unmarshalJSON[[]Property](box.Value)
			if err != nil {
				return err
			}
			v.Value = value
		case bytes.HasPrefix(valueJSON, []byte("{")):
			value, err := unmarshalJSON[EmbeddedEntity](box.Value)
			if err != nil {
				return err
			}
			v.Value = value
		default:
			return fmt.Errorf("entity value must be an array or object")
		}
	case FloatType:
		value, err := unmarshalJSON[float64](box.Value)
		if err != nil {
			return err
		}
		v.Value = value
	case GeoPointType:
		value, err := unmarshalJSON[GeoPoint](box.Value)
		if err != nil {
			return err
		}
		v.Value = value
	case IntType:
		value, err := unmarshalJSON[int64](box.Value)
		if err != nil {
			return err
		}
		v.Value = value
	case KeyType:
		value, err := unmarshalJSON[*Key](box.Value)
		if err != nil {
			return err
		}
		v.Value = value
	case NullType:
		v.Value = nil
	case StringType:
		value, err := unmarshalJSON[string](box.Value)
		if err != nil {
			return err
		}
		v.Value = value
	default:
		return fmt.Errorf("unknown type: %s", box.Type)
	}

	return nil
}
