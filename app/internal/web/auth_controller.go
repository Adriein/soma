package web

import (
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
		err := c.service.ConnectNutritionApp(ctx, 1, "dd")

		if err != nil {
			ginErr := gin.Error{
				Err: err,
			}

			ctx.Errors = append(ctx.Errors, &ginErr)
		}
	}
}
