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

func TestFilterParserParseFilterInvalid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		filter string
	}{
		{
			name:   "HAS ANCESTOR on non-key property",
			filter: `name HAS ANCESTOR KEY(ParentKind, "parent")`,
		},
		{
			name:   "HAS ANCESTOR with non-key value",
			filter: `__key__ HAS ANCESTOR "parent"`,
		},
		{
			name:   "multiple ancestor conditions",
			filter: `__key__ HAS ANCESTOR KEY(A, "a") AND __key__ HAS ANCESTOR KEY(B, "b")`,
		},
		{
			name:   "ancestor condition inside OR",
			filter: `__key__ HAS ANCESTOR KEY(ParentKind, "parent") OR status = "active"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if _, _, err := (&FilterParser{}).ParseFilter(tt.filter); err == nil {
				t.Fatalf("ParseFilter(%q) error = nil, want error", tt.filter)
			}
		})
	}
}

func TestFilterParserParseFilterAncestorConjunction(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		filter       string
		wantAncestor *internaldatastore.Key
		wantFilter   clouddatastore.EntityFilter
	}{
		{
			name:   "HAS ANCESTOR AND property filter",
			filter: `__key__ HAS ANCESTOR KEY(ParentKind, "parent") AND status = "active"`,
			wantAncestor: &internaldatastore.Key{
				Kind: "ParentKind", Name: "parent",
			},
			wantFilter: clouddatastore.PropertyFilter{
				FieldName: "status", Operator: "=", Value: "active",
			},
		},
		{
			name:   "property filter AND HAS ANCESTOR",
			filter: `status = "active" AND __key__ HAS ANCESTOR KEY(ParentKind, "parent")`,
			wantAncestor: &internaldatastore.Key{
				Kind: "ParentKind", Name: "parent",
			},
			wantFilter: clouddatastore.PropertyFilter{
				FieldName: "status", Operator: "=", Value: "active",
			},
		},
		{
			name:   "nested conjunction with one ancestor constraint",
			filter: `(__key__ HAS ANCESTOR KEY(ParentKind, "parent") AND a = 1) AND b = 2`,
			wantAncestor: &internaldatastore.Key{
				Kind: "ParentKind", Name: "parent",
			},
			wantFilter: clouddatastore.AndFilter{
				Filters: []clouddatastore.EntityFilter{
					clouddatastore.PropertyFilter{FieldName: "a", Operator: "=", Value: int64(1)},
					clouddatastore.PropertyFilter{FieldName: "b", Operator: "=", Value: int64(2)},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ancestor, filter, err := (&FilterParser{}).ParseFilter(tt.filter)
			if err != nil {
				t.Fatalf("ParseFilter() error = %v", err)
			}

			if diff := cmp.Diff(tt.wantAncestor, ancestor, cmp.AllowUnexported(internaldatastore.Key{})); diff != "" {
				t.Fatalf("ancestor mismatch (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.wantFilter, filter); diff != "" {
				t.Fatalf("filter mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestQueryParserParseGQLAncestorConjunction(t *testing.T) {
	t.Parallel()

	query, _, aggregationQuery, err := (&QueryParser{}).ParseGQL(`
		SELECT * FROM Child
		WHERE __key__ HAS ANCESTOR KEY(Parent, "parent") AND status = "active"
	`)
	if err != nil {
		t.Fatalf("ParseGQL() error = %v", err)
	}
	if query == nil || aggregationQuery != nil {
		t.Fatalf("ParseGQL() returned query=%v, aggregationQuery=%v, want a non-aggregation query", query, aggregationQuery)
	}
}
