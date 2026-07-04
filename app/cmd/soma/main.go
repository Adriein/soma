package main

import (
	"context"
	"fmt"
	"os"

	"github.com/adriein/soma/app/internal"
	"github.com/adriein/soma/app/internal/server"
	"github.com/adriein/soma/app/pkg/constants"
	_ "github.com/lib/pq"
)

func main() {
	app := internal.NewApp()

	if len(os.Args) < 2 {
		worker := app.Modules.Worker
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		worker.Start(ctx)

		server.New(os.Getenv(constants.ServerPort))

		return
	}

	switch os.Args[1] {
	default:
		fmt.Printf("Unknown command: %s\n", os.Args[1])
		os.Exit(1)
	}
}
