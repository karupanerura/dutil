package parser

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/karupanerura/datastore-cli/internal/datastore"
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
	if !strings.HasPrefix(src, "key(") {
		return nil, fmt.Errorf("invalid key: %q", src)
	}

	i := len("key(")
	key := &datastore.Key{Namespace: p.Namespace}
	for i < len(src) {
		// should not be last
		if i == len(src)-1 {
			return nil, fmt.Errorf("invalid key: %q", src)
		}

		// skip whitespace
		for src[i] == ' ' {
			i++
			if i == len(src) {
				return nil, fmt.Errorf("invalid key: %q", src)
			}
		}

		// quoted kind or raw kind literal
		if src[i] == '`' {
			j := strings.IndexByte(src[i+1:], '`')
			if j == -1 {
				return nil, fmt.Errorf("invalid key: %q", src)
			}
			key.Kind = src[i+1 : i+j+1]
			i = i + 1 + j + 1
			if i == len(src) || src[i] != ',' {
				return nil, fmt.Errorf("invalid key: %q", src)
			}
			i++
		} else {
			j := strings.IndexAny(src[i:], ", ")
			if j == -1 || src[j] == ' ' {
				return nil, fmt.Errorf("invalid key: %q", src)
			}
			key.Kind = src[i : i+j]
			i = i + j + 1
		}
		if i == len(src) {
			return nil, fmt.Errorf("invalid key: %q", src)
		}

		// skip whitespace
		for src[i] == ' ' {
			i++
			if i == len(src) {
				return nil, fmt.Errorf("invalid key: %q", src)
			}
		}

		// quoted key literal
		if src[i] == '"' || src[i] == '\'' {
			quote := src[i]
			j := strings.IndexByte(src[i+1:], quote)
			if j == -1 {
				return nil, fmt.Errorf("invalid key: %q", src)
			}
			key.Name = src[i+1 : i+j+1]
			i = i + 1 + j + 1
			if i == len(src) {
				return nil, fmt.Errorf("invalid key: %q", src)
			}
		} else {
			// numeric key literal
			j := strings.IndexFunc(src[i:], func(r rune) bool {
				isDigit := '0' <= r && r <= '9'
				return !isDigit
			})
			if j == 0 || j == -1 {
				return nil, fmt.Errorf("invalid key: %q", src)
			}
			var err error
			key.ID, err = strconv.ParseInt(src[i:i+j], 10, 64)
			if err != nil {
				return nil, err
			}
			i = i + j
			if i == len(src) {
				return nil, fmt.Errorf("invalid key: %q", src)
			}
			if key.ID == 0 {
				return nil, fmt.Errorf("invalid key: %q (id must not be 0)", src)
			}
		}

		// skip whitespace
		for src[i] == ' ' {
			i++
			if i == len(src) {
				return nil, fmt.Errorf("invalid key: %q", src)
			}
		}

		if src[i] == ',' {
			i++
			key = &datastore.Key{Namespace: p.Namespace, Parent: key}
			continue
		} else if src[i] == ')' {
			i++
			if i != len(src) {
				return nil, fmt.Errorf("invalid key: %q", src)
			}
		}
	}
	return key, nil
}
