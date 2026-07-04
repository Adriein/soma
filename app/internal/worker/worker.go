package worker

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/adriein/soma/app/internal/customer"
	"github.com/adriein/soma/app/pkg/vendor"
)

type Worker struct {
	logger       *slog.Logger
	customerServ customer.CustomerService
	telegram     vendor.Bot
	ch           chan vendor.TelegramUpdate
}

func New(customerServ customer.CustomerService, logger *slog.Logger, telegram vendor.Bot) *Worker {
	ch := make(chan vendor.TelegramUpdate, 20)

	return &Worker{
		logger:       logger,
		telegram:     telegram,
		customerServ: customerServ,
		ch:           ch,
	}
}

func (w *Worker) Start(ctx context.Context) {
	go w.Dispatch(ctx, w.ch)

	go w.telegram.Poll(ctx, w.ch)
}

func (w *Worker) Dispatch(ctx context.Context, ch <-chan vendor.TelegramUpdate) {
	for {
		select {
		case <-ctx.Done():
			return

		case update, ok := <-ch:
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
