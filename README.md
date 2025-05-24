# dutil

Google Cloud Firestore datastore mode unofficial CLI client and utilities. (dutil is named from `Datastore UTILity`.)

## Limitation

Support for datastore mode only. (Patches welcome)

## Examples

```prompt
$ dutil io lookup -p my-project1 'KEY(MyKind, "foo")' | dutil convert table
$ dutil io gql -p my-project1  'SELECT * FROM MyKind WHERE prop > 2'
$ dutil io query MyKind -p my-project1 --ancestor 'KEY(MyParentKind, "foo")' > dump.jsonl
$ dutil io upsert -p my-project2 < dump.jsonl
$ dutil io query MyKind -p my-project1 --where 'prop > 2' --keys-only --format=encoded | xargs dutil io delete -p my-project1
```

## Install

Pre-built binaries are available on: https://github.com/karupanerura/dutil/releases/tag/v0.2.0

```prompt
$ VERSION=0.2.0
$ curl -sfLO https://github.com/karupanerura/dutil/releases/download/v${VERSION}/dutil_${VERSION}_$(go env GOOS)_$(go env GOARCH).tar.gz
$ tar zxf dutil_${VERSION}_$(go env GOOS)_$(go env GOARCH).tar.gz
$ install -m 0755 dutil $PREFIX
$ rm dutil dutil_${VERSION}_$(go env GOOS)_$(go env GOARCH).tar.gz
```

## Usage

### dutil io

I/O utilities.

```
Usage: dutil io <command>

Flags:
  -h, --help       Show context-sensitive help.
      --version    Show version

Commands:
  io lookup --projectId=STRING <keys> ...

  io query --projectId=STRING <kind>

  io insert --projectId=STRING

  io update --projectId=STRING

  io upsert --projectId=STRING

  io delete --projectId=STRING <keys> ...

  io gql --projectId=STRING <query>
```

#### dutil io lookup

```
Usage: dutil io lookup --projectId=STRING <keys> ...

Arguments:
  <keys> ...    Keys to lookup (format:
                https://support.google.com/cloud/answer/6361641)

Flags:
  -h, --help                    Show context-sensitive help.
      --version                 Show version

  -p, --projectId=STRING        Google Cloud Project ID ($DATASTORE_PROJECT_ID)
  -d, --databaseId=STRING       Cloud Datastore database ID
  -n, --namespace=STRING        Cloud Datastore namespace
      --emulator-host=STRING    Cloud Datastore emulator host
                                ($DATASTORE_EMULATOR_HOST)
      --with-metadata           Lookup with internal metadata in datastore
                                (EXPERIMENTAL)
```

NOTE: `--with-metadata` is an experimental feature to lookup with datastore internal metadata.
To simplify implementation, it separates API calls for each key.

#### dutil io query

```
Usage: dutil io query --projectId=STRING <kind>

Arguments:
  <kind>    Entity kind

Flags:
  -h, --help                    Show context-sensitive help.
      --version                 Show version

  -p, --projectId=STRING        Google Cloud Project ID ($DATASTORE_PROJECT_ID)
  -d, --databaseId=STRING       Cloud Datastore database ID
  -n, --namespace=STRING        Cloud Datastore namespace
      --emulator-host=STRING    Cloud Datastore emulator host
                                ($DATASTORE_EMULATOR_HOST)
      --key-format="json"       Key format to output for keys only query

Query
  --keys-only                    Return only keys of entities
  --ancestor=STRING              Ancestor key to query (format:
                                 https://support.google.com/cloud/answer/6361641)
  --distinct
  --distinctOn=DISTINCTON,...
  --project=PROJECT,...
  --filter=STRING                Entity filter query (format:
                                 GQL compound-condition
                                 https://cloud.google.com/datastore/docs/reference/gql_reference)
  --order=ORDER,...              Comma separated property names with optional
                                 '-' prefix for descending order
  --limit=INT                    Limit number of entities to query
  --offset=INT                   Offset number of entities to query
  --explain                      Explain query execution plan

Aggregation
  --count=COUNT            Count entities using aggregation query, the value
                           is alias name of the count result. (e.g. --count= or
                           --count=myAlias)
  --sum=FIELD-AND-ALIAS    Sum entities field using aggregation query, the value
                           is a target field name and optional alias name. (e.g.
                           --sum=myField or --sum=myField=myAlias)
  --avg=FIELD-AND-ALIAS    Average entities field using aggregation query, the
                           value is a target field name and optional alias name.
                           (e.g. --sum=myField or --sum=myField=myAlias)
```

