package datastore

import (
	"fmt"
	"time"

	"cloud.google.com/go/datastore"
)

type Client = datastore.Client

type Transaction = datastore.Transaction

type Mutation = datastore.Mutation

var (
	NewInsert = datastore.NewInsert
	NewUpdate = datastore.NewUpdate
)

type MultiError = datastore.MultiError

var ErrNoSuchEntity = datastore.ErrNoSuchEntity

type Query = datastore.Query

type AggregationQuery = datastore.AggregationQuery

var NewQuery = datastore.NewQuery

type EntityFilter = datastore.EntityFilter

type AndFilter = datastore.AndFilter

type OrFilter = datastore.OrFilter

type PropertyFilter = datastore.PropertyFilter

type RunOption = datastore.RunOption

type (
	ExplainOptions = datastore.ExplainOptions
	ExplainMetrics = datastore.ExplainMetrics
)

type GeoPoint struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

type Type string

const (
	ArrayType     Type = "array"
	BlobType      Type = "blob"
	BoolType      Type = "bool"
	TimestampType Type = "timestamp"
	EntityType    Type = "entity"
	FloatType     Type = "float"
	GeoPointType  Type = "geo"
	IntType       Type = "int"
	KeyType       Type = "key"
	NullType      Type = "null"
	StringType    Type = "string"
)

func getType(v any) Type {
	if v == nil {
		return NullType
	}

	switch v.(type) {
	case []any:
		return ArrayType
	case []byte:
		return BlobType
	case bool:
		return BoolType
	case time.Time:
		return TimestampType
	case *datastore.Entity:
		return EntityType
	case float64:
		return FloatType
	case datastore.GeoPoint:
		return GeoPointType
	case int64:
		return IntType
	case *datastore.Key:
		return KeyType
	case string:
		return StringType
	default:
		panic(fmt.Sprintf("unknown type: %+v", v))
	}
}
