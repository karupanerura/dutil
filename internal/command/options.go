package command

import "github.com/karupanerura/datastore-cli/internal/datastore"

type Options struct {
	// ProjectID is Google Cloud project ID
	ProjectID string `short:"p" name:"projectId" env:"DATASTORE_PROJECT_ID" help:"Google Cloud Project ID" required:""`

	// DatabaseID is Cloud Datastore database ID
	DatabaseID string `short:"d" name:"databaseId" help:"Cloud Datastore database ID" optional:""`

	// Namespace is Cloud Datastore namespace
	Namespace string `short:"n" name:"namespace" help:"Cloud Datastore namespace" optional:""`

	// EmulatorHost is Cloud Datastore emulator host
	EmulatorHost string `env:"DATASTORE_EMULATOR_HOST" help:"Cloud Datastore emulator host" optional:""`
}

func (c *Options) Datastore() datastore.Options {
	return datastore.Options{
		ProjectID:  c.ProjectID,
		DatabaseID: c.DatabaseID,
		Emulator:   c.EmulatorHost,
	}
}
