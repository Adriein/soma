package server

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/adriein/soma/app/pkg/vendor"
)

type Worker struct {
	logger *slog.Logger
	ch     <-chan vendor.TelegramUpdate
}

func NewWorker(logger *slog.Logger, ch <-chan vendor.TelegramUpdate) *Worker {
	return &Worker{
		logger: logger,
		ch:     ch,
	}
}

func (w *Worker) Dispatch(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return

		case update, ok := <-w.ch:
			if !ok {
				w.logger.Info("Channel closed. Exiting dispatcher loop.")
				return
			}

			w.handleUpdate(update)
		}
	}
}

func (w *Worker) handleUpdate(update vendor.TelegramUpdate) {
	fmt.Printf("Dispatched update: %+v\n", update)
}
