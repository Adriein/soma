package customer

import (
	"context"
	"errors"
	"fmt"

	"github.com/adriein/soma/app/pkg/vendor"
	"github.com/rotisserie/eris"
)

var CustomerAlreadyAuthorized = eris.New("Customer is already authorized")

type CustomerService interface {
	ConnectNutritionApp(ctx context.Context, chatID int64, customerName string) error
	VerifyToken(tokenSecret string) error
}

type Service struct {
	nutritionDiaryAPI vendor.NutritionDiary
	bot               vendor.Bot
	repo              CustomerRepository
}

func NewService(nutritionDiaryAPI vendor.NutritionDiary, bot vendor.Bot, repo CustomerRepository) *Service {
	return &Service{
		nutritionDiaryAPI: nutritionDiaryAPI,
		bot:               bot,
		repo:              repo,
	}
}

func (s *Service) ConnectNutritionApp(ctx context.Context, chatID int64, customerName string) error {
	customer, err := s.repo.GetByTelegramChatID(ctx, chatID)

	if customer != nil {
		return CustomerAlreadyAuthorized
	}

	if err != nil && !errors.Is(err, ErrCustomerNotFound) {
		return eris.Wrap(err, "Error getting customer by chat id")
	}

	oauth, err := s.nutritionDiaryAPI.GetToken()

	if err != nil {
		return eris.Wrap(err, "Error getting the unauthorized token")
	}

	customer = &Customer{
		Name:           customerName,
		TelegramChatID: chatID,
		Token:          oauth.OAuthToken,
		TokenSecret:    oauth.OAuthTokenSecret,
	}

	if err := s.repo.Save(ctx, customer); err != nil {
		return eris.Wrap(err, "Error saving the customer")
	}

	authURL, err := s.nutritionDiaryAPI.AuthorizeToken(oauth)

	if err != nil {
		return eris.Wrap(err, "Error authorizing the token")
	}

	text := fmt.Sprintf(
		`👋 *Hello %s, welcome to Soma\!*
		To automatically sync your nutrition data, we just need to connect your accounts\.
		🔐 Tap the *Authorize* button below to grant us permission to read your *FatSecret* food entries and paste the code into the DB\.
		_You will be safely redirected to the official FatSecret authorization page\._`,
		customer.Name,
	)

	payload := vendor.OutgoingMessage{
		ChatID:    chatID,
		Text:      text,
		ParseMode: "MarkdownV2",
		ReplyMarkup: vendor.InlineKeyboardMarkup{
			InlineKeyboard: [][]vendor.InlineKeyboardButton{
				{
					{Text: "Authorize", Url: *authURL},
				},
			},
		},
	}

	if err := s.bot.SendMessage(ctx, payload); err != nil {
		return eris.Wrap(err, "Error sending msg to telegram")
	}

	return nil
}

func (s *Service) VerifyToken(tokenSecret string) error {
	return nil
}
