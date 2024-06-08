package main

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"

	"github.com/alecthomas/kong"
	"github.com/karupanerura/dutil/internal/command"
	"github.com/karupanerura/dutil/internal/command/io"
	"github.com/karupanerura/dutil/internal/version"
)

type CLI struct {
	command.GlobalOptions
	IO io.Commands `cmd:""`
}

func main() {
	var opts CLI
	c := kong.Parse(&opts)
	c.FatalIfErrorf(c.Error)

	if opts.Version {
		fmt.Println(version.Name)
		return
	}

	ctx := context.Background()
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP)
	defer stop()

	c.BindTo(ctx, (*context.Context)(nil))
	c.BindTo(opts.GlobalOptions, (*command.GlobalOptions)(nil))
	c.FatalIfErrorf(c.Run())
}
