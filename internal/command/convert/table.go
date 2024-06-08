package convert

import (
	"cmp"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"slices"
	"strconv"
	"time"

	"github.com/karupanerura/dutil/internal/command"
	"github.com/karupanerura/dutil/internal/datastore"
	"github.com/syohex/go-texttable"
)

type TableCommand struct {
	From string `name:"from" short:"f" enum:"key,entity,explain" default:"entity" help:"Type of JSON structure to convert to table"`
}

func (r *TableCommand) Run(ctx context.Context, opts command.GlobalOptions) error {
	var iterate func(*json.Decoder) func(func(tableEntry, error) bool)
	switch r.From {
	case "key":
		iterate = keyToTableEntryIter
	case "entity":
		iterate = entityToTableEntryReader
	case "explain":
		iterate = explainToTableEntryReader
	}

	decoder := json.NewDecoder(os.Stdin)
	for table, err := range buildTables(iterate(decoder)) {
		if err != nil {
			return err
		}
		fmt.Println(table.Draw())
	}
	return nil
}


type tableEntry struct {
	Header []string
	Row    []string
}

func (e *tableEntry) push(e2 tableEntry) {
	e.Header = append(e.Header, e2.Header...)
	e.Row = append(e.Row, e2.Row...)
}

type tableBuilder struct {
	header []string
	table  *texttable.TextTable
}

func buildTables(iter func(func(tableEntry, error) bool)) func(func(*texttable.TextTable, error) bool) {
	return func(yield func(*texttable.TextTable, error) bool) {
		b := tableBuilder{}
		for entry, err := range iter {
			if err != nil {
				yield(nil, err)
				return
			}
			if table := b.buildTable(entry); table != nil {
				if !yield(table, nil) {
					break
				}
			}
		}
		if b.table != nil {
			yield(b.table, nil)
		}
	}
}

func (b *tableBuilder) buildTable(e tableEntry) *texttable.TextTable {
	if b.table == nil {
		b.initTable(e)
		return nil
	}

	if slices.Equal(b.header, e.Header) {
		b.table.AddRow(e.Row...)
		return nil
	} else {
		completed := b.table
		b.initTable(e)
		return completed
	}
}

func (b *tableBuilder) initTable(e tableEntry) {
	b.table = &texttable.TextTable{}
	b.table.SetHeader(e.Header...)
	b.table.AddRow(e.Row...)
	b.header = e.Header
}

type jsonReader[T any] struct {
	decoder *json.Decoder
}

func (r *jsonReader[T]) Iter() func(func(T, error) bool) {
	return func(yield func(T, error) bool) {
		for {
			var v T
			if err := r.decoder.Decode(&v); errors.Is(err, io.EOF) {
				return
			} else if err != nil {
				var zero T
				yield(zero, err)
				return
			}
			if !yield(v, nil) {
				break
			}
		}
	}
}

func keyToTableEntryIter(decoder *json.Decoder) func(func(tableEntry, error) bool) {
	return func(yield func(tableEntry, error) bool) {
		reader := &jsonReader[*datastore.Key]{decoder: decoder}
		for key, err := range reader.Iter() {
			if err != nil {
				yield(tableEntry{}, err)
				break
			}
			if !yield(keyToTableEntry(key), nil) {
				break
			}
		}
	}
}

func keyToTableEntry(key *datastore.Key) (entry tableEntry) {
	for key != nil {
		var value string
		if key.ID != 0 {
			entry.Header = append(entry.Header, "ID", "Kind")
			value = strconv.FormatInt(key.ID, 10)
		} else if key.Name != "" {
			entry.Header = append(entry.Header, "Name", "Kind")
			value = key.Name
		}
		if key.Namespace != "" {
			entry.Header = append(entry.Header, "Namespace")
			entry.Row = append(entry.Row, key.Namespace)
		}
		entry.Row = append(entry.Row, value, key.Kind)
		key = key.Parent
	}
	slices.Reverse(entry.Header)
	slices.Reverse(entry.Row)
	return
}

func entityToTableEntryReader(decoder *json.Decoder) func(func(tableEntry, error) bool) {
	return func(yield func(tableEntry, error) bool) {
		reader := &jsonReader[*datastore.Entity]{decoder: decoder}
		for entity, err := range reader.Iter() {
			if err != nil {
				yield(tableEntry{}, err)
				break
			}

			var entry tableEntry
			if entity.Key != nil {
				entry = keyToTableEntry(entity.Key)
			}
			appendToTableEntryFromProperties(&entry, "", entity.Properties)
			if entity.Metadata != nil {
				entry.Header = append(entry.Header, "Version", "CreateTime", "UpdateTime")
				entry.Row = append(entry.Row, strconv.FormatInt(entity.Metadata.Version, 10), entity.Metadata.CreateTime.String(), entity.Metadata.UpdateTime.String())
			}
			if !yield(entry, nil) {
				break
			}
		}
	}
}

