package convert_test

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/karupanerura/dutil/internal/command"
	"github.com/karupanerura/dutil/internal/command/convert"
)

func TestKeyCommand_Run(t *testing.T) {
	t.Parallel()

	for _, from := range []string{"json", "gql", "encoded", "proto"} {
		from := from

		input := getExpected(from)
		t.Run("From"+strings.ToUpper(from), func(t *testing.T) {
			t.Parallel()

			for _, to := range []string{"json", "gql", "encoded", "proto"} {
				to := to

				t.Run("To"+strings.ToUpper(to), func(t *testing.T) {
					t.Parallel()
					stdin := strings.NewReader(input)
					expected := getExpected(to)

					stdout := strings.Builder{}
					cmd := &convert.KeyCommand{From: from, To: to}
					if err := cmd.Run(t.Context(), command.GlobalOptions{Stdin: stdin, Stdout: &stdout}); err != nil {
						t.Fatal(err)
					}

					if df := cmp.Diff(expected, strings.TrimSuffix(stdout.String(), "\n")); df != "" {
						t.Errorf("expected convert: %s -> %s: %s", from, to, df)
					}
				})
			}
		})

		t.Run("AutoFrom"+strings.ToUpper(from), func(t *testing.T) {
			t.Parallel()

			for _, to := range []string{"json", "gql", "encoded", "proto"} {
				t.Run("To"+strings.ToUpper(to), func(t *testing.T) {
					t.Parallel()

					stdin := strings.NewReader(input)
					expected := getExpected(to)

					stdout := strings.Builder{}
					cmd := &convert.KeyCommand{From: "auto", To: to}
					if err := cmd.Run(t.Context(), command.GlobalOptions{Stdin: stdin, Stdout: &stdout}); err != nil {
						t.Fatal(err)
					}

					if df := cmp.Diff(expected, strings.TrimSuffix(stdout.String(), "\n")); df != "" {
						t.Errorf("expected convert key: %s -> %s: %s", from, to, df)
					}
				})
			}
		})
	}
}

func getExpected(format string) string {
	switch format {
	case "gql":
		return strings.Join([]string{
			`KEY(Foo,"FooName1")`,
			`KEY(Bar,1)`,
			`KEY(Foo,"FooName2",Baz,"BazName2")`,
		}, "\n")
	case "json":
		return strings.Join([]string{
			`{"kind":"Foo","name":"FooName1"}`,
			`{"kind":"Bar","id":1}`,
			`{"kind":"Baz","name":"BazName2","parent":{"kind":"Foo","name":"FooName2"}}`,
		}, "\n")
	case "encoded":
		return strings.Join([]string{
			"Eg8KA0ZvbxoIRm9vTmFtZTE",
			"EgcKA0JhchAB",
			"Eg8KA0ZvbxoIRm9vTmFtZTISDwoDQmF6GghCYXpOYW1lMg",
		}, "\n")
	case "proto":
		// workaround for unstable proto encoder when debugging or cover
		if testing.CoverMode() != "" {
			return strings.Join([]string{
				"path%3A++%7B%0A++kind%3A++%22Foo%22%0A++name%3A++%22FooName1%22%0A%7D%0A",
				"path%3A++%7B%0A++kind%3A++%22Bar%22%0A++id%3A++1%0A%7D%0A",
				"path%3A++%7B%0A++kind%3A++%22Foo%22%0A++name%3A++%22FooName2%22%0A%7D%0Apath%3A++%7B%0A++kind%3A++%22Baz%22%0A++name%3A++%22BazName2%22%0A%7D%0A",
			}, "\n")
		}
		return strings.Join([]string{
			"path%3A+%7B%0A++kind%3A+%22Foo%22%0A++name%3A+%22FooName1%22%0A%7D%0A",
			"path%3A+%7B%0A++kind%3A+%22Bar%22%0A++id%3A+1%0A%7D%0A",
			"path%3A+%7B%0A++kind%3A+%22Foo%22%0A++name%3A+%22FooName2%22%0A%7D%0Apath%3A+%7B%0A++kind%3A+%22Baz%22%0A++name%3A+%22BazName2%22%0A%7D%0A",
		}, "\n")
	default:
		panic("unknown format: " + format)
	}
}
