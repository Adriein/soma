package customer

import (
	"context"
	"errors"

	"github.com/adriein/soma/app/pkg/vendor"
	"github.com/rotisserie/eris"
)

var CustomerAlreadyAuthorized = eris.New("Customer is already authorized")

type CustomerService interface {
	ConnectNutritionApp(ctx context.Context, chatID int64) (*string, error)
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

func (s *Service) ConnectNutritionApp(ctx context.Context, chatID int64) (*string, error) {
	customer, err := s.repo.GetByTelegramChatID(ctx, chatID)

	if customer != nil {
		return nil, CustomerAlreadyAuthorized
	}

	if err != nil && !errors.Is(err, ErrCustomerNotFound) {
		return nil, eris.Wrap(err, "Error getting customer by chat id")
	}

	oauth, err := s.nutritionDiaryAPI.GetToken()

	if err != nil {
		return nil, eris.Wrap(err, "Error getting the unauthorized token")
	}

	authURL, err := s.nutritionDiaryAPI.AuthorizeToken(oauth)

	if err != nil {
		return nil, eris.Wrap(err, "Error authorizing the token")
	}

	payload := vendor.OutgoingMessage{
		ChatID: chatID,
		Text:   "The following button redirects you to Fatsecret Auth page, you must give us permisions to read your food entries",
		ReplyMarkup: vendor.InlineKeyboardMarkup{
			InlineKeyboard: [][]vendor.InlineKeyboardButton{
				{
					{Text: "Authorize", Url: *authURL},
				},
			},
		},
	}

	s.bot.SendMessage(ctx, payload)

	return authURL, nil
}

func (s *Service) VerifyToken(tokenSecret string) error {
	return nil
}
