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

func (c *AuthController) Auth() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		_, err := c.service.ConnectNutritionApp()

		if err != nil {
			ginErr := gin.Error{
				Err: err,
			}

			ctx.Errors = append(ctx.Errors, &ginErr)
		}

		ctx.JSON(http.StatusOK, gin.H{"ok": true})
	}
}

func (c *AuthController) AuthWebhook() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		c.service.VerifyToken("")

		ctx.JSON(http.StatusOK, gin.H{"ok": true})
	}
}
