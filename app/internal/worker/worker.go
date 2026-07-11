package worker

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/adriein/soma/app/internal/coach"
	"github.com/adriein/soma/app/internal/customer"
	"github.com/adriein/soma/app/pkg/vendor"
	"github.com/rotisserie/eris"
)

type Worker struct {
	logger       *slog.Logger
	telegram     vendor.Bot
	ch           chan vendor.TelegramUpdate
	customerServ customer.CustomerService
	coachServ    coach.CoachService
}

func New(
	customerServ customer.CustomerService,
	coachServ coach.CoachService,
	logger *slog.Logger,
	telegram vendor.Bot,
) *Worker {

	ch := make(chan vendor.TelegramUpdate, 20)

	return &Worker{
		logger:       logger,
		telegram:     telegram,
		customerServ: customerServ,
		coachServ:    coachServ,
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

	args := strings.Fields(update.Message.Text)

	if len(args) == 0 {
		w.logger.Error("No command provided")

		return
	}

	command := args[0]

	switch command {
	case "/start":
		err := w.handleConnect(ctx, update)

		if err != nil {
			w.logger.Error("Failed connection", "error_details", eris.ToJSON(err, true))
		}
	case "/auth":
		err := w.handleExchangeToken(ctx, update, args)

		if err != nil {
			w.logger.Error(eris.ToString(err, true))
		}
	case "/assessment":
		err := w.handleAssessment(ctx, update)

		if err != nil {
			w.logger.Error(eris.ToString(err, true))
		}
	default:
		w.logger.Error(fmt.Sprintf("Command %s not implemented", command))
	}
}

func (w *Worker) handleConnect(ctx context.Context, update vendor.TelegramUpdate) error {
	return w.customerServ.ConnectNutritionApp(ctx, update.Message.Chat.ID, update.Message.From.FirstName)
}

func (w *Worker) handleExchangeToken(ctx context.Context, update vendor.TelegramUpdate, args []string) error {
	verificator := args[1]

	return w.customerServ.ExchangeToken(ctx, update.Message.Chat.ID, verificator)
}

func (w *Worker) handleAssessment(ctx context.Context, update vendor.TelegramUpdate) error {
	w.coachServ.Assessment(ctx, update.Message.Chat.ID)

	return nil
}