func appendToTableEntryFromProperties(entry *tableEntry, prefix string, props []datastore.Property) {
	slices.SortFunc(props, func(lhs, rhs datastore.Property) int {
		return cmp.Compare(lhs.Name, rhs.Name)
	})
	for _, prop := range props {
		appendToTableEntryFromProperty(entry, prefix, prop)
	}
}

func appendToTableEntryFromProperty(entry *tableEntry, prefix string, prop datastore.Property) {
	v := prop.Value
	switch v.Type {
	case datastore.ArrayType:
		prefix := prefix + prop.Name
		for i, v := range v.Value.([]datastore.Value) {
			appendToTableEntryFromProperty(entry, prefix, datastore.Property{
				Name:  "[" + strconv.Itoa(i) + "]",
				Value: v,
			})
		}
		return

	case datastore.BlobType:
		b64 := base64.RawURLEncoding.EncodeToString(v.Value.([]byte))
		entry.Header = append(entry.Header, prefix+prop.Name)
		entry.Row = append(entry.Row, "b64"+strconv.Quote(b64))
		return

	case datastore.BoolType:
		entry.Header = append(entry.Header, prefix+prop.Name)
		entry.Row = append(entry.Row, strconv.FormatBool(v.Value.(bool)))
		return

	case datastore.TimestampType:
		entry.Header = append(entry.Header, prefix+prop.Name)
		entry.Row = append(entry.Row, v.Value.(time.Time).String())
		return

	case datastore.EntityType:
		appendToTableEntryFromProperties(entry, prop.Name+".", v.Value.([]datastore.Property))
		return

	case datastore.FloatType:
		entry.Header = append(entry.Header, prefix+prop.Name)
		entry.Row = append(entry.Row, strconv.FormatFloat(v.Value.(float64), 'f', -1, 64))
		return

	case datastore.GeoPointType:
		geo := v.Value.(datastore.GeoPoint)
		entry.Header = append(entry.Header, prefix+prop.Name)
		entry.Row = append(entry.Row, fmt.Sprintf("geo(lat: %f, lng: %f)", geo.Lat, geo.Lng))
		return

	case datastore.IntType:
		entry.Header = append(entry.Header, prefix+prop.Name)
		entry.Row = append(entry.Row, strconv.FormatInt(v.Value.(int64), 10))
		return

	case datastore.KeyType:
		v := v.Value.(*datastore.Key)
		entry.Header = append(entry.Header, prefix+prop.Name)
		entry.Row = append(entry.Row, v.String())
		return

	case datastore.NullType:
		entry.Header = append(entry.Header, prefix+prop.Name)
		entry.Row = append(entry.Row, "NULL")
		return

	case datastore.StringType:
		entry.Header = append(entry.Header, prefix+prop.Name)
		entry.Row = append(entry.Row, v.Value.(string))
		return

	default:
		panic("unknown type: " + v.Type)
	}
}

func explainToTableEntryReader(decoder *json.Decoder) func(func(tableEntry, error) bool) {
	return func(yield func(tableEntry, error) bool) {
		reader := &jsonReader[*datastore.ExplainMetrics]{decoder: decoder}
		for metrics, err := range reader.Iter() {
			if err != nil {
				yield(tableEntry{}, err)
				break
			}

			var entry tableEntry
			for i, indexUsed := range metrics.PlanSummary.IndexesUsed {
				prefix := "IndexesUsed[" + strconv.Itoa(i) + "]."
				if indexUsed != nil {
					for k, v := range *indexUsed {
						entry.Header = append(entry.Header, prefix+k)
						entry.Row = append(entry.Row, fmt.Sprint(v))
					}
				}
			}
			entry.Header = append(entry.Header, "ResultsReturned", "ExecutionDuration", "ReadOperations")
			entry.Row = append(entry.Row, strconv.Itoa(int(metrics.ExecutionStats.ResultsReturned)), metrics.ExecutionStats.ExecutionDuration.String(), strconv.Itoa(int(metrics.ExecutionStats.ReadOperations)))
			if metrics.ExecutionStats.DebugStats != nil {
				b, err := json.Marshal(metrics.ExecutionStats.DebugStats)
				if err != nil {
					yield(tableEntry{}, err)
					break
				}

				entry.Header = append(entry.Header, "DebugStats")
				entry.Row = append(entry.Row, string(b))
			}
			if !yield(entry, nil) {
				break
			}
		}
	}
}
