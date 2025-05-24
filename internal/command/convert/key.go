package convert

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"log"
	"os"
	"strings"

	"github.com/karupanerura/dutil/internal/command"
	"github.com/karupanerura/dutil/internal/datastore"
	"github.com/karupanerura/dutil/internal/parser"
)

type KeyCommand struct {
	From string `name:"from" enum:"json,gql,encoded,auto" default:"auto" help:"Source key format"`
	To   string `name:"to" enum:"json,gql,encoded" default:"encoded" help:"Result key format"`
}

func (r *KeyCommand) Run(ctx context.Context, opts command.GlobalOptions) error {
	reader := newKeyReader(r.From, os.Stdin)
	writer := newKeyWriter(r.To, os.Stdout)
	for {
		key, err := reader.Read()
		if err == io.EOF {
			return writer.Flush()
		} else if err != nil {
			return err
		}

		err = writer.Write(key)
		if err != nil {
			return err
		}
	}
}

type keyReader interface {
	Read() (*datastore.Key, error)
}

func newKeyReader(format string, reader io.Reader) keyReader {
	switch format {
	case "json":
		return &jsonKeyReader{decoder: json.NewDecoder(reader)}
	case "gql":
		return &gqlKeyReader{reader: bufio.NewReader(reader)}
	case "encoded":
		return &encodedKeyReader{reader: bufio.NewReader(reader)}
	case "auto":
		return newKeyReader(detectFormat(reader))
	default:
		panic("unknown format:" + format)
	}
}

func detectFormat(reader io.Reader) (string, io.Reader) {
	var buffer [4]byte
	if _, err := io.ReadFull(reader, buffer[:]); err != nil {
		log.Panicf("Failed to read to detect key format: %+v", err)
	}

	header := string(buffer[:])
	if header[0] == '{' {
		return "json", io.MultiReader(strings.NewReader(header), reader)
	}
	if header[3] == '(' {
		return "gql", io.MultiReader(strings.NewReader(header), reader)
	}
	return "encoded", io.MultiReader(strings.NewReader(header), reader)
}

type jsonKeyReader struct {
	decoder *json.Decoder
}

func (r *jsonKeyReader) Read() (key *datastore.Key, err error) {
	err = r.decoder.Decode(&key)
	return
}

type gqlKeyReader struct {
	reader    *bufio.Reader
	keyParser *parser.KeyParser
}

func (r *gqlKeyReader) Read() (*datastore.Key, error) {
	line, _, err := r.reader.ReadLine()
	if err != nil {
		return nil, err
	}
	return r.keyParser.ParseKey(string(line))
}

type encodedKeyReader struct {
	reader *bufio.Reader
}

func (r *encodedKeyReader) Read() (*datastore.Key, error) {
	line, _, err := r.reader.ReadLine()
	if err != nil {
		return nil, err
	}
	return datastore.DecodeKey(string(line))
}

type keyWriter interface {
	Write(*datastore.Key) error
	Flush() error
}

func newKeyWriter(format string, writer io.Writer) keyWriter {
	switch format {
	case "json":
		return &jsonKeyWriter{encoder: json.NewEncoder(writer)}
	case "gql":
		return &gqlKeyWriter{writer: bufio.NewWriter(writer)}
	case "encoded":
		return &encodedKeyWriter{writer: bufio.NewWriter(writer)}
	default:
		panic("unknown format: " + format)
	}
}

type jsonKeyWriter struct {
	encoder *json.Encoder
}

func (w *jsonKeyWriter) Write(key *datastore.Key) error {
	return w.encoder.Encode(key)
}

func (w *jsonKeyWriter) Flush() error {
	return nil
}

type ioByteStringWriter interface {
	io.Writer
	io.ByteWriter
}

type gqlKeyWriter struct {
	writer *bufio.Writer
}

func (w *gqlKeyWriter) Write(key *datastore.Key) error {
	_, err := w.writer.WriteString(key.String())
	if err != nil {
		return err
	}
	return w.writer.WriteByte('\n')
}

func (w *gqlKeyWriter) Flush() error {
	return w.writer.Flush()
}

type encodedKeyWriter struct {
	writer *bufio.Writer
}

func (w *encodedKeyWriter) Write(key *datastore.Key) error {
	_, err := w.writer.WriteString(key.ToDatastore().Encode())
	if err != nil {
		return err
	}
	return w.writer.WriteByte('\n')
}

func (w *encodedKeyWriter) Flush() error {
	return w.writer.Flush()
}
