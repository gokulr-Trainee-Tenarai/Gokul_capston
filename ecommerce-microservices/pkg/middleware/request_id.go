package middleware

import (
	"crypto/rand"
	"encoding/hex"

	"github.com/gin-gonic/gin"
)

const HeaderXRequestID = "X-Request-ID"

func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		reqID := c.GetHeader(HeaderXRequestID)
		if reqID == "" {
			bytes := make([]byte, 8)
			_, _ = rand.Read(bytes)
			reqID = hex.EncodeToString(bytes)
		}
		c.Set("request_id", reqID)
		c.Writer.Header().Set(HeaderXRequestID, reqID)
		c.Next()
	}
}
