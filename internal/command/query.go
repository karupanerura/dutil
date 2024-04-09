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
	AncestorKey string   `name:"ancestor" optional:"" help:"Ancestor key to query (format: https://support.google.com/cloud/answer/6361641)"`
	Distinct    bool     `name:"distinct" optional:""`
	DistinctOn  []string `name:"distinctOn" optional:""`
	Project     []string `name:"project" optional:""`
	Filter      string   `name:"filter" optional:"" help:"(Not supported) Entity filter quert (format: GQL compound-condition https://cloud.google.com/datastore/docs/reference/gql_reference)"`
}

func (r *QueryCommand) Run(ctx context.Context, opts Options) error {
	client, err := datastore.NewClient(ctx, opts.Datastore())
	if err != nil {
		return err
	}

	query := datastore.NewQuery(r.Kind)
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
		if filter != datastore.NopFilter {
			query = query.FilterEntity(filter)
		}
	}

	iter := client.Run(ctx, query)
	encoder := json.NewEncoder(os.Stdout)
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
