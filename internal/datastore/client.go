package datastore

import (
	"context"
	"os"

	"cloud.google.com/go/datastore"
)

func NewClient(ctx context.Context, opts Options) (*datastore.Client, error) {
	if opts.Emulator != "" {
		os.Setenv("DATASTORE_EMULATOR_HOST", opts.Emulator)
	}
	return datastore.NewClientWithDatabase(ctx, opts.ProjectID, opts.DatabaseID)
}