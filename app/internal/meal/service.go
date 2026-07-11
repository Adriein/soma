package meal

import (
	"context"
	"time"

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

	location, err := time.LoadLocation("Europe/Madrid")

	if err != nil {
		return nil, eris.Wrap(err, "Error loading location")
	}

	nowInMadrid := time.Now().In(location)

	from := nowInMadrid.AddDate(0, 0, -5)

	s.nutritionAPI.GetDiaryEntries(oauth, from)

	return nil, nil
}
