package server

import (
	"context"

	"github.com/adriein/soma/app/pkg/vendor"
)

type Worker struct {
	ch chan<- vendor.TelegramUpdate
}

func NewWorker(ch chan<- vendor.TelegramUpdate) *Worker {
	return &Worker{
		ch: ch,
	}
}

func (w *Worker) Dispatch(ctx context.Context) {

}
