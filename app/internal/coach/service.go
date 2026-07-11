package coach

import (
	"context"
	"fmt"

	"github.com/adriein/soma/app/internal/meal"
	"github.com/adriein/soma/app/pkg/vendor"
	"github.com/rotisserie/eris"
)

type CoachService interface {
	Assessment(ctx context.Context, chatID int64) (*ActionPlan, error)
}

type Service struct {
	mealServ meal.MealService
	aiServ   vendor.AI
}

func NewService(mealServ meal.MealService, aiServ vendor.AI) *Service {
	return &Service{
		mealServ: mealServ,
		aiServ:   aiServ,
	}
}

func (s *Service) Assessment(ctx context.Context, chatID int64) (*ActionPlan, error) {
	meals, err := s.mealServ.Get(ctx, chatID)

	if err != nil {
		return nil, eris.Wrap(err, "Assessment error getting meals")
	}

	if meals == nil {
		return nil, eris.New("Assessment error, no meals found")
	}

	aiAssessment, err := s.aiServ.Ask("")

	if err != nil {
		return nil, eris.Wrap(err, "Assesment error calling AI")
	}

	fmt.Print(aiAssessment)

	return nil, nil
}
