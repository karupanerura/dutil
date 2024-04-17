package command

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"os"
	"strings"

	"github.com/c-bata/go-prompt"
	"github.com/karupanerura/datastore-cli/internal/datastore"
	"github.com/karupanerura/datastore-cli/internal/parser"
	"google.golang.org/api/iterator"
)

type GQLCommand struct {
	Query string `arg:"" name:"query" optional:"" help:"GQL Query"`
}

func (r *GQLCommand) Run(ctx context.Context, opts Options) error {
	client, err := datastore.NewClient(ctx, opts.Datastore())
	if err != nil {
		return err
	}

	qp := &parser.QueryParser{Namespace: opts.Namespace}
	if r.Query != "" {
		return r.execute(ctx, client, qp, r.Query)
	}

	properties := map[string][]string{}
	{
		kinds, err := client.GetAll(ctx, datastore.NewQuery("__kind__").KeysOnly(), nil)
		if err != nil {
			return err
		}
		for _, kind := range kinds {
			props, err := client.GetAll(ctx, datastore.NewQuery("__property__").Ancestor(kind).KeysOnly(), nil)
			if err != nil {
				return err
			}
			properties[kind.Name] = make([]string, len(props))
			for _, prop := range props {
				properties[kind.Name] = append(properties[kind.Name], prop.Name)
			}
		}
	}

	history := []string{}
	exiting := false
	for {
		in := prompt.Input(">>> ", func(d prompt.Document) []prompt.Suggest {
			current := strings.ToUpper(strings.TrimSpace(d.CurrentLineBeforeCursor()))
			if i := strings.LastIndexByte(current, ';'); i != -1 {
				current = current[i+1:]
			}

			if _, ss, isFrom := strings.Cut(current, "FROM"); isFrom {
				ss = strings.TrimSpace(ss)
				suggests := make([]prompt.Suggest, 0, len(properties))
				for kind := range properties {
					suggests = append(suggests, prompt.Suggest{Text: kind})
				}
				return prompt.FilterHasPrefix(suggests, ss, true)
			}
			return nil
		}, prompt.OptionHistory(history))
		if in == "" {
			if exiting {
				return nil
			}
			exiting = true
			continue
		} else if in == "quit" || in == "exit" {
			return nil
		}

		history = append(history, in)
		queries := strings.Split(in, ";")
		for _, query := range queries {
			if strings.TrimSpace(query) == "" {
				continue
			}

			if err := r.execute(ctx, client, qp, query); err != nil {
				log.Println(err)
			}
		}
	}
}

func (r *GQLCommand) execute(ctx context.Context, client *datastore.Client, qp *parser.QueryParser, query string) error {
	q, aq, err := qp.ParseGQL(query)
	if err != nil {
		return err
	}
	if aq != nil {
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

	iter := client.Run(ctx, q)
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
			if err := encoder.Encode(datastore.FromDatastoreKey(key)); err != nil {
				return err
			}
		} else {
			if err := encoder.Encode(entity); err != nil {
				return err
			}
		}
	}
	return nil
}
