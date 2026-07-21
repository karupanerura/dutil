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

		return ancestor, combineEntityFilters(leftFilter, rightFilter, func(filters []datastore.EntityFilter) datastore.EntityFilter {
			return datastore.AndFilter{Filters: filters}
		}), nil

	case *gqlparser.OrCompoundCondition:
		leftAncestor, leftFilter, err := p.convertCondition(c.Left)
		if err != nil {
			return nil, nil, err
		}

		rightAncestor, rightFilter, err := p.convertCondition(c.Right)
		if err != nil {
			return nil, nil, err
		}

		if leftAncestor != nil || rightAncestor != nil {
			return nil, nil, fmt.Errorf("HAS ANCESTOR is not valid within an OR condition")
		}

		return nil, combineEntityFilters(leftFilter, rightFilter, func(filters []datastore.EntityFilter) datastore.EntityFilter {
			return datastore.OrFilter{Filters: filters}
		}), nil

	case *gqlparser.ForwardComparatorCondition:
		if c.Comparator == gqlparser.HasAncestorForwardComparator {
			if c.Property.String() != "__key__" {
				return nil, nil, fmt.Errorf("HAS ANCESTOR is only valid for __key__")
			}
			key, ok := c.Value.(*gqlparser.Key)
			if !ok {
				return nil, nil, fmt.Errorf("HAS ANCESTOR value must be a key")
			}
			return p.convertKey(key), nil, nil
		}
		value := c.Value
		if c.Property.String() == "__key__" {
			values, ok := c.Value.([]any)
			if !ok {
				values = []any{c.Value}
			}
			keys := make([]any, len(values))
			for i, v := range values {
				key, ok := v.(*gqlparser.Key)
				if !ok {
					return nil, datastore.PropertyFilter{}, fmt.Errorf("__key__ comparator value must be a key")
				}
				keys[i] = p.convertKey(key).ToDatastore()
			}
			value = keys
		}
		operator, err := convertForwardComparator(c.Comparator)
		if err != nil {
			return nil, nil, err
		}
		return nil, datastore.PropertyFilter{
			FieldName: c.Property.String(),
			Operator:  operator,
			Value:     value,
		}, nil

	case *gqlparser.EitherComparatorCondition:
		if values, isSlice := c.Value.([]any); isSlice {
			if c.Property.String() == "__key__" {
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
					FieldName: c.Property.String(),
					Operator:  "in",
					Value:     values,
				}, nil
			case gqlparser.NotEqualsEitherComparator:
				return nil, datastore.PropertyFilter{
					FieldName: c.Property.String(),
					Operator:  "not-in",
					Value:     values,
				}, nil
			default:
				// not a special case, so do following code.
			}
		}

		value := c.Value
		if c.Property.String() == "__key__" {
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
					FieldName: c.Property.String(),
					Operator:  "in",
					Value:     []any{nil},
				}, nil
			case gqlparser.NotEqualsEitherComparator:
				return nil, datastore.PropertyFilter{
					FieldName: c.Property.String(),
					Operator:  "not-in",
					Value:     []any{nil},
				}, nil
			default:
				return nil, nil, fmt.Errorf("unsupported comparator with NULL: %v", c.Comparator)
			}
		}
		return nil, datastore.PropertyFilter{
			FieldName: c.Property.String(),
			Operator:  string(c.Comparator),
			Value:     value,
		}, nil
	default:
		return nil, nil, fmt.Errorf("unknown condition: %T", c)
	}
}

// combineEntityFilters merges two possibly-nil EntityFilters. A nil operand
// means "no filter contributed by that side" (e.g. a bare HAS ANCESTOR
// condition) and is dropped, regardless of why it is nil.
func combineEntityFilters(left, right datastore.EntityFilter, wrap func([]datastore.EntityFilter) datastore.EntityFilter) datastore.EntityFilter {
	switch {
	case left == nil:
		return right
	case right == nil:
		return left
	default:
		return wrap([]datastore.EntityFilter{left, right})
	}
}

func convertForwardComparator(c gqlparser.ForwardComparator) (string, error) {
	switch c {
	case gqlparser.InForwardComparator:
		return "in", nil
	case gqlparser.NotInForwardComparator:
		return "not-in", nil
	default:
		return "", fmt.Errorf("unknown forward comparator: %s", c)
	}
}

func (p *FilterParser) convertKey(src *gqlparser.Key) *datastore.Key {
	namespace := string(src.Namespace)
	if namespace == "" {
		namespace = p.Namespace
	}

	rootKey := &datastore.Key{Namespace: namespace}
	key := rootKey
	for i := len(src.Path) - 1; i >= 0; i-- {
		key.ID = src.Path[i].ID
		key.Name = src.Path[i].Name
		key.Namespace = namespace
		key.Kind = string(src.Path[i].Kind)
		if i != 0 {
			parent := &datastore.Key{Namespace: namespace}
			key.Parent = parent
			key = parent
		}
	}
	return rootKey
}
