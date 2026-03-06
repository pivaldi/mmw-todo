package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/pivaldi/mmw/todo"
)

func main() {
	// 1. Create a root context that cancels on SIGINT or SIGTERM
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	app := todo.New()

	// 2. Clean up resources when the application exits
	defer app.Close()

	// 3. Pass the context down the chain
	if err := app.Bootstrap(ctx, nil); err != nil {
		panic(err)
	}

	if err := app.Run(ctx); err != nil {
		panic(err)
	}
}
