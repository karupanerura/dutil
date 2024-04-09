package parser

import (
	"fmt"

	"github.com/karupanerura/datastore-cli/internal/datastore"
	"github.com/karupanerura/gqlparser"
)

type FilterParser struct {
	Namespace string
}

// ParseFilter parses GQL compound-condition
func (p *FilterParser) ParseFilter(s string) (datastore.EntityFilter, error) {
	parsed, err := gqlparser.ParseCondition(gqlparser.NewLexer(s))
	if err != nil {
		return nil, fmt.Errorf("gqlparser.ParseCondition: %w", err)
	}

	return convertCondition(parsed), nil
}

func convertCondition(c gqlparser.Condition) datastore.EntityFilter {
	switch c := c.(type) {
	case *gqlparser.AndCompoundCondition:
		return datastore.AndFilter{
			Filters: []datastore.EntityFilter{
				convertCondition(c.Left),
				convertCondition(c.Right),
			},
		}

	case *gqlparser.OrCompoundCondition:
		return datastore.OrFilter{
			Filters: []datastore.EntityFilter{
				convertCondition(c.Left),
				convertCondition(c.Right),
			},
		}

	case *gqlparser.IsNullCondition:
		return datastore.PropertyFilter{
			FieldName: c.Property,
			Operator:  "=",
			Value:     nil,
		}

	case *gqlparser.ForwardComparatorCondition:
		return datastore.PropertyFilter{
			FieldName: c.Property,
			Operator:  convertForwardComparator(c.Comparator),
			Value:     c.Value,
		}

	case *gqlparser.EitherComparatorCondition:
		return datastore.PropertyFilter{
			FieldName: c.Property,
			Operator:  string(c.Comparator),
			Value:     c.Value,
		}

	case *gqlparser.BackwardComparatorCondition:
		return datastore.PropertyFilter{
			FieldName: c.Property,
			Operator:  convertBackwardComparator(c.Comparator),
			Value:     c.Value,
		}

	default:
		panic(fmt.Sprintf("unknown condition: %T", c))
	}
}

func convertForwardComparator(c gqlparser.ForwardComparator) string {
	switch c {
	case gqlparser.ContainsForwardComparator:
		return "="
	case gqlparser.HasAncestorForwardComparator:
		panic("HAS ANCESTOR is not supported, please use --ancestor option")
	case gqlparser.InForwardComparator:
		return "in"
	case gqlparser.NotInForwardComparator:
		return "not-in"
	default:
		panic(fmt.Sprintf("unknown forward comparator: %s", c))
	}
}

func convertBackwardComparator(c gqlparser.BackwardComparator) string {
	switch c {
	case gqlparser.InBackwardComparator:
		return "="
	case gqlparser.HasDescendantBackwardComparator:
		panic("HAS DESCENDANT is not supported, please use --ancestor option")
	default:
		panic(fmt.Sprintf("unknown backward comparator: %s", c))
	}
}
