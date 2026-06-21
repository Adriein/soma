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

type Soma struct {
	app       *internal.App
	gin       *gin.Engine
	validator *validator.Validate
}

func New(port string) *Soma {
	app := internal.NewApp()

	engine := gin.New()

	// Disable trusted proxy warning.
	engine.SetTrustedProxies(nil)

	engine.Use(middleware.Error(), gin.Logger(), gin.Recovery(), middleware.Tracer())

	soma := &Soma{
		app:       app,
		gin:       engine,
		validator: validator.New(),
	}

	soma.routeSetup()

	if ginErr := engine.Run(port); ginErr != nil {
		err := eris.Wrap(ginErr, "Error starting HTTP server")

		log.Fatal(eris.ToString(err, true))
	}

	slog.Info("Starting the Soma at " + port)

	return soma
}

func (t *Soma) routeSetup() {
	//HEALTH CHECK
	t.gin.GET("/health", web.NewHealthController(t.app).Get())

	//AUTH
	t.gin.GET("/auth", web.NewAuthController(t.app).Get())
}
