package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"slices"
	"syscall"

	"github.com/alecthomas/kong"
	"github.com/karupanerura/datastore-cli/internal/command"
	"github.com/karupanerura/datastore-cli/internal/version"
)

type CLI struct {
	command.Options
	Lookup command.LookupCommand `cmd:""`
	Query  command.QueryCommand  `cmd:""`
	Upsert command.UpsertCommand `cmd:""`
	Delete command.DeleteCommand `cmd:""`
	GQL    command.GQLCommand    `cmd:""`
}

func main() {
	if slices.Contains(os.Args[1:], "--version") {
		fmt.Println(version.Value)
		return
	}

	var opts CLI
	c := kong.Parse(&opts)
	c.FatalIfErrorf(c.Error)

	ctx := context.Background()
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP)
	defer stop()

	c.BindTo(ctx, (*context.Context)(nil))
	c.BindTo(opts.Options, (*command.Options)(nil))
	c.FatalIfErrorf(c.Run())
}
