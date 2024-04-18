package parser

import (
	"fmt"

	"github.com/karupanerura/datastore-cli/internal/datastore"
	"github.com/karupanerura/gqlparser"
)

type KeyParser struct {
	Namespace string
}

// ParseKey parses key(<kind>, <identifier>) or encoded keys.
// see also ParseKey.
func (p *KeyParser) ParseKeys(src []string) (dest datastore.Keys, err error) {
	dest = make(datastore.Keys, len(src))
	for i, key := range src {
		dest[i], err = p.ParseKey(key)
		if err != nil {
			return nil, err
		}
	}
	return
}

// ParseKey parses key(<kind>, <identifier>) or encoded key
// ref. https://support.google.com/cloud/answer/6361641
func (p *KeyParser) ParseKey(src string) (*datastore.Key, error) {
	if key, err := datastore.DecodeKey(src); err == nil {
		return key, nil
	}

	parsedKey, err := gqlparser.ParseKey(gqlparser.NewLexer(src))
	if err != nil {
		return nil, fmt.Errorf("gqlparser.ParseKey: %w", err)
	}

	rootKey := &datastore.Key{Kind: parsedKey.Namespace, Namespace: parsedKey.Namespace}
	if rootKey.Namespace == "" {
		rootKey.Namespace = p.Namespace
	}
	key := rootKey
	for i := len(parsedKey.Path) - 1; i >= 0; i-- {
		key.ID = parsedKey.Path[i].ID
		key.Name = parsedKey.Path[i].Name
		key.Kind = string(parsedKey.Path[i].Kind)
		if i != 0 {
			parent := &datastore.Key{Kind: parsedKey.Namespace, Namespace: parsedKey.Namespace}
			if parent.Namespace == "" {
				parent.Namespace = p.Namespace
			}
			key.Parent = parent
			key = parent
		}
	}
	return rootKey, nil
}
