# datastore-cli

Google Cloud Datastore unofficial CLI client.

## Usage

```
Usage: datastore-cli --projectId=STRING <command>

Flags:
  -h, --help                    Show context-sensitive help.
  -p, --projectId=STRING        Google Cloud Project ID ($DATASTORE_PROJECT_ID)
  -d, --databaseId=STRING       Cloud Datastore database ID
  -n, --namespace=STRING        Cloud Datastore namespace
      --emulator-host=STRING    Cloud Datastore emulator host ($DATASTORE_EMULATOR_HOST)

Commands:
  lookup --projectId=STRING <keys> ...

  query --projectId=STRING <kind>

  upsert --projectId=STRING

  delete --projectId=STRING <keys> ...

  gql --projectId=STRING <query>

Run "datastore-cli <command> --help" for more information on a command.
```

### lookup

```
Usage: datastore-cli lookup --projectId=STRING <keys> ...

Arguments:
  <keys> ...    Keys to lookup (format: https://support.google.com/cloud/answer/6361641)

Flags:
  -h, --help                    Show context-sensitive help.
  -p, --projectId=STRING        Google Cloud Project ID ($DATASTORE_PROJECT_ID)
  -d, --databaseId=STRING       Cloud Datastore database ID
  -n, --namespace=STRING        Cloud Datastore namespace
      --emulator-host=STRING    Cloud Datastore emulator host ($DATASTORE_EMULATOR_HOST)

      --with-metadata           Lookup with internal metadata in datastore (EXPERIMENTAL)
```

NOTE: `--with-metadata` is an experimental feature to lookup with datastore internal metadata.
To simplify implementation, it separates API calls for each key.

### query

```
Usage: datastore-cli query --projectId=STRING <kind>

Arguments:
  <kind>    Entity kind

Flags:
  -h, --help                    Show context-sensitive help.
  -p, --projectId=STRING        Google Cloud Project ID ($DATASTORE_PROJECT_ID)
  -d, --databaseId=STRING       Cloud Datastore database ID
  -n, --namespace=STRING        Cloud Datastore namespace
      --emulator-host=STRING    Cloud Datastore emulator host ($DATASTORE_EMULATOR_HOST)

      --key-format="json"       Key format to output for keys only query

Query
  --keys-only                    Return only keys of entities
  --ancestor=STRING              Ancestor key to query (format: https://support.google.com/cloud/answer/6361641)
  --distinct
  --distinctOn=DISTINCTON,...
  --project=PROJECT,...
  --filter=STRING                Entity filter query (format: GQL compound-condition https://cloud.google.com/datastore/docs/reference/gql_reference)
  --order=ORDER,...              Comma separated property names with optional '-' prefix for descending order
  --limit=INT                    Limit number of entities to query
  --offset=INT                   Offset number of entities to query

Aggregation
  --count=COUNT            Count entities using aggregation query, the value is alias name of the count result. (e.g. --count= or --count=myAlias)
  --sum=FIELD-AND-ALIAS    Sum entities field using aggregation query, the value is a target field name and optional alias name. (e.g. --sum=myField or --sum=myField=myAlias)
  --avg=FIELD-AND-ALIAS    Average entities field using aggregation query, the value is a target field name and optional alias name. (e.g. --sum=myField or --sum=myField=myAlias)
```

### gql

```
Usage: datastore-cli gql --projectId=STRING <query>

Arguments:
  <query>    GQL Query

Flags:
  -h, --help                    Show context-sensitive help.
  -p, --projectId=STRING        Google Cloud Project ID ($DATASTORE_PROJECT_ID)
  -d, --databaseId=STRING       Cloud Datastore database ID
  -n, --namespace=STRING        Cloud Datastore namespace
      --emulator-host=STRING    Cloud Datastore emulator host ($DATASTORE_EMULATOR_HOST)
```

### upsert

```
Usage: datastore-cli upsert --projectId=STRING

Flags:
  -h, --help                    Show context-sensitive help.
  -p, --projectId=STRING        Google Cloud Project ID ($DATASTORE_PROJECT_ID)
  -d, --databaseId=STRING       Cloud Datastore database ID
  -n, --namespace=STRING        Cloud Datastore namespace
      --emulator-host=STRING    Cloud Datastore emulator host ($DATASTORE_EMULATOR_HOST)

  -f, --force                   Force upsert without confirmation ($DATASTORE_CLI_FORCE_UPSERT)
  -c, --commit                  Commit transaction without confirmation
  -s, --silent                  Silent mode
```

### delete

```
Usage: datastore-cli delete --projectId=STRING <keys> ...

Arguments:
  <keys> ...    Keys to delete (format: https://support.google.com/cloud/answer/6361641)

Flags:
  -h, --help                    Show context-sensitive help.
  -p, --projectId=STRING        Google Cloud Project ID ($DATASTORE_PROJECT_ID)
  -d, --databaseId=STRING       Cloud Datastore database ID
  -n, --namespace=STRING        Cloud Datastore namespace
      --emulator-host=STRING    Cloud Datastore emulator host ($DATASTORE_EMULATOR_HOST)

  -f, --force                   Force delete without confirmation ($DATASTORE_CLI_FORCE_DELETE)
  -c, --commit                  Commit transaction without confirmation
  -s, --silent                  Silent mode
```

## Examples

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