#### dutil io gql

```
Usage: dutil io gql --projectId=STRING <query>

Arguments:
  <query>    GQL Query

Flags:
  -h, --help                    Show context-sensitive help.
  -p, --projectId=STRING        Google Cloud Project ID ($DATASTORE_PROJECT_ID)
  -d, --databaseId=STRING       Cloud Datastore database ID
  -n, --namespace=STRING        Cloud Datastore namespace
      --emulator-host=STRING    Cloud Datastore emulator host ($DATASTORE_EMULATOR_HOST)
```

#### dutil io insert

```
Usage: dutil io insert --projectId=STRING

Flags:
  -h, --help                    Show context-sensitive help.
      --version                 Show version

  -p, --projectId=STRING        Google Cloud Project ID ($DATASTORE_PROJECT_ID)
  -d, --databaseId=STRING       Cloud Datastore database ID
  -n, --namespace=STRING        Cloud Datastore namespace
      --emulator-host=STRING    Cloud Datastore emulator host
                                ($DATASTORE_EMULATOR_HOST)
  -f, --force                   Force insert without confirmation
                                ($DATASTORE_CLI_FORCE_INSERT)
  -c, --commit                  Commit transaction without confirmation
  -s, --silent                  Silent mode
```

#### dutil io update

```
Usage: dutil io update --projectId=STRING

Flags:
  -h, --help                    Show context-sensitive help.
      --version                 Show version

  -p, --projectId=STRING        Google Cloud Project ID ($DATASTORE_PROJECT_ID)
  -d, --databaseId=STRING       Cloud Datastore database ID
  -n, --namespace=STRING        Cloud Datastore namespace
      --emulator-host=STRING    Cloud Datastore emulator host
                                ($DATASTORE_EMULATOR_HOST)
  -f, --force                   Force update without confirmation
                                ($DATASTORE_CLI_FORCE_UPDATE)
  -c, --commit                  Commit transaction without confirmation
  -s, --silent                  Silent mode
```

#### dutil io upsert

```
Usage: dutil io upsert --projectId=STRING

Flags:
  -h, --help                    Show context-sensitive help.
      --version                 Show version

  -p, --projectId=STRING        Google Cloud Project ID ($DATASTORE_PROJECT_ID)
  -d, --databaseId=STRING       Cloud Datastore database ID
  -n, --namespace=STRING        Cloud Datastore namespace
      --emulator-host=STRING    Cloud Datastore emulator host
                                ($DATASTORE_EMULATOR_HOST)
  -f, --force                   Force upsert without confirmation
                                ($DATASTORE_CLI_FORCE_UPSERT)
  -c, --commit                  Commit transaction without confirmation
  -s, --silent                  Silent mode
```

#### dutil io delete

```
Usage: dutil io delete --projectId=STRING <keys> ...

Arguments:
  <keys> ...    Keys to delete (format:
                https://support.google.com/cloud/answer/6361641)

Flags:
  -h, --help                    Show context-sensitive help.
      --version                 Show version

  -p, --projectId=STRING        Google Cloud Project ID ($DATASTORE_PROJECT_ID)
  -d, --databaseId=STRING       Cloud Datastore database ID
  -n, --namespace=STRING        Cloud Datastore namespace
      --emulator-host=STRING    Cloud Datastore emulator host
                                ($DATASTORE_EMULATOR_HOST)
  -f, --force                   Force delete without confirmation
                                ($DATASTORE_CLI_FORCE_DELETE)
  -c, --commit                  Commit transaction without confirmation
  -s, --silent                  Silent mode
```

### dutil convert

Data format converters.

```
Usage: dutil convert <command>

Flags:
  -h, --help       Show context-sensitive help.
      --version    Show version

Commands:
  convert table

  convert key
```

#### dutil convert table

```
Usage: dutil convert table

Flags:
  -h, --help             Show context-sensitive help.
      --version          Show version

  -f, --from="entity"    Type of JSON structure to convert to table
```

#### dutil convert key

```
Usage: dutil convert key

Flags:
  -h, --help            Show context-sensitive help.
      --version         Show version

      --from="auto"     Source key format
      --to="encoded"    Result key format
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
