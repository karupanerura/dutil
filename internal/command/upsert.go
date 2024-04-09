package command

import (
	"context"
	"encoding/json"
	"io"
	"os"

	"github.com/karupanerura/datastore-cli/internal/datastore"
)

type UpsertCommand struct {
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

	if _, err := client.PutMulti(ctx, keys.ToDatastore(), entities); err != nil {
		return err
	}
	return nil
}
