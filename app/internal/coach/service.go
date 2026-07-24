package coach

import (
	"bytes"
	"context"
	"errors"
	"fmt"
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
	CtxSessionID                 = "SessionID"
)

type CoachService interface {
	Assessment(ctx context.Context, chatID int64) error
}

type Service struct {
	customerServ   customer.CustomerService
	mealServ       meal.MealService
	aiServ         vendor.AI
	assessmentRepo AssessmentRepository
	bot            vendor.Bot
	logger         *slog.Logger
}

func NewService(
	customerServ customer.CustomerService,
	mealServ meal.MealService,
	aiServ vendor.AI,
	assessmentRepo AssessmentRepository,
	bot vendor.Bot,
	logger *slog.Logger,
) *Service {
	return &Service{
		customerServ:   customerServ,
		mealServ:       mealServ,
		aiServ:         aiServ,
		assessmentRepo: assessmentRepo,
		bot:            bot,
		logger:         logger,
	}
}

func (s *Service) Assessment(ctx context.Context, chatID int64) error {
	//TODO: añadir el analisis por semana en lugar de un todo
	var data AssessmentData

	sessionID := helper.TinyUuid()

	ctx = context.WithValue(ctx, CtxSessionID, sessionID)

	customer, err := s.customerServ.GetCustomer(ctx, chatID)

	if err != nil {
		return eris.Wrap(err, "Error fetching customer")
	}

	data.Profile = customer

	meals, err := s.collectMeals(ctx, &data)

	if err != nil {
		return err
	}

	data.Diet = meals

	//assessment, err := s.assessmentRepo.GetByID(ctx, 3)

	assessment, err := s.aiAssessment(ctx, &data)

	if err != nil {
		return err
	}

	s.logger.Debug(assessment.Content)

	if err := s.assessmentRepo.Save(ctx, assessment); err != nil {
		return eris.Wrap(err, "Error saving the ai generated assessment")
	}

	if len(assessment.Content) >= vendor.TelegramMaxMessageCharLength {
		messageChunks := helper.SplitMessage(assessment.Content, AssessmentChunkMaxCharLength)

		for _, chunk := range messageChunks {
			message := vendor.OutgoingMessage{
				ChatID:    data.Profile.TelegramChatID,
				Text:      fmt.Sprintf("*SomaBot \\#%s*\n\n%s", sessionID, helper.CustomXMLToMarkdownV2(chunk)),
				ParseMode: vendor.TelegramMarkdownV2,
			}

			if err := helper.ValidateMarkdownV2(message.Text); err != nil {
				return eris.Wrap(err, "Error in markdown validation")
			}

			if err := s.bot.SendMessage(ctx, message); err != nil {
				return err
			}
		}

		return nil
	}

	message := vendor.OutgoingMessage{
		ChatID:    data.Profile.TelegramChatID,
		Text:      fmt.Sprintf("*SomaBot \\#%s*\n\n%s", sessionID, helper.CustomXMLToMarkdownV2(assessment.Content)),
		ParseMode: vendor.TelegramMarkdownV2,
	}

	if err := helper.ValidateMarkdownV2(message.Text); err != nil {
		return eris.Wrap(err, "Error in markdown validation")
	}

	if err := s.bot.SendMessage(ctx, message); err != nil {
		return eris.Wrap(err, "Error sending the coach assessment")
	}

	return nil
}

func (s *Service) collectMeals(ctx context.Context, data *AssessmentData) ([]*DiaryEntry, error) {
	feedback := vendor.OutgoingMessage{
		ChatID:    data.Profile.TelegramChatID,
		Text:      fmt.Sprintf("*SomaBot \\#%s*\n🍉 Recopilando datos de todas las comidas\\.\\.\\.", ctx.Value(CtxSessionID)),
		ParseMode: vendor.TelegramMarkdownV2,
	}

	if err := s.bot.SendMessage(ctx, feedback); err != nil {
		return nil, eris.Wrap(err, "Error sending feedback message")
	}

	loc, err := time.LoadLocation(constants.TimeLocationMadrid)

	if err != nil {
		return nil, eris.Wrap(err, "Error loading location")
	}

	now := time.Now().In(loc)

	//The fist date I've added meals is 2026-06-25
	targetDate := time.Date(2026, time.June, 25, 0, 0, 0, 0, loc)

	dur := now.Sub(targetDate)

	daysElapsed := int(dur.Hours() / 24)

	var results []*DiaryEntry

	for day := daysElapsed; day >= 0; day-- {
		meals, err := s.mealServ.Get(ctx, data.Profile, day)

		if err != nil {
			return nil, eris.Wrapf(err, "Error fetching meals in day: %d", day)
		}

		entry := &DiaryEntry{
			Date:  now.AddDate(0, 0, -day),
			Meals: meals,
		}

		results = append(results, entry)
	}

	return results, nil
}

func (s *Service) aiAssessment(ctx context.Context, data *AssessmentData) (*Assessment, error) {
	tmpl, err := template.New(CoachTmpl).Parse(prompts.NutritionCoachPromptTmpl)

	if err != nil {
		return nil, eris.Wrap(err, "Error loading the prompt.tmpl file")
	}

	var tmplBuff bytes.Buffer

	err = tmpl.Execute(&tmplBuff, data)

	if err != nil {
		return nil, eris.Wrap(err, "Error parsing prompt.tmpl file")
	}

	prompt := tmplBuff.String()

	feedback := vendor.OutgoingMessage{
		ChatID:    data.Profile.TelegramChatID,
		Text:      fmt.Sprintf("*SomaBot \\#%s*\n🤖 Preguntando a los expertos de silicio\\.\\.\\.", ctx.Value(CtxSessionID)),
		ParseMode: vendor.TelegramMarkdownV2,
	}

	if err := s.bot.SendMessage(ctx, feedback); err != nil {
		return nil, eris.Wrap(err, "Error sending feedback message")
	}

	aiRes, err := s.aiServ.Ask(prompt)

	if err != nil {
		if errors.Is(err, vendor.ErrAISpikeDemand) {
			feedback = vendor.OutgoingMessage{
				ChatID:    data.Profile.TelegramChatID,
				Text:      fmt.Sprintf("*SomaBot \\#%s*\n🤡 IA saturada, porfavor intentelo luego\\.", ctx.Value(CtxSessionID)),
				ParseMode: vendor.TelegramMarkdownV2,
			}

			if err := s.bot.SendMessage(ctx, feedback); err != nil {
				return nil, eris.Wrap(err, "Error sending feedback message")
			}
		}

		return nil, eris.Wrap(err, "Assesment error calling AI")
	}

	escapedRes := helper.EscapeText(aiRes.Text())

	return &Assessment{Content: escapedRes}, nil
}
