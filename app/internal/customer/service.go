package customer

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/adriein/soma/app/pkg/vendor"
	"github.com/rotisserie/eris"
)

type CustomerService interface {
	GetCustomer(ctx context.Context, chatID int64) (*Customer, error)
	ConnectNutritionApp(ctx context.Context, chatID int64, customerName string) error
	ExchangeToken(ctx context.Context, chatID int64, tokenVerifier string) error
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

func (s *Service) GetCustomer(ctx context.Context, chatID int64) (*Customer, error) {
	customer, err := s.repo.GetByTelegramChatID(ctx, chatID)

	if err != nil {
		return nil, eris.Wrap(err, "Error fetching customer")
	}

	return customer, nil
}

func (s *Service) ConnectNutritionApp(ctx context.Context, chatID int64, customerName string) error {
	customer, err := s.repo.GetByTelegramChatID(ctx, chatID)

	if customer != nil {
		var markdown strings.Builder

		greetings := fmt.Sprintf("👋 *Hello %s\\!*\n", customer.Name)
		info := "You have already connected *FatSecret*, to proceed just type /assessment\\."

		markdown.WriteString(greetings)
		markdown.WriteString(info)

		payload := vendor.OutgoingMessage{
			ChatID:    chatID,
			Text:      markdown.String(),
			ParseMode: "MarkdownV2",
		}

		if err := s.bot.SendMessage(ctx, payload); err != nil {
			return eris.Wrap(err, "Error sending msg to telegram")
		}

		return nil
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

	authURL, err := s.nutritionDiaryAPI.AuthorizeToken(oauth)

	if err != nil {
		return eris.Wrap(err, "Error authorizing the token")
	}

	var markdown strings.Builder

	greetings := fmt.Sprintf("👋 *Hello %s, welcome to Soma\\!*\n\n", customer.Name)
	info := "Let's sync your nutrition data\\! To get started, we just need to connect your *FatSecret* account\\.\n\n"
	instructions := "🔐 Tap the *Authorize* button below to get your secure code\\.\n"
	nextSteps := "Once you have your code, come back here and reply with: `/auth \\<your_code\\>`\\.\n\n"
	footer := "_You will be safely redirected to the official FatSecret authorization page\\._"

	markdown.WriteString(greetings)
	markdown.WriteString(info)
	markdown.WriteString(instructions)
	markdown.WriteString(nextSteps)
	markdown.WriteString(footer)

	payload := vendor.OutgoingMessage{
		ChatID:    chatID,
		Text:      markdown.String(),
		ParseMode: "MarkdownV2",
		ReplyMarkup: &vendor.InlineKeyboardMarkup{
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

	if err := s.repo.Save(ctx, customer); err != nil {
		return eris.Wrap(err, "Error saving the customer")
	}

	return nil
}

func (s *Service) ExchangeToken(ctx context.Context, chatID int64, tokenVerifier string) error {
	customer, err := s.repo.GetByTelegramChatID(ctx, chatID)

	if err != nil {
		return eris.Wrap(err, "Error fetching the customer")
	}

	verifier, err := strconv.Atoi(tokenVerifier)

	if err != nil {
		return eris.Wrap(err, "Error converting verifier into an int")
	}

	payload := &vendor.OAuth{
		OAuthToken:       customer.Token,
		OAuthTokenSecret: customer.TokenSecret,
		OauthVerifyCode:  verifier,
	}

	oauth, err := s.nutritionDiaryAPI.VerifyToken(payload)

	if err != nil {
		return eris.Wrap(err, "Error exchanging the token")
	}

	customer.Token = oauth.OAuthToken
	customer.TokenVerifier = verifier

	if err := s.repo.Save(ctx, customer); err != nil {
		return eris.Wrap(err, "Error updating the customer")
	}

	return nil
}
