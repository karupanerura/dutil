package datastore

import (
	"encoding/json"
	"fmt"
	"time"

	"cloud.google.com/go/datastore"
)

type Value struct {
	Type  Type `json:"type"`
	Value any  `json:"value"`
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
		v.Value = dest

	case GeoPointType:
		src := src.(datastore.GeoPoint)
		v.Value = GeoPoint(src)

	case KeyType:
		src := src.(*datastore.Key)
		v.Value = fromDatastoreKey(src)

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
		src := v.Value.([]Property)
		dest := &datastore.Entity{Properties: make([]datastore.Property, len(src))}
		for i, v := range src {
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
		value, err := unmarshalJSON[[]Property](box.Value)
		if err != nil {
			return err
		}
		v.Value = value
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
