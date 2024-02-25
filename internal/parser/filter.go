package parser

import (
	"fmt"

	"github.com/karupanerura/datastore-cli/internal/datastore"
)

type FilterParser struct {
	Namespace string
}

// ParseFilter parses GQL compound-condition
func (p *FilterParser) ParseFilter(s string) (datastore.EntityFilter, error) {
	return nil, fmt.Errorf("filter parser is not yet implemented")
}
