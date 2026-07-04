package internal

import (
	"database/sql"
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/adriein/soma/app/database"
	"github.com/adriein/soma/app/internal/customer"
	"github.com/adriein/soma/app/internal/worker"
	"github.com/adriein/soma/app/pkg/constants"
	"github.com/adriein/soma/app/pkg/helper"
	"github.com/adriein/soma/app/pkg/vendor"
	"github.com/joho/godotenv"
)

type Modules struct {
	Customer customer.CustomerService
	Worker   *worker.Worker
}

type App struct {
	Database *sql.DB
	Modules  *Modules
	Logger   *slog.Logger
}

func NewApp() *App {
	if os.Getenv(constants.Env) != constants.Prod {
		dotenvErr := godotenv.Load()

		if dotenvErr != nil {
			log.Fatal("Error loading .env file")
		}
	}

	checker := helper.NewEnvVarChecker(
		constants.DatabaseUser,
		constants.DatabasePassword,
		constants.DatabaseName,
		constants.ServerPort,
		constants.Env,
		constants.GeminiApiKey,
		constants.TelegramBotApiToken,
		constants.FatSecretClientId,
		constants.FatSecretApiKeyOauth1,
		constants.FatSecretApiKeyOauth2,
	)

	if envCheckerErr := checker.Check(); envCheckerErr != nil {
		log.Fatal(envCheckerErr.Error())
	}

	logger := initLogger()

	db := database.New()
	modules := initModules(db, logger)

	return &App{
		Database: db,
		Modules:  modules,
		Logger:   logger,
	}
}

func initLogger() *slog.Logger {
	opts := &slog.HandlerOptions{
		ReplaceAttr: func(groups []string, attr slog.Attr) slog.Attr {
			if attr.Key == slog.TimeKey {
				formatted := attr.Value.Time().UTC().Format(time.DateTime)

				return slog.String(slog.TimeKey, formatted)
			}

			return attr
		},
	}

	if os.Getenv(constants.Env) == constants.Dev {
		return slog.New(slog.NewTextHandler(os.Stdout, opts))
	}

	return slog.New(slog.NewJSONHandler(os.Stdout, opts))

}

func initModules(db *sql.DB, logger *slog.Logger) *Modules {
	telegram := vendor.NewTelegramBot()
	fsApi := vendor.NewFatSecret()

	customerRepo := customer.NewPgCustomerRepository(db)
	customerServ := customer.NewService(fsApi, telegram, customerRepo)

	worker := worker.New(customerServ, logger, telegram)

	return &Modules{
		Customer: customerServ,
		Worker:   worker,
	}
}
