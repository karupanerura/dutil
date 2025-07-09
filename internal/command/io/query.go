package io

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"google.golang.org/api/iterator"

	"github.com/karupanerura/dutil/internal/command"
	"github.com/karupanerura/dutil/internal/datastore"
	"github.com/karupanerura/dutil/internal/parser"
)

type FieldAndAlias struct {
	Field string
	Alias string
}

func (a *FieldAndAlias) UnmarshalText(text []byte) error {
	if n := bytes.IndexByte(text, '='); n != -1 {
		a.Field = string(text[:n])
		a.Alias = string(text[n+1:])
	} else {
		a.Field = string(text)
	}
	return nil
}

type QueryCommand struct {
	DatastoreOptions
	Kind        string        `arg:"" name:"kind" help:"Entity kind"`
	KeyFormat   string        `name:"key-format" enum:"json,gql,encoded,proto" default:"json" help:"Key format to output for keys only query"`
	KeysOnly    bool          `name:"keys-only" optional:"" group:"Query" help:"Return only keys of entities"`
	AncestorKey string        `name:"ancestor" optional:"" group:"Query" help:"Ancestor key to query (format: https://support.google.com/cloud/answer/6361641)"`
	Distinct    bool          `name:"distinct" optional:"" group:"Query"`
	DistinctOn  []string      `name:"distinctOn" optional:"" group:"Query"`
	Project     []string      `name:"project" optional:"" group:"Query"`
	Filter      string        `name:"filter" optional:"" group:"Query" help:"Entity filter query (format: GQL compound-condition https://cloud.google.com/datastore/docs/reference/gql_reference)"`
	Order       []string      `name:"order" optional:"" group:"Query" help:"Comma separated property names with optional '-' prefix for descending order"`
	Limit       int           `name:"limit" optional:""  group:"Query" help:"Limit number of entities to query"`
	Offset      int           `name:"offset" optional:"" group:"Query" help:"Offset number of entities to query"`
	Explain     bool          `name:"explain" optional:"" group:"Query" help:"Explain query execution plan"`
	Count       *string       `name:"count" optional:"" group:"Aggregation" help:"Count entities using aggregation query, the value is alias name of the count result. (e.g. --count= or --count=myAlias)"`
	Sum         FieldAndAlias `name:"sum" optional:"" group:"Aggregation" help:"Sum entities field using aggregation query, the value is a target field name and optional alias name. (e.g. --sum=myField or --sum=myField=myAlias)"`
	Average     FieldAndAlias `name:"avg" optional:"" group:"Aggregation" help:"Average entities field using aggregation query, the value is a target field name and optional alias name. (e.g. --sum=myField or --sum=myField=myAlias)"`
}

func (r *QueryCommand) Run(ctx context.Context, opts command.GlobalOptions) error {
	client, err := r.DatastoreOptions.CreateClient(ctx)
	if err != nil {
		return err
	}
	defer client.Close()

	query := datastore.NewQuery(r.Kind)
	if r.KeysOnly {
		query = query.KeysOnly()
	}
	if r.Namespace != "" {
		query = query.Namespace(r.Namespace)
	}
	if r.AncestorKey != "" {
		keyParser := &parser.KeyParser{Namespace: r.Namespace}
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
		filterParser := &parser.FilterParser{Namespace: r.Namespace}
		ancestor, filter, err := filterParser.ParseFilter(r.Filter)
		if err != nil {
			return fmt.Errorf("filterParser.ParseFilter: %w", err)
		}
		if ancestor != nil {
			return fmt.Errorf("ancestor condition is not supported, use --ancestor option instead")
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
	if r.Count != nil || r.Sum.Field != "" || r.Average.Field != "" {
		aq := query.NewAggregationQuery()
		if r.Count != nil {
			aq = aq.WithCount(string(*r.Count))
		}
		if r.Sum.Field != "" {
			aq = aq.WithSum(r.Sum.Field, r.Sum.Alias)
		}
		if r.Average.Field != "" {
			aq = aq.WithAvg(r.Average.Field, r.Average.Alias)
		}
		ar, err := client.RunAggregationQuery(ctx, aq)
		if err != nil {
			return err
		}

		props := datastore.NewPropertiesByProtoValueMap(ar)
		b, err := json.Marshal(props)
		if err != nil {
			return err
		}

		_, err = io.Copy(os.Stdout, io.MultiReader(bytes.NewReader(b), strings.NewReader("\n")))
		if err != nil {
			return err
		}
		return nil
	}

	options := []datastore.RunOption{}
	if r.Explain {
		options = append(options, datastore.ExplainOptions{Analyze: true})
	}

	keyFormatter := datastore.KeyFormatter{Format: r.KeyFormat}

	iter := client.RunWithOptions(ctx, query, options...)
	if r.Explain {
		// read all
		for {
			if _, err := iter.Next(nil); err == iterator.Done {
				return json.NewEncoder(os.Stdout).Encode(iter.ExplainMetrics)
			} else if err != nil {
				return err
			}
		}
	}

	encoder := json.NewEncoder(os.Stdout)
	for {
		var entity datastore.Entity
		key, err := iter.Next(&entity)
		if err == iterator.Done {
			break
		} else if err != nil {
			return err
		}

		if len(entity.Properties) == 0 {
			key := keyFormatter.FormatKey(datastore.FromDatastoreKey(key))
			if s, ok := key.(string); ok {
				os.Stdout.WriteString(s)
				os.Stdout.WriteString("\n")
			} else {
				if err := encoder.Encode(key); err != nil {
					return err
				}
			}
		} else {
			if err := encoder.Encode(entity); err != nil {
				return err
			}
		}
	}
	return nil
}
