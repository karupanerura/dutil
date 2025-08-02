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
}
