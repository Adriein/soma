package worker

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/adriein/soma/app/internal/customer"
	"github.com/adriein/soma/app/pkg/vendor"
	"github.com/rotisserie/eris"
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
	w.logger.Debug(fmt.Sprintf("Dispatched update: %+v\n", update))

	switch update.Message.Text {
	case "/start":
		err := w.handleConnect(ctx, update)

		if err != nil {
			w.logger.Error(eris.ToString(err, true))
		}
	case "/auth":
		err := w.handleExchangeToken(ctx, update)

		if err != nil {
			w.logger.Error(eris.ToString(err, true))
		}
	case "/assessment":
		err := w.handleAssessment(ctx)

		if err != nil {
			w.logger.Error(eris.ToString(err, true))
		}
	}
	fmt.Printf("Dispatched update: %+v\n", update)
}

func (w *Worker) handleConnect(ctx context.Context, update vendor.TelegramUpdate) error {
	return w.customerServ.ConnectNutritionApp(ctx, update.Message.Chat.ID, update.Message.From.FirstName)
}

func (w *Worker) handleExchangeToken(ctx context.Context, update vendor.TelegramUpdate) error {
	return w.customerServ.ExchangeToken(ctx, update.Message.Chat.ID, update.Message.Text)
}

func (w *Worker) handleAssessment(ctx context.Context) error {
	return nil
}
