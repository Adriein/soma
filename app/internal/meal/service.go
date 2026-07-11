package meal

import (
	"context"

	"github.com/adriein/soma/app/internal/customer"
	"github.com/adriein/soma/app/pkg/vendor"
	"github.com/rotisserie/eris"
)

type MealService interface {
	Get(ctx context.Context, chatID int64) ([]*Meal, error)
}

type Service struct {
	nutritionAPI vendor.NutritionDiary
	customerServ customer.CustomerService
}

func NewService(
	nutritionAPI vendor.NutritionDiary,
	customerServ customer.CustomerService,
) *Service {
	return &Service{
		nutritionAPI: nutritionAPI,
		customerServ: customerServ,
	}
}

func (s *Service) Get(ctx context.Context, chatID int64) ([]*Meal, error) {
	customer, err := s.customerServ.GetCustomer(ctx, chatID)

	if err != nil {
		return nil, eris.Wrap(err, "Error fetching customer")
	}

	oauth := &vendor.OAuth{
		OAuthToken:       customer.Token,
		OAuthTokenSecret: customer.TokenSecret,
		OauthVerifyCode:  customer.TokenVerifier,
	}

	s.nutritionAPI.GetDiaryEntries(oauth)

	return nil, nil
}
