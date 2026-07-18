package datastore

import (
	"testing"

	clouddatastore "cloud.google.com/go/datastore"
	"google.golang.org/api/option"
)

func TestNewLowLevelClientSDKCompatibility(t *testing.T) {
	const (
		projectID  = "reflection-smoke-project"
		databaseID = "reflection-smoke-database"
	)

	client, err := clouddatastore.NewClientWithDatabase(
		t.Context(),
		projectID,
		databaseID,
		option.WithoutAuthentication(),
	)
	if err != nil {
		t.Fatalf("create Datastore client: %v", err)
	}
	t.Cleanup(func() {
		if err := client.Close(); err != nil {
			t.Errorf("close Datastore client: %v", err)
		}
	})

	lowLevelClient := NewLowLevelClient(client)
	if lowLevelClient == nil {
		t.Fatal("NewLowLevelClient returned nil")
	}
	if lowLevelClient.lc == nil {
		t.Fatal("NewLowLevelClient did not extract the low-level Datastore client")
	}
	if got := lowLevelClient.dataset; got != projectID {
		t.Errorf("dataset = %q, want %q", got, projectID)
	}
	if got := lowLevelClient.databaseID; got != databaseID {
		t.Errorf("databaseID = %q, want %q", got, databaseID)
	}
}
