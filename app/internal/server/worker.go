package server

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/adriein/soma/app/internal"
	"github.com/adriein/soma/app/internal/customer"
	"github.com/adriein/soma/app/pkg/vendor"
)

type Worker struct {
	logger       *slog.Logger
	ch           <-chan vendor.TelegramUpdate
	customerServ customer.CustomerService
}

func NewWorker(app *internal.App, ch <-chan vendor.TelegramUpdate) *Worker {
	return &Worker{
		logger:       app.Logger,
		ch:           ch,
		customerServ: app.Modules.Customer,
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

			w.handleUpdate(ctx, update)
		}
	}
}

func (w *Worker) handleUpdate(ctx context.Context, update vendor.TelegramUpdate) {
	w.logger.Info("Dispatched update")
	switch update.Message.Text {
	case "/start":
		w.handleAuth(ctx, update)
	}
	fmt.Printf("Dispatched update: %+v\n", update)
}

func (w *Worker) handleAuth(ctx context.Context, update vendor.TelegramUpdate) error {
	w.customerServ.ConnectNutritionApp(ctx, update.Message.Chat.ID)

	return nil
}
