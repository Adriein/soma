package coach

import (
	"bytes"
	"context"
	"log/slog"
	"text/template"
	"time"

	"github.com/adriein/soma/app/internal/customer"
	"github.com/adriein/soma/app/internal/meal"
	"github.com/adriein/soma/app/pkg/constants"
	"github.com/adriein/soma/app/pkg/helper"
	"github.com/adriein/soma/app/pkg/prompts"
	"github.com/adriein/soma/app/pkg/vendor"
	"github.com/rotisserie/eris"
)

const (
	CoachTmpl                    = "coach"
	AssessmentChunkMaxCharLength = 2000
)

type CoachService interface {
	Assessment(ctx context.Context, chatID int64) error
}

type Service struct {
	customerServ customer.CustomerService
	mealServ     meal.MealService
	aiServ       vendor.AI
	bot          vendor.Bot
	logger       *slog.Logger
}

func NewService(
	customerServ customer.CustomerService,
	mealServ meal.MealService,
	aiServ vendor.AI,
	bot vendor.Bot,
	logger *slog.Logger,
) *Service {
	return &Service{
		customerServ: customerServ,
		mealServ:     mealServ,
		aiServ:       aiServ,
		bot:          bot,
		logger:       logger,
	}
}

func (s *Service) Assessment(ctx context.Context, chatID int64) error {
	//TODO: the main idea is to get everything the first time and store the evaluation in the db with the date then the next times i only take from the last evaluation to today
	//TODO: añadir el analisis por semana en lugar de un todo
	loc, err := time.LoadLocation(constants.TimeLocationMadrid)

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

	feedback := vendor.OutgoingMessage{
		ChatID: data.Profile.TelegramChatID,
		Text:   "🍉 Recopilando datos de todas las comidas...",
	}

	if err := s.bot.SendMessage(ctx, feedback); err != nil {
		return eris.Wrap(err, "Error sending feedback message")
	}

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

	feedback = vendor.OutgoingMessage{
		ChatID: data.Profile.TelegramChatID,
		Text:   "🤖 Preguntando a los expertos de silicio...",
	}

	if err := s.bot.SendMessage(ctx, feedback); err != nil {
		return eris.Wrap(err, "Error sending feedback message")
	}

	aiRes, err := s.aiServ.Ask(prompt)

	if err != nil {
		return eris.Wrap(err, "Assesment error calling AI")
	}

	text := helper.EscapeText(aiRes.Text())

	if len(text) >= vendor.TelegramMaxMessageCharLength {
		messageChunks := helper.SplitMessage(text, AssessmentChunkMaxCharLength)

		for _, chunk := range messageChunks {
			message := vendor.OutgoingMessage{
				ChatID:    data.Profile.TelegramChatID,
				Text:      chunk,
				ParseMode: vendor.TelegramMarkdownV2,
			}

			if err := s.bot.SendMessage(ctx, message); err != nil {
				return err
			}
		}

		return nil
	}

	message := vendor.OutgoingMessage{
		ChatID:    data.Profile.TelegramChatID,
		Text:      text,
		ParseMode: vendor.TelegramMarkdownV2,
	}

	if err := s.bot.SendMessage(ctx, message); err != nil {
		return eris.Wrap(err, "Error sending the coach assessment")
	}

	return nil
}
