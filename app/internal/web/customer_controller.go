package web

import (
	"net/http"

	"github.com/adriein/soma/app/internal"
	"github.com/adriein/soma/app/internal/customer"
	"github.com/gin-gonic/gin"
)

type CustomerController struct{
	service customer.CustomerService
}

func NewCustomerController(app *internal.App) *CustomerController {
	return &CustomerController{
		service: app.Modules.Customer,
	}
}

func (c *CustomerController) Get() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		c.service.ConnectNutritionApp()
		
		ctx.JSON(http.StatusOK, gin.H{"ok": true})
	}
}
