package io

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/karupanerura/dutil/internal/command"
	"github.com/karupanerura/dutil/internal/datastore"
)

type InsertCommand struct {
	DatastoreOptions
	Force  bool `name:"force" short:"f" optional:"" env:"DATASTORE_CLI_FORCE_INSERT" help:"Force insert without confirmation"`
	Commit bool `name:"commit" short:"c" optional:"" help:"Commit transaction without confirmation"`
	Silent bool `name:"silent" short:"s" optional:"" help:"Silent mode"`
}

func (r *InsertCommand) Run(ctx context.Context, opts command.GlobalOptions) error {
	client, err := r.DatastoreOptions.CreateClient(ctx)
	if err != nil {
		return err
	}
	defer client.Close()

	var keys datastore.Keys
	var mutations []*datastore.Mutation
	decoder := json.NewDecoder(os.Stdin)
	for {
		var entity *datastore.Entity
		if err := decoder.Decode(&entity); err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		keys = append(keys, entity.Key)
		mutations = append(mutations, datastore.NewInsert(entity.Key.ToDatastore(), entity))
	}

	// pre confirmation
	if !r.Silent {
		log.Printf("%d keys to insert:", len(keys))
		for _, key := range keys {
			log.Println(key.String())
		}
	}
	if !r.Force && !confirm("Insert these entities?") {
		return fmt.Errorf("aborted")
	}

	if _, err = client.RunInTransaction(ctx, func(tx *datastore.Transaction) error {
		if _, err := tx.Mutate(mutations...); err != nil {
			return fmt.Errorf("client.Mutate: %w", err)
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
