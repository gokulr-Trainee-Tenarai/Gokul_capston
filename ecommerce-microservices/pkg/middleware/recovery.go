package middleware

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

func RecoveryMiddleware(serviceName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				reqID, _ := c.Get("request_id")
				slog.Error("Unhandled panic intercepted",
					slog.String("service", serviceName),
					slog.Any("request_id", reqID),
					slog.Any("panic", r),
				)
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"error": "internal server error: panic recovered",
				})
			}
		}()
		c.Next()
	}
}
