package command

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/karupanerura/datastore-cli/internal/datastore"
)

type UpsertCommand struct {
	Force  bool `name:"force" short:"f" optional:"" env:"DATASTORE_CLI_FORCE_UPSERT" help:"Force upsert without confirmation"`
	Commit bool `name:"commit" short:"c" optional:"" help:"Commit transaction without confirmation"`
	Silent bool `name:"silent" short:"s" optional:"" help:"Silent mode"`
}

func (r *UpsertCommand) Run(ctx context.Context, opts Options) error {
	client, err := datastore.NewClient(ctx, opts.Datastore())
	if err != nil {
		return err
	}

	var entities []*datastore.Entity
	var keys datastore.Keys
	decoder := json.NewDecoder(os.Stdin)
	for {
		var entity *datastore.Entity
		if err := decoder.Decode(&entity); err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		keys = append(keys, entity.Key)
		entities = append(entities, entity)
	}

	// pre confirmation
	if !r.Silent {
		log.Printf("%d keys to upsert:", len(keys))
		for _, key := range keys {
			log.Println(key.String())
		}
	}
	if !r.Force && !confirm("Update or insert these entities?") {
		return fmt.Errorf("aborted")
	}

	if _, err = client.RunInTransaction(ctx, func(tx *datastore.Transaction) error {
		if _, err := tx.PutMulti(keys.ToDatastore(), entities); err != nil {
			return fmt.Errorf("client.PutMulti: %w", err)
		}

		// post confirmation
		if !r.Force && !r.Commit && !confirm("Commit?") {
			return fmt.Errorf("aborted")
		}

		return nil
	}); err != nil {
		return fmt.Errorf("client.RunInTransaction: %w", err)
	}

	return nil
}
