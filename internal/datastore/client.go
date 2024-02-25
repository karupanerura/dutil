package datastore

import (
	"context"

	"cloud.google.com/go/datastore"
)

func NewClient(ctx context.Context, opts Options) (*datastore.Client, error) {
	return datastore.NewClientWithDatabase(ctx, opts.ProjectID, opts.DatabaseID)
}
