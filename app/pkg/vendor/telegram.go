package vendor

import "context"

type TelegramUpdate struct{}

type TelegramBot struct {
	ch chan<- TelegramUpdate
}

func NewTelegramBot(ch chan<- TelegramUpdate) *TelegramBot {
	return &TelegramBot{
		ch: ch,
	}
}

func (b *TelegramBot) Poll(ctx context.Context) {

}
