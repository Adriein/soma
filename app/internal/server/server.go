package server

import (
	"log"
	"log/slog"

	"github.com/adriein/soma/app/internal"
	"github.com/adriein/soma/app/internal/web"
	"github.com/adriein/soma/app/pkg/middleware"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/rotisserie/eris"
)

type TibiaChar struct {
	app       *internal.App
	gin       *gin.Engine
	validator *validator.Validate
}

func New(port string) *TibiaChar {
	app := internal.NewApp()

	engine := gin.New()

	// Disable trusted proxy warning.
	engine.SetTrustedProxies(nil)

	engine.Use(middleware.Error(), gin.Logger(), gin.Recovery(), middleware.Tracer())

	tibiaChar := &TibiaChar{
		app:       app,
		gin:       engine,
		validator: validator.New(),
	}

	tibiaChar.routeSetup()

	if ginErr := engine.Run(port); ginErr != nil {
		err := eris.Wrap(ginErr, "Error starting HTTP server")

		log.Fatal(eris.ToString(err, true))
	}

	slog.Info("Starting the TibiaChar at " + port)

	return tibiaChar
}

func (t *TibiaChar) routeSetup() {
	//HEALTH CHECK
	t.gin.GET("/health", web.NewHealthController(t.app).Get())
}
