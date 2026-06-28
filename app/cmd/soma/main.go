package main

import (
	"context"
	"fmt"
	"os"

	"github.com/adriein/soma/app/internal"
	"github.com/adriein/soma/app/internal/server"
	"github.com/adriein/soma/app/pkg/constants"
	"github.com/adriein/soma/app/pkg/vendor"
	_ "github.com/lib/pq"
)

func main() {
	app := internal.NewApp()

	if len(os.Args) < 2 {
		ch := make(chan vendor.TelegramUpdate, 20)

		worker := server.NewWorker(app.Logger, ch)
		telegram := vendor.NewTelegramBot(ch)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		go worker.Dispatch(ctx)

		go telegram.Poll(ctx)

		server.New(os.Getenv(constants.ServerPort))

		return
	}

	switch os.Args[1] {
	default:
		fmt.Printf("Unknown command: %s\n", os.Args[1])
		os.Exit(1)
	}
}
