package middleware

import (
	"github.com/adriein/soma/app/pkg/helper"
	"github.com/gin-gonic/gin"
)

const TraceIDKey = "TraceID"

func Tracer() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		traceID := helper.TraceID()

		ctx.Set(TraceIDKey, traceID)

		ctx.Writer.Header().Set("X-Trace-ID", traceID)

		ctx.Next()
	}
}
