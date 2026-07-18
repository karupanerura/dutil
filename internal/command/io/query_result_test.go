package io

import (
	"encoding/json"
	"strings"
	"testing"

	clouddatastore "cloud.google.com/go/datastore"
	"github.com/google/go-cmp/cmp"

	"github.com/karupanerura/dutil/internal/datastore"
	"github.com/karupanerura/dutil/internal/parser"
)

func TestQueryCommand_EmptyPropertyEntityOutput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		keysOnly bool
		want     string
	}{
		{
			name:     "normal query",
			keysOnly: false,
			want:     "{\"key\":{\"kind\":\"Task\",\"name\":\"empty\"}}\n",
		},
		{
			name:     "keys-only query",
			keysOnly: true,
			want:     "{\"kind\":\"Task\",\"name\":\"empty\"}\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cmd := QueryCommand{KeysOnly: tt.keysOnly}
			got := renderEmptyPropertyEntity(t, cmd.KeysOnly)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("unexpected output (-want +got):\n%s", diff)
			}
		})
	}
}

func TestGQLCommand_EmptyPropertyEntityOutput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		query string
		want  string
	}{
		{
			name:  "SELECT __key__",
			query: "SELECT __key__ FROM Task",
			want:  "{\"kind\":\"Task\",\"name\":\"empty\"}\n",
		},
		{
			name:  "SELECT *",
			query: "SELECT * FROM Task",
			want:  "{\"key\":{\"kind\":\"Task\",\"name\":\"empty\"}}\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			query, keysOnly, aggregationQuery, err := (&parser.QueryParser{}).ParseGQL(tt.query)
			if err != nil {
				t.Fatalf("ParseGQL() error = %v", err)
			}
			if query == nil || aggregationQuery != nil {
				t.Fatalf("ParseGQL() returned query=%v, aggregationQuery=%v, want a non-aggregation query", query, aggregationQuery)
			}

			got := renderEmptyPropertyEntity(t, keysOnly)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("unexpected output (-want +got):\n%s", diff)
			}
		})
	}
}

func renderEmptyPropertyEntity(t *testing.T, keysOnly bool) string {
	t.Helper()

	key := clouddatastore.NameKey("Task", "empty", nil)
	entity := datastore.Entity{Key: datastore.FromDatastoreKey(key)}
	var stdout strings.Builder
	if err := writeQueryResult(
		&stdout,
		json.NewEncoder(&stdout),
		datastore.KeyFormatter{Format: "json"},
		key,
		entity,
		keysOnly,
	); err != nil {
		t.Fatalf("writeQueryResult() error = %v", err)
	}
	return stdout.String()
}
