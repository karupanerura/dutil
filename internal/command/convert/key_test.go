package convert_test

import (
	"net/url"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/karupanerura/dutil/internal/command"
	"github.com/karupanerura/dutil/internal/command/convert"
	"github.com/karupanerura/dutil/internal/datastore"
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

					assertKeyOutput(t, to, expected, stdout.String())
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

					assertKeyOutput(t, to, expected, stdout.String())
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
		return strings.Join([]string{
			"path%3A+%7B%0A++kind%3A+%22Foo%22%0A++name%3A+%22FooName1%22%0A%7D%0A",
			"path%3A+%7B%0A++kind%3A+%22Bar%22%0A++id%3A+1%0A%7D%0A",
			"path%3A+%7B%0A++kind%3A+%22Foo%22%0A++name%3A+%22FooName2%22%0A%7D%0Apath%3A+%7B%0A++kind%3A+%22Baz%22%0A++name%3A+%22BazName2%22%0A%7D%0A",
		}, "\n")
	default:
		panic("unknown format: " + format)
	}
}

func assertKeyOutput(t *testing.T, format, expected, actual string) {
	t.Helper()

	if !strings.HasSuffix(actual, "\n") {
		t.Fatal("key output does not end with a newline")
	}
	actual = strings.TrimSuffix(actual, "\n")
	if format != "proto" {
		if df := cmp.Diff(expected, actual); df != "" {
			t.Errorf("unexpected key output (-want +got):\n%s", df)
		}
		return
	}

	expectedLines := strings.Split(expected, "\n")
	actualLines := strings.Split(actual, "\n")
	if len(actualLines) != len(expectedLines) {
		t.Fatalf("unexpected number of proto keys: got %d, want %d", len(actualLines), len(expectedLines))
	}

	for i := range expectedLines {
		assertProtoKey(t, i, expectedLines[i], actualLines[i])
	}
}

func assertProtoKey(t *testing.T, line int, expected, actual string) {
	t.Helper()

	decoded, err := url.QueryUnescape(actual)
	if err != nil {
		t.Fatalf("proto key line %d is not valid query encoding: %v", line+1, err)
	}
	if reencoded := url.QueryEscape(decoded); reencoded != actual {
		t.Errorf("proto key line %d is not canonically query encoded: got %q, want %q", line+1, actual, reencoded)
	}

	want, err := datastore.ParseEncodedProtoKey(expected)
	if err != nil {
		t.Fatalf("invalid expected proto key on line %d: %v", line+1, err)
	}
	got, err := datastore.ParseEncodedProtoKey(actual)
	if err != nil {
		t.Fatalf("invalid proto key output on line %d: %v", line+1, err)
	}
	if df := cmp.Diff(want, got); df != "" {
		t.Errorf("unexpected proto key on line %d (-want +got):\n%s", line+1, df)
	}
}
