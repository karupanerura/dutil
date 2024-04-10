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
func (p *FilterParser) ParseFilter(s string) (*datastore.Key, datastore.EntityFilter, error) {
	parsed, err := gqlparser.ParseCondition(gqlparser.NewLexer(s))
	if err != nil {
		return nil, nil, fmt.Errorf("gqlparser.ParseCondition: %w", err)
	}

	return p.convertCondition(parsed)
}

func (p *FilterParser) convertCondition(c gqlparser.Condition) (*datastore.Key, datastore.EntityFilter, error) {
	switch c := c.(type) {
	case *gqlparser.AndCompoundCondition:
		leftAncestor, leftFilter, err := p.convertCondition(c.Left)
		if err != nil {
			return nil, nil, err
		}

		rightAncestor, rightFilter, err := p.convertCondition(c.Right)
		if err != nil {
			return nil, nil, err
		}

		var ancestor *datastore.Key
		if leftAncestor != nil && rightAncestor != nil {
			return nil, nil, fmt.Errorf("multiple ancestor conditions are invalid")
		} else if leftAncestor != nil {
			ancestor = leftAncestor
		} else if rightAncestor != nil {
			ancestor = rightAncestor
		}

		return ancestor, datastore.AndFilter{
			Filters: []datastore.EntityFilter{leftFilter, rightFilter},
		}, nil

	case *gqlparser.OrCompoundCondition:
		leftAncestor, leftFilter, err := p.convertCondition(c.Left)
		if err != nil {
			return nil, nil, err
		}

		rightAncestor, rightFilter, err := p.convertCondition(c.Right)
		if err != nil {
			return nil, nil, err
		}

		var ancestor *datastore.Key
		if leftAncestor != nil && rightAncestor != nil {
			return nil, nil, fmt.Errorf("multiple ancestor conditions are invalid")
		} else if leftAncestor != nil {
			ancestor = leftAncestor
		} else if rightAncestor != nil {
			ancestor = rightAncestor
		}

		return ancestor, datastore.OrFilter{
			Filters: []datastore.EntityFilter{leftFilter, rightFilter},
		}, nil

	case *gqlparser.IsNullCondition:
		return nil, datastore.PropertyFilter{
			FieldName: c.Property,
			Operator:  "=",
			Value:     nil,
		}, nil

	case *gqlparser.ForwardComparatorCondition:
		if c.Comparator == gqlparser.HasAncestorForwardComparator {
			if c.Property != "__key__" {
				return nil, nil, fmt.Errorf("HAS ANCESTOR is only valid for __key__")
			}
			key, ok := c.Value.(*gqlparser.Key)
			if !ok {
				panic("HAS ANCESTOR value must be a key")
			}
			return p.convertKey(key), nil, nil
		}
		return nil, datastore.PropertyFilter{
			FieldName: c.Property,
			Operator:  convertForwardComparator(c.Comparator),
			Value:     c.Value,
		}, nil

	case *gqlparser.EitherComparatorCondition:
		return nil, datastore.PropertyFilter{
			FieldName: c.Property,
			Operator:  string(c.Comparator),
			Value:     c.Value,
		}, nil

	case *gqlparser.BackwardComparatorCondition:
		if c.Comparator == gqlparser.HasDescendantBackwardComparator {
			if c.Property != "__key__" {
				return nil, nil, fmt.Errorf("HAS ANCESTOR is only valid for __key__")
			}
			key, ok := c.Value.(*gqlparser.Key)
			if !ok {
				panic("HAS ANCESTOR value must be a key")
			}
			return p.convertKey(key), nil, nil
		}
		return nil, datastore.PropertyFilter{
			FieldName: c.Property,
			Operator:  convertBackwardComparator(c.Comparator),
			Value:     c.Value,
		}, nil

	default:
		panic(fmt.Sprintf("unknown condition: %T", c))
	}
}

func convertForwardComparator(c gqlparser.ForwardComparator) string {
	switch c {
	case gqlparser.ContainsForwardComparator:
		return "="
	case gqlparser.HasAncestorForwardComparator:
		panic("HAS ANCESTOR is not supported")
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
		panic("HAS DESCENDANT is not supported")
	default:
		panic(fmt.Sprintf("unknown backward comparator: %s", c))
	}
}

func (p *FilterParser) convertKey(src *gqlparser.Key) *datastore.Key {
	key := &datastore.Key{Namespace: p.Namespace}
	for i := len(src.Path) - 1; i >= 0; i-- {
		key.ID = src.Path[i].ID
		key.Name = src.Path[i].Name
		key.Namespace = string(src.Namespace)
		key.Kind = string(src.Path[i].Kind)
		if i != 0 {
			parent := &datastore.Key{Namespace: p.Namespace}
			key.Parent = parent
		}
	}
	return key
}
