package command

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/karupanerura/datastore-cli/internal/datastore"
	"github.com/karupanerura/datastore-cli/internal/parser"
	"google.golang.org/api/iterator"
)

type QueryCommand struct {
	Kind        string   `arg:"" name:"kind" help:"Entity kind"`
	KeysOnly    bool     `name:"keys-only" optional:"" help:"Return only keys of entities"`
	AncestorKey string   `name:"ancestor" optional:"" help:"Ancestor key to query (format: https://support.google.com/cloud/answer/6361641)"`
	Distinct    bool     `name:"distinct" optional:""`
	DistinctOn  []string `name:"distinctOn" optional:""`
	Project     []string `name:"project" optional:""`
	Filter      string   `name:"filter" optional:"" help:"Entity filter query (format: GQL compound-condition https://cloud.google.com/datastore/docs/reference/gql_reference)"`
	Order       []string `name:"order" optional:"" help:"Comma separated property names with optional '-' prefix for descending order"`
	Limit       int      `name:"limit" optional:"" help:"Limit number of entities to query"`
	Offset      int      `name:"offset" optional:"" help:"Offset number of entities to query"`
}

func (r *QueryCommand) Run(ctx context.Context, opts Options) error {
	client, err := datastore.NewClient(ctx, opts.Datastore())
	if err != nil {
		return err
	}

	query := datastore.NewQuery(r.Kind)
	if r.KeysOnly {
		query = query.KeysOnly()
	}
	if r.AncestorKey != "" {
		keyParser := &parser.KeyParser{Namespace: opts.Namespace}
		key, err := keyParser.ParseKey(r.AncestorKey)
		if err != nil {
			return fmt.Errorf("keyParser.ParseKey: %w", err)
		}
		query = query.Ancestor(key.ToDatastore())
	}
	if r.Distinct {
		query = query.Distinct()
	}
	if len(r.DistinctOn) != 0 {
		query = query.DistinctOn(r.DistinctOn...)
	}
	if len(r.Project) != 0 {
		query = query.Project(r.Project...)
	}
	if r.Filter != "" {
		filterParser := &parser.FilterParser{Namespace: opts.Namespace}
		filter, err := filterParser.ParseFilter(r.Filter)
		if err != nil {
			return fmt.Errorf("filterParser.ParseFilter: %w", err)
		}
		query = query.FilterEntity(filter)
	}
	for _, order := range r.Order {
		query = query.Order(order)
	}
	if r.Limit != 0 {
		query = query.Limit(r.Limit)
	}
	if r.Offset != 0 {
		query = query.Offset(r.Offset)
	}

	iter := client.Run(ctx, query)
	encoder := json.NewEncoder(os.Stdout)
	if r.KeysOnly {
		for {
			key, err := iter.Next(nil)
			if err == iterator.Done {
				return nil
			} else if err != nil {
				return err
			}
			if err := encoder.Encode(datastore.FromDatastoreKey(key)); err != nil {
				return err
			}
		}
	} else {
		for {
			var entity datastore.Entity
			_, err := iter.Next(&entity)
			if err == iterator.Done {
				break
			} else if err != nil {
				return err
			}

			if err := encoder.Encode(entity); err != nil {
				return err
			}
		}
		return nil
	}
}
