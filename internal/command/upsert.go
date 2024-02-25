package command

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
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
	rdr := bufio.NewReader(os.Stdin)
	for {
		line, _, err := rdr.ReadLine()
		if errors.Is(err, io.EOF) {
			break
		}

		var entity *datastore.Entity
		if err := json.Unmarshal(line, &entity); err != nil {
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
