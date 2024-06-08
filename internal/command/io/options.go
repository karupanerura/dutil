package io

import (
	"context"

	"github.com/karupanerura/dutil/internal/datastore"
)

type DatastoreOptions struct {
	// ProjectID is Google Cloud project ID
	ProjectID string `short:"p" name:"projectId" env:"DATASTORE_PROJECT_ID" help:"Google Cloud Project ID" required:""`

	// DatabaseID is Cloud Datastore database ID
	DatabaseID string `short:"d" name:"databaseId" help:"Cloud Datastore database ID" optional:""`

	// Namespace is Cloud Datastore namespace
	Namespace string `short:"n" name:"namespace" help:"Cloud Datastore namespace" optional:""`

	// EmulatorHost is Cloud Datastore emulator host
	EmulatorHost string `env:"DATASTORE_EMULATOR_HOST" help:"Cloud Datastore emulator host" optional:""`
}

func (c *DatastoreOptions) CreateClient(ctx context.Context) (*datastore.Client, error) {
	return datastore.NewClient(ctx, datastore.Options{
		ProjectID:  c.ProjectID,
		DatabaseID: c.DatabaseID,
		Emulator:   c.EmulatorHost,
	})
}
