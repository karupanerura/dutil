# datastore-cli

Google Cloud Datastore unofficial CLI client.

## Usage

```prompt
$ datastore-cli -p my-project1 lookup 'KEY(MyKind, "foo")'
$ datastore-cli -p my-project1 gql 'SELECT * FROM MyKind WHERE prop > 2'
$ datastore-cli -p my-project1 query MyKind --ancestor 'KEY(MyParentKind, "foo")' > dump.jsonl
$ datastore-cli -p my-project2 upsert < dump.jsonl
$ datastore-cli -p my-project1 query MyKind --where 'prop > 2' --keys-only --format=encoded | xargs datastore-cli -p my-project1 delete
```

## Install

Pre-built binaries are available on: https://github.com/karupanerura/datastore-cli/releases/tag/v0.0.4

```prompt
$ VERSION=0.0.4
$ curl -sfLO https://github.com/karupanerura/datastore-cli/releases/download/v${VERSION}/datastore-cli_${VERSION}_$(go env GOOS)_$(go env GOARCH).tar.gz
$ tar zxf datastore-cli_${VERSION}_$(go env GOOS)_$(go env GOARCH).tar.gz
$ install -m 0755 datastore-cli $PREFIX
$ rm datastore-cli datastore-cli_${VERSION}_$(go env GOOS)_$(go env GOARCH).tar.gz
```

## Format

This command dumps and upsert (insert or update) with [JSON Lines](https://jsonlines.org/) format.
The JSONs format written in TypeScript is following:

### Key

Key is a datastore key for entity.

```typescript
type Key = {
    namespace?: string
    kind: string
    name?: string
    id?: number // int64
    parent?: Key
}
```

### Entity

Entity is a datastore entity.

```typescript
type Entity = {
    key: Key
    properties: Property[]
}
```

### Property

Property is a property of datastore entity.

```typescript
type Property = {
    name: string
    noIndex?: true
} & Value
```

### Value

Value is a value of datastore property.
This is object has the properties `type` and `value`.
The data type of `value` is determined by the `type` value.

```typescript
type Value = {
    type: "array"
    value: Value[]
} | {
    type: "blob"
    value: string // standard base64-encoded string
} | {
    type: "bool"
    value: boolean
} | {
    type: "timestamp"
    value: string // RFC3339 format
} | {
    type: "entity"
    value: []Property
} | {
    type: "float"
    value: number // float64
} | {
    type: "array"
    value: []Value
} | {
    type: "geo"
    value: GeoPoint // Explained in subsequent sections
} | {
    type: "int"
    value: number // int64
} | {
    type: "key"
    value: Key
} | {
    type: "null"
    value: null
} | {
    type: "string"
    value: string
}
```

### GeoPoint

GeoPoint represents a geographical point with latitude and longitude coordinates.

```typescript
type GeoPoint = {
    lat: number // float64
    lng: number // float64
}
```
