package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/alecthomas/kong"
	"github.com/karupanerura/datastore-cli/internal/command"
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
