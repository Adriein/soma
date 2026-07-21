package meal

import (
	"context"
	"time"

	"github.com/adriein/soma/app/internal/customer"
	"github.com/adriein/soma/app/pkg/constants"
	"github.com/adriein/soma/app/pkg/vendor"
	"github.com/rotisserie/eris"
)

type MealService interface {
	Get(ctx context.Context, customer *customer.Customer, days int) ([]*Meal, error)
}

type Service struct {
	nutritionAPI vendor.NutritionDiary
}

func NewService(
	nutritionAPI vendor.NutritionDiary,
) *Service {
	return &Service{
		nutritionAPI: nutritionAPI,
	}
}

func (s *Service) Get(ctx context.Context, customer *customer.Customer, days int) ([]*Meal, error) {
	location, err := time.LoadLocation(constants.TimeLocationMadrid)

	if err != nil {
		return nil, eris.Wrap(err, "Error loading location")
	}

	nowInMadrid := time.Now().In(location)

	from := nowInMadrid.AddDate(0, 0, -days)

	oauth := &vendor.OAuth{
		OAuthToken:       customer.Token,
		OAuthTokenSecret: customer.TokenSecret,
		OauthVerifyCode:  customer.TokenVerifier,
	}

	result, err := s.nutritionAPI.GetDiaryEntries(oauth, from)

	if err != nil {
		return nil, eris.Wrap(err, "Error fetching meals")
	}

	meals := make([]*Meal, len(result.Entries.Meals))

	for idx, fsMeal := range result.Entries.Meals {
		meals[idx] = ToDomain(&fsMeal)
	}

	return meals, nil
}
