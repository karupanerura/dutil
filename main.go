package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"reflect"
	"syscall"

	"github.com/alecthomas/kong"
	"github.com/karupanerura/dutil/internal/command"
	"github.com/karupanerura/dutil/internal/command/convert"
	iocommand "github.com/karupanerura/dutil/internal/command/io"
	"github.com/karupanerura/dutil/internal/version"
)

type CLI struct {
	command.GlobalOptions
	IO      iocommand.Commands `cmd:""`
	Convert convert.Commands   `cmd:""`
}

func main() {
	var kongOptions = []kong.Option{
		kong.Name("dutil"),
		kong.NamedMapper("stdin", getFileReaderMapper(os.Stdin)),
		kong.NamedMapper("stdout", getFileWriterMapper(os.Stdout)),
		kong.NamedMapper("stderr", getFileWriterMapper(os.Stderr)),
	}

	var opts CLI
	c := kong.Parse(&opts, kongOptions...)
	c.FatalIfErrorf(c.Error)

	if opts.Version {
		fmt.Println(version.Name)
		return
	}

	// change default logger outputs
	if opts.Stderr != os.Stderr {
		log.Default().SetOutput(opts.Stderr)
	}

	ctx := context.Background()
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP)
	defer stop()

	c.BindTo(ctx, (*context.Context)(nil))
	c.BindTo(opts.GlobalOptions, (*command.GlobalOptions)(nil))
	c.FatalIfErrorf(c.Run())
}

func getFileReaderMapper(def io.Reader) kong.Mapper {
	return kong.MapperFunc(func(ctx *kong.DecodeContext, target reflect.Value) error {
		var path string
		err := ctx.Scan.PopValueInto("reader", &path)
		if err != nil {
			return err
		}

		var reader io.Reader
		if path == "-" {
			reader = def
		} else {
			path = kong.ExpandPath(path)
			reader, err = os.Open(path)
			if err != nil {
				return err
			}
		}
		target.Set(reflect.ValueOf(reader))
		return nil
	})
}

func getFileWriterMapper(def io.Writer) kong.Mapper {
	return kong.MapperFunc(func(ctx *kong.DecodeContext, target reflect.Value) error {
		var path string
		err := ctx.Scan.PopValueInto("writer", &path)
		if err != nil {
			return err
		}

		var writer io.Writer
		if path == "-" {
			writer = def
		} else {
			path = kong.ExpandPath(path)
			writer, err = os.OpenFile(path, os.O_WRONLY, os.ModeType)
			if err != nil {
				return err
			}
		}
		target.Set(reflect.ValueOf(writer))
		return nil
	})
}
