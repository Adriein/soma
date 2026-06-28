package vendor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/adriein/soma/app/pkg/constants"
	"github.com/rotisserie/eris"
)

const (
	TelegramApiURL              = "https://api.telegram.org"
	TelegramApiGetUpdatesMethod = "getUpdates"
)

type TelegramUser struct {
	ID        int    `json:"id"`
	FirstName string `json:"first_name"`
	Username  string `json:"username"`
}

type TelegramChat struct {
	ID   int64  `json:"id"`
	Type string `json:"type"`
}

type TelegramMessage struct {
	ID       int          `json:"message_id"`
	ThreadID int          `json:"message_thread_id"`
	From     TelegramUser `json:"from"`
	Chat     TelegramChat `json:"chat"`
	Text     string       `json:"text"`
	Date     int          `json:"date"`
}
type TelegramUpdate struct {
	ID      int             `json:"update_id"`
	Message TelegramMessage `json:"message"`
	Err     error
}

type GetUpdatesPayload struct {
	Offset  int `json:"offset,omitempty"`
	Timeout int `json:"timeout,omitempty"`
}

type GetUpdatesRes struct {
	Ok     bool             `json:"ok"`
	Result []TelegramUpdate `json:"result"`
}

type TelegramBot struct {
	ch  chan<- TelegramUpdate
	url string
}

func NewTelegramBot(ch chan<- TelegramUpdate) *TelegramBot {
	token := os.Getenv(constants.TelegramBotApiToken)
	url := fmt.Sprintf("%s/bot%s/%s", TelegramApiURL, token, TelegramApiGetUpdatesMethod)

	return &TelegramBot{
		ch:  ch,
		url: url,
	}
}

func (b *TelegramBot) notify(ctx context.Context, update TelegramUpdate) {
	select {
	case b.ch <- update:
	case <-ctx.Done():
	}
}

func (b *TelegramBot) Poll(ctx context.Context) {
	for {
		if ctx.Err() != nil {
			return
		}

		payload := GetUpdatesPayload{
			Timeout: 30,
			Offset:  1,
		}

		body, err := json.Marshal(payload)

		if err != nil {
			b.notify(ctx, TelegramUpdate{Err: eris.Wrap(err, "Failed encoding payload")})
			time.Sleep(time.Second * 2)
			continue
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, b.url, bytes.NewReader(body))

		if err != nil {
			b.notify(ctx, TelegramUpdate{Err: eris.Wrap(err, "Failed creating the request")})
			time.Sleep(time.Second * 2)
			return
		}

		req.Header.Set("Content-Type", "application/json")

		res, err := http.DefaultClient.Do(req)

		if err != nil {
			b.notify(ctx, TelegramUpdate{Err: eris.Wrap(err, "Failed doing the request")})
			time.Sleep(time.Second * 2)
			return
		}

		if res.StatusCode != http.StatusOK {
			b.notify(ctx, TelegramUpdate{Err: eris.Errorf("Error with status code %d", res.StatusCode)})
			res.Body.Close()
			time.Sleep(time.Second * 2)
			return
		}

		var response GetUpdatesRes
		err = json.NewDecoder(res.Body).Decode(&response)
		res.Body.Close()

		if err != nil {
			b.notify(ctx, TelegramUpdate{Err: eris.Wrap(err, "Error decoding json")})
			res.Body.Close()
			time.Sleep(time.Second * 2)
			return
		}

		for _, update := range response.Result {
			b.notify(ctx, update)
		}

	}
}
