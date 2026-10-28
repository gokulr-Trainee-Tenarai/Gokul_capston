package middleware

import (
	"log/slog"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

var jsonLogger = slog.New(slog.NewJSONHandler(os.Stdout, nil))

func StructuredLogger(serviceName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		reqID, _ := c.Get("request_id")
		reqIDStr, ok := reqID.(string)
		if !ok || reqIDStr == "" {
			reqIDStr = c.Writer.Header().Get(HeaderXRequestID)
		}

		jsonLogger.Info("HTTP Request Completed",
			slog.String("service", serviceName),
			slog.String("request_id", reqIDStr),
			slog.String("method", c.Request.Method),
			slog.String("path", c.Request.URL.Path),
			slog.Int("status", c.Writer.Status()),
			slog.Duration("duration", time.Since(start)),
			slog.String("client_ip", c.ClientIP()),
		)
	}
}
