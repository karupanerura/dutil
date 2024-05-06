package command

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/karupanerura/datastore-cli/internal/datastore"
	"github.com/karupanerura/datastore-cli/internal/parser"
)

type LookupCommand struct {
	Keys         []string `arg:"" name:"keys" help:"Keys to lookup (format: https://support.google.com/cloud/answer/6361641)"`
	WithMetadata bool     `name:"with-metadata" help:"Lookup with internal metadata in datastore (EXPERIMENTAL)"`
}

func (r *LookupCommand) Run(ctx context.Context, opts Options) error {
	client, err := datastore.NewClient(ctx, opts.Datastore())
	if err != nil {
		return err
	}
	defer client.Close()

	keyParser := &parser.KeyParser{Namespace: opts.Namespace}
	keys, err := keyParser.ParseKeys(r.Keys)
	if err != nil {
		return fmt.Errorf("keyParser.ParseKeys: %w", err)
	}

	entities := make([]*datastore.Entity, len(keys))
	if err := client.GetMulti(ctx, keys.ToDatastore(), entities); err != nil {
		var mErr datastore.MultiError
		if errors.As(err, &mErr) {
			for _, err := range mErr {
				if err != nil && !errors.Is(err, datastore.ErrNoSuchEntity) {
					return mErr
				}
			}
		} else {
			return err
		}
	}

	if r.WithMetadata {
		llc := datastore.NewLowLevelClient(client)
		for i, key := range keys {
			meta, err := llc.GetMetadata(ctx, key.ToDatastore())
			if err != nil {
				return err
			}
			entities[i].Metadata = meta
		}
	}

	encoder := json.NewEncoder(os.Stdout)
	for _, entity := range entities {
		if err := encoder.Encode(entity); err != nil {
			return err
		}
	}
	return nil
}
