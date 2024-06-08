package io

import (
	"context"
	"fmt"
	"log"

	"github.com/karupanerura/dutil/internal/command"
	"github.com/karupanerura/dutil/internal/datastore"
	"github.com/karupanerura/dutil/internal/parser"
)

type DeleteCommand struct {
	DatastoreOptions
	Keys   []string `arg:"" name:"keys" help:"Keys to delete (format: https://support.google.com/cloud/answer/6361641)"`
	Force  bool     `name:"force" short:"f" optional:"" env:"DATASTORE_CLI_FORCE_DELETE" help:"Force delete without confirmation"`
	Commit bool     `name:"commit" short:"c" optional:"" help:"Commit transaction without confirmation"`
	Silent bool     `name:"silent" short:"s" optional:"" help:"Silent mode"`
}

func (r *DeleteCommand) Run(ctx context.Context, opts command.GlobalOptions) error {
	client, err := r.DatastoreOptions.CreateClient(ctx)
	if err != nil {
		return err
	}
	defer client.Close()

	keyParser := &parser.KeyParser{Namespace: r.Namespace}
	keys, err := keyParser.ParseKeys(r.Keys)
	if err != nil {
		return fmt.Errorf("keyParser.ParseKeys: %w", err)
	}

	// pre confirmation
	if !r.Silent {
		log.Printf("%d keys to delete:", len(keys))
		for _, key := range keys {
			log.Println(key.String())
		}
	}
	if !r.Force && !confirm("Delete or insert these entities?") {
		return fmt.Errorf("aborted")
	}

	if _, err = client.RunInTransaction(ctx, func(tx *datastore.Transaction) error {
		if err := tx.DeleteMulti(keys.ToDatastore()); err != nil {
			return fmt.Errorf("client.DeleteMulti: %w", err)
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
