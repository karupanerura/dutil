package io

type Commands struct {
	Lookup LookupCommand `cmd:""`
	Query  QueryCommand  `cmd:""`
	Insert InsertCommand `cmd:""`
	Update UpdateCommand `cmd:""`
	Upsert UpsertCommand `cmd:""`
	Delete DeleteCommand `cmd:""`
	GQL    GQLCommand    `cmd:""`
}
