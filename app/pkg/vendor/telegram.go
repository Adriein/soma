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
	TelegramApiURL               = "https://api.telegram.org"
	TelegramApiGetUpdatesMethod  = "getUpdates"
	TelegramApiSendMessageMethod = "sendMessage"
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

type IncomingMessage struct {
	ID       int          `json:"message_id"`
	ThreadID int          `json:"message_thread_id"`
	From     TelegramUser `json:"from"`
	Chat     TelegramChat `json:"chat"`
	Text     string       `json:"text"`
	Date     int          `json:"date"`
}
type TelegramUpdate struct {
	ID      int             `json:"update_id"`
	Message IncomingMessage `json:"message"`
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

type InlineKeyboardMarkup struct {
	InlineKeyboard [][]InlineKeyboardButton `json:"inline_keyboard"`
}

type InlineKeyboardButton struct {
	Text string `json:"text"`
	Url  string `json:"url"`
}

type OutgoingMessage struct {
	ChatID      int64                `json:"chat_id"`
	Text        string               `json:"text"`
	ReplyMarkup InlineKeyboardMarkup `json:"reply_markup"`
}

type Bot interface {
	Poll(ctx context.Context, ch chan<- TelegramUpdate)
	SendMessage(ctx context.Context, payload OutgoingMessage) error
}

type TelegramBot struct {
	url string
}

func NewTelegramBot() *TelegramBot {
	token := os.Getenv(constants.TelegramBotApiToken)
	url := fmt.Sprintf("%s/bot%s", TelegramApiURL, token)

	return &TelegramBot{
		url: url,
	}
}

func (b *TelegramBot) notify(ctx context.Context, ch chan<- TelegramUpdate, update TelegramUpdate) {
	select {
	case ch <- update:
	case <-ctx.Done():
	}
}

func (b *TelegramBot) Poll(ctx context.Context, ch chan<- TelegramUpdate) {
	messageOffset := 0

	for {
		if ctx.Err() != nil {
			return
		}

		payload := GetUpdatesPayload{
			Timeout: 30,
			Offset:  messageOffset,
		}

		body, err := json.Marshal(payload)

		if err != nil {
			b.notify(ctx, ch, TelegramUpdate{Err: eris.Wrap(err, "Failed encoding payload")})
			time.Sleep(time.Second * 2)
			continue
		}

		url := fmt.Sprintf("%s/%s", b.url, TelegramApiGetUpdatesMethod)

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))

		if err != nil {
			b.notify(ctx, ch, TelegramUpdate{Err: eris.Wrap(err, "Failed creating the request")})
			time.Sleep(time.Second * 2)
			return
		}

		req.Header.Set("Content-Type", "application/json")

		res, err := http.DefaultClient.Do(req)

		if err != nil {
			b.notify(ctx, ch, TelegramUpdate{Err: eris.Wrap(err, "Failed doing the request")})
			time.Sleep(time.Second * 2)
			return
		}

		if res.StatusCode != http.StatusOK {
			b.notify(ctx, ch, TelegramUpdate{Err: eris.Errorf("Error with status code %d", res.StatusCode)})
			res.Body.Close()
			time.Sleep(time.Second * 2)
			return
		}

		var response GetUpdatesRes
		err = json.NewDecoder(res.Body).Decode(&response)
		res.Body.Close()

		if err != nil {
			b.notify(ctx, ch, TelegramUpdate{Err: eris.Wrap(err, "Error decoding json")})
			res.Body.Close()
			time.Sleep(time.Second * 2)
			return
		}

		for _, update := range response.Result {
			b.notify(ctx, ch, update)
			messageOffset = update.ID + 1
		}
	}
}

func (b *TelegramBot) SendMessage(ctx context.Context, payload OutgoingMessage) error {
	body, err := json.Marshal(payload)

	if err != nil {
		return eris.Wrap(err, "Error marshalling payload")
	}

	url := fmt.Sprintf("%s/%s", b.url, TelegramApiSendMessageMethod)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))

	if err != nil {
		return eris.Wrap(err, "Error creating the request")
	}

	req.Header.Set("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)

	if err != nil {
		return eris.Wrap(err, "Error doing the request")
	}

	if res.StatusCode != http.StatusOK {
		return eris.Wrap(err, "Error, http status is not 200")
	}

	return nil
}
