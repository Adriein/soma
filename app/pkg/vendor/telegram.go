package vendor

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/adriein/soma/app/pkg/constants"
	"github.com/rotisserie/eris"
)

const (
	TelegramApiURL              = "https://api.telegram.org/"
	TelegramApiGetUpdatesMethod = "getUpdates"
)

type TelegramUser struct {
	ID        int    `json:"id"`
	FirstName string `json:"first_name"`
	Username  string `json:"username"`
}

type TelegramMessage struct {
	ID       int          `json:"message_id"`
	ThreadID int          `json:"message_thread_id"`
	From     TelegramUser `json:"from"`
	Text     string       `json:"text"`
	Date     int          `json:"date"`
}
type TelegramUpdate struct {
	ID      int             `json:"update_id"`
	Message TelegramMessage `json:"message"`
	Err     error
}

type TelegramBot struct {
	ch  chan<- TelegramUpdate
	url string
}

func NewTelegramBot(ch chan<- TelegramUpdate) *TelegramBot {
	token := os.Getenv(constants.TelegramBotApiToken)
	url := fmt.Sprintf("%s/%s/%s", TelegramApiURL, token, TelegramApiGetUpdatesMethod)

	return &TelegramBot{
		ch:  ch,
		url: url,
	}
}

func (b *TelegramBot) send(ctx context.Context, update TelegramUpdate) {
	select {
	case b.ch <- update:
	case <-ctx.Done():
	}
}

func (b *TelegramBot) Poll(ctx context.Context) {
	if ctx.Err() != nil {
		return
	}

	req, err := http.NewRequestWithContext(ctx, "GET", b.url, nil)

	if err != nil {
		b.send(ctx, TelegramUpdate{Err: eris.Wrap(err, "Failed creating the request")})
		return
	}

	res, err := http.DefaultClient.Do(req)

	if err != nil {
		b.send(ctx, TelegramUpdate{Err: eris.Wrap(err, "Failed doing the request")})
		return
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		b.send(ctx, TelegramUpdate{Err: eris.Errorf("Error with status code %d", res.StatusCode)})
		return
	}

	var update TelegramUpdate
	err = json.NewDecoder(res.Body).Decode(&update)
	if err != nil {
		b.send(ctx, TelegramUpdate{Err: eris.Wrap(err, "Error decoding json")})
		return
	}

	b.send(ctx, update)
}
