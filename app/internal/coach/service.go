package coach

import (
	"github.com/adriein/soma/app/internal/meal"
	"github.com/rotisserie/eris"
)

type CoachService interface {
	Assessment() *ActionPlan
}

type Service struct {
	mealService meal.MealService
}

func NewService() *Service {
	return &Service{}
}

func (s *Service) Assessment() (*ActionPlan, error) {
	meals, err := s.mealService.Get()

	if err != nil {
		return nil, eris.Wrap(err, "Assessment error getting meals")
	}

	if meals == nil {
		return nil, eris.New("Assessment error, no meals found")
	}

	return nil, nil
}
