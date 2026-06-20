package web

import (
	"github.com/adriein/soma/app/internal"
	"github.com/gin-gonic/gin"
)

type HealthController struct{}

func NewHealthController(app *internal.App) *HealthController {
	return &HealthController{}
}

func (c *HealthController) Get() gin.HandlerFunc {
	return func(ctx *gin.Context) {
	}
}
