package coach

import (
	"bytes"
	"context"
	"fmt"
	"text/template"
	"time"

	"github.com/adriein/soma/app/internal/customer"
	"github.com/adriein/soma/app/internal/meal"
	"github.com/adriein/soma/app/pkg/prompts"
	"github.com/adriein/soma/app/pkg/vendor"
	"github.com/rotisserie/eris"
)

const CoachTmpl = "coach"

type CoachService interface {
	Assessment(ctx context.Context, chatID int64) error
}

type Service struct {
	customerServ customer.CustomerService
	mealServ     meal.MealService
	aiServ       vendor.AI
}

func NewService(
	customerServ customer.CustomerService,
	mealServ meal.MealService,
	aiServ vendor.AI,
) *Service {
	return &Service{
		customerServ: customerServ,
		mealServ:     mealServ,
		aiServ:       aiServ,
	}
}

func (s *Service) Assessment(ctx context.Context, chatID int64) error {
	//TODO: the main idea is to get everything the first time and store the evaluation in the db with the date then the next times i only take from the last evaluation to today
	loc, err := time.LoadLocation("Europe/Madrid")

	if err != nil {
		return eris.Wrap(err, "Error loading location")
	}

	now := time.Now().In(loc)

	//The fist date I've added meals is 2026-06-25
	targetDate := time.Date(2026, time.June, 25, 0, 0, 0, 0, loc)

	dur := now.Sub(targetDate)

	daysElapsed := int(dur.Hours() / 24)

	var data AssessmentData

	customer, err := s.customerServ.GetCustomer(ctx, chatID)

	if err != nil {
		return eris.Wrap(err, "Error fetching customer")
	}

	data.Profile = customer

	for day := daysElapsed; day >= 0; day-- {
		meals, err := s.mealServ.Get(ctx, data.Profile, day)

		if err != nil {
			return eris.Wrapf(err, "Error fetching meals in day: %d", day)
		}

		entry := &DiaryEntry{
			Date:  now.AddDate(0, 0, -day),
			Meals: meals,
		}

		data.Diet = append(data.Diet, entry)
	}

	tmpl, err := template.New(CoachTmpl).Parse(prompts.NutritionCoachPromptTmpl)

	if err != nil {
		return eris.Wrap(err, "Error loading the prompt.tmpl file")
	}

	var tmplBuff bytes.Buffer

	err = tmpl.Execute(&tmplBuff, data)

	if err != nil {
		return eris.Wrap(err, "Error parsing prompt.tmpl file")
	}

	prompt := tmplBuff.String()

	aiAssessment, err := s.aiServ.Ask(prompt)

	if err != nil {
		return eris.Wrap(err, "Assesment error calling AI")
	}

	fmt.Print(aiAssessment)

	return nil
}
