package web

import (
	"net/http"

	"github.com/adriein/soma/app/internal"
	"github.com/adriein/soma/app/internal/customer"
	"github.com/gin-gonic/gin"
)

type AuthController struct {
	service customer.CustomerService
}

func NewAuthController(app *internal.App) *AuthController {
	return &AuthController{
		service: app.Modules.Customer,
	}
}

func (c *AuthController) Get() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		c.service.ConnectNutritionApp()

		ctx.JSON(http.StatusOK, gin.H{"ok": true})
	}
}
