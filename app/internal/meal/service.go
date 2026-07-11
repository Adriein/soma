package meal

import (
	"context"
	"time"

	"github.com/adriein/soma/app/pkg/vendor"
	"github.com/rotisserie/eris"
)

type MealService interface {
	Get(ctx context.Context, oauth *vendor.OAuth, days int) ([]*Meal, error)
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

func (s *Service) Get(ctx context.Context, oauth *vendor.OAuth, days int) ([]*Meal, error) {
	location, err := time.LoadLocation("Europe/Madrid")

	if err != nil {
		return nil, eris.Wrap(err, "Error loading location")
	}

	nowInMadrid := time.Now().In(location)

	from := nowInMadrid.AddDate(0, 0, -days)

	s.nutritionAPI.GetDiaryEntries(oauth, from)

	return nil, nil
}
