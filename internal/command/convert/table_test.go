package convert_test

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/karupanerura/dutil/internal/command"
	"github.com/karupanerura/dutil/internal/command/convert"
)

func TestTableCommand_Run(t *testing.T) {
	t.Parallel()

	t.Run("FromKey", func(t *testing.T) {
		t.Parallel()

		stdin := strings.NewReader(strings.Join([]string{
			`{"kind":"Foo","name":"FooName1"}`,
			`{"kind":"Foo","name":"FooName2"}`,
			`{"kind":"Bar","id":1}`,
			`{"kind":"Bar","id":2}`,
			`{"kind":"Bar","id":3}`,
			`{"kind":"Baz","name":"BazName1","parent":{"kind":"Foo","name":"FooName1"}}`,
			`{"kind":"Baz","name":"BazName2","parent":{"kind":"Foo","name":"FooName1"}}`,
			`{"kind":"Baz","name":"BazName1","parent":{"kind":"Foo","name":"FooName2"}}`,
			`{"kind":"Baz","name":"BazName2","parent":{"kind":"Foo","name":"FooName2"}}`,
		}, "\n"))

		stdout := strings.Builder{}
		cmd := &convert.TableCommand{From: "key"}
		if err := cmd.Run(t.Context(), command.GlobalOptions{Stdin: stdin, Stdout: &stdout}); err != nil {
			t.Fatal(err)
		}

		expected := `+------+----------+
| Kind | Name     |
+------+----------+
| Foo  | FooName1 |
| Foo  | FooName2 |
+------+----------+
+------+----+
| Kind | ID |
+------+----+
| Bar  |  1 |
| Bar  |  2 |
| Bar  |  3 |
+------+----+
+------+----------+------+----------+
| Kind | Name     | Kind | Name     |
+------+----------+------+----------+
| Foo  | FooName1 | Baz  | BazName1 |
| Foo  | FooName1 | Baz  | BazName2 |
| Foo  | FooName2 | Baz  | BazName1 |
| Foo  | FooName2 | Baz  | BazName2 |
+------+----------+------+----------+
`
		if df := cmp.Diff(expected, stdout.String()); df != "" {
			t.Errorf("expected table convert from key: %s", df)
		}
	})

	t.Run("FromMissingEntity", func(t *testing.T) {
		t.Parallel()

		stdin := strings.NewReader("null\n")
		stdout := strings.Builder{}
		cmd := &convert.TableCommand{From: "entity"}
		if err := cmd.Run(t.Context(), command.GlobalOptions{Stdin: stdin, Stdout: &stdout}); err != nil {
			t.Fatal(err)
		}

		if df := cmp.Diff("", stdout.String()); df != "" {
			t.Errorf("expected missing entity to be skipped: %s", df)
		}
	})

	t.Run("FromEntityWithKeylessEmbeddedEntity", func(t *testing.T) {
		t.Parallel()

		stdin := strings.NewReader(`{"key":{"kind":"Foo","name":"foo"},"properties":[{"name":"child","type":"entity","value":[{"name":"value","type":"string","value":"bar"}]}]}`)
		stdout := strings.Builder{}
		cmd := &convert.TableCommand{From: "entity"}
		if err := cmd.Run(t.Context(), command.GlobalOptions{Stdin: stdin, Stdout: &stdout}); err != nil {
			t.Fatal(err)
		}

		expected := `+------+------+-------------+
| Kind | Name | child.value |
+------+------+-------------+
| Foo  | foo  | bar         |
+------+------+-------------+
`
		if df := cmp.Diff(expected, stdout.String()); df != "" {
			t.Errorf("expected table convert from keyless embedded entity: %s", df)
		}
	})

	t.Run("FromEntityWithKeyedEmbeddedEntity", func(t *testing.T) {
		t.Parallel()

		stdin := strings.NewReader(`{"key":{"kind":"Foo","name":"foo"},"properties":[{"name":"child","type":"entity","value":{"key":{"kind":"Child","name":"child-key"},"properties":[{"name":"value","type":"string","value":"bar"}]}}]}`)
		stdout := strings.Builder{}
		cmd := &convert.TableCommand{From: "entity"}
		if err := cmd.Run(t.Context(), command.GlobalOptions{Stdin: stdin, Stdout: &stdout}); err != nil {
			t.Fatal(err)
		}

		expected := `+------+------+------------------------+-------------+
| Kind | Name | child.__key__          | child.value |
+------+------+------------------------+-------------+
| Foo  | foo  | KEY(Child,"child-key") | bar         |
+------+------+------------------------+-------------+
`
		if df := cmp.Diff(expected, stdout.String()); df != "" {
			t.Errorf("expected table convert from keyed embedded entity: %s", df)
		}
	})
}
