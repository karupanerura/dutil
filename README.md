# datastore-cli

Google Cloud Datastore unofficial CLI client.

## Usage

```prompt
$ datastore-cli -p my-project1 lookup 'KEY(MyKind, "foo")'
$ datastore-cli -p my-project1 gql 'SELECT * FROM MyKind WHERE prop > 2'
$ datastore-cli -p my-project1 query --ancestor 'KEY(MyParentKind, "foo")' > dump.jsonl
$ datastore-cli -p my-project2 upsert < dump.jsonl
```

## Install

Pre-built binaryies are available on: https://github.com/karupanerura/datastore-cli/releases/tag/v0.0.3

```prompt
$ VERSION=0.0.1
$ curl -sfLO https://github.com/karupanerura/datastore-cli/releases/download/v${VERSION}/datastore-cli_${VERSION}_$(go env GOOS)_$(go env GOARCH).tar.gz
$ tar zxf datastore-cli_${VERSION}_$(go env GOOS)_$(go env GOARCH).tar.gz
$ install -m 0755 datastore-cli $PREFIX
$ rm datastore-cli datastore-cli_${VERSION}_$(go env GOOS)_$(go env GOARCH).tar.gz
```
