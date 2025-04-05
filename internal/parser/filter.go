package parser

import (
	"fmt"

	"github.com/karupanerura/dutil/internal/datastore"
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

	return p.convertCondition(parsed.Normalize())
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

	case *gqlparser.ForwardComparatorCondition:
		if c.Comparator == gqlparser.HasAncestorForwardComparator {
			if c.Property != "__key__" {
				panic("HAS ANCESTOR is only valid for __key__")
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
		if values, isSlice := c.Value.([]any); isSlice {
			if c.Property == "__key__" {
				keys := make([]any, len(values))
				for i, v := range values {
					key, ok := v.(*gqlparser.Key)
					if !ok {
						return nil, datastore.PropertyFilter{}, fmt.Errorf("__key__ comparator value must be a key")
					}
					keys[i] = p.convertKey(key).ToDatastore()
				}
				values = keys
			}
			switch c.Comparator {
			case gqlparser.EqualsEitherComparator:
				return nil, datastore.PropertyFilter{
					FieldName: c.Property,
					Operator:  "in",
					Value:     values,
				}, nil
			case gqlparser.NotEqualsEitherComparator:
				return nil, datastore.PropertyFilter{
					FieldName: c.Property,
					Operator:  "not-in",
					Value:     values,
				}, nil
			default:
				// not a special case, so do following code.
			}
		}

		value := c.Value
		if c.Property == "__key__" {
			key, ok := c.Value.(*gqlparser.Key)
			if !ok {
				return nil, datastore.PropertyFilter{}, fmt.Errorf("__key__ comparator value must be a key")
			}

			value = p.convertKey(key).ToDatastore()
		}
		if value == nil {
			// workaround: IS NULL filter will be rejected
			switch c.Comparator {
			case gqlparser.EqualsEitherComparator:
				return nil, datastore.PropertyFilter{
					FieldName: c.Property,
					Operator:  "in",
					Value:     []any{nil},
				}, nil
			case gqlparser.NotEqualsEitherComparator:
				return nil, datastore.PropertyFilter{
					FieldName: c.Property,
					Operator:  "not-in",
					Value:     []any{nil},
				}, nil
			default:
				return nil, nil, fmt.Errorf("unsupported comparator with NULL: %v", c.Comparator)
			}
		}
		return nil, datastore.PropertyFilter{
			FieldName: c.Property,
			Operator:  string(c.Comparator),
			Value:     value,
		}, nil
	default:
		panic(fmt.Sprintf("unknown condition: %T", c))
	}
}

func convertForwardComparator(c gqlparser.ForwardComparator) string {
	switch c {
	case gqlparser.InForwardComparator:
		return "in"
	case gqlparser.NotInForwardComparator:
		return "not-in"
	default:
		panic(fmt.Sprintf("unknown forward comparator: %s", c))
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
