package parser

import (
	"testing"

	clouddatastore "cloud.google.com/go/datastore"
	"github.com/google/go-cmp/cmp"
	internaldatastore "github.com/karupanerura/dutil/internal/datastore"
)

func TestFilterParserParseFilterKeyLiterals(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		namespace     string
		filter        string
		wantAncestor  *internaldatastore.Key
		wantFilterKey any
	}{
		{
			name:   "equality comparison with multi-element key",
			filter: `__key__ = KEY(ParentKind, "parent", ChildKind, "child")`,
			wantFilterKey: &clouddatastore.Key{Kind: "ChildKind", Name: "child", Parent: &clouddatastore.Key{
				Kind: "ParentKind", Name: "parent",
			}},
		},
		{
			name:   "IN comparison with multi-element key",
			filter: `__key__ IN (KEY(ParentKind, "parent", ChildKind, "child"))`,
			wantFilterKey: []any{&clouddatastore.Key{Kind: "ChildKind", Name: "child", Parent: &clouddatastore.Key{
				Kind: "ParentKind", Name: "parent",
			}}},
		},
		{
			name:   "HAS ANCESTOR",
			filter: `__key__ HAS ANCESTOR KEY(ParentKind, "parent")`,
			wantAncestor: &internaldatastore.Key{
				Kind: "ParentKind", Name: "parent",
			},
		},
		{
			name:      "namespace fallback",
			namespace: "default-ns",
			filter:    `__key__ = KEY(ParentKind, "parent", ChildKind, "child")`,
			wantFilterKey: &clouddatastore.Key{Kind: "ChildKind", Name: "child", Namespace: "default-ns", Parent: &clouddatastore.Key{
				Kind: "ParentKind", Name: "parent", Namespace: "default-ns",
			}},
		},
		{
			name:      "explicit namespace override",
			namespace: "default-ns",
			filter:    `__key__ = KEY(NAMESPACE("explicit-ns"), ParentKind, "parent", ChildKind, "child")`,
			wantFilterKey: &clouddatastore.Key{Kind: "ChildKind", Name: "child", Namespace: "explicit-ns", Parent: &clouddatastore.Key{
				Kind: "ParentKind", Name: "parent", Namespace: "explicit-ns",
			}},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ancestor, filter, err := (&FilterParser{Namespace: tt.namespace}).ParseFilter(tt.filter)
			if err != nil {
				t.Fatalf("ParseFilter() error = %v", err)
			}

			if diff := cmp.Diff(tt.wantAncestor, ancestor, cmp.AllowUnexported(internaldatastore.Key{})); diff != "" {
				t.Fatalf("ancestor mismatch (-want +got):\n%s", diff)
			}

			if tt.wantFilterKey == nil {
				return
			}

			propertyFilter, ok := filter.(clouddatastore.PropertyFilter)
			if !ok {
				t.Fatalf("filter = %T, want datastore.PropertyFilter", filter)
			}
			if diff := cmp.Diff(tt.wantFilterKey, propertyFilter.Value, cmp.AllowUnexported(clouddatastore.Key{})); diff != "" {
				t.Fatalf("filter value mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
