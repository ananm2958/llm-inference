package middleware

import (
    "log/slog"
    "time"

    "github.com/gin-gonic/gin"
)

func StructuredLogger() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        c.Next()

        slog.Info("request",
            "method", c.Request.Method,
            "path", c.Request.URL.Path,
            "status", c.Writer.Status(),
            "latency_ms", time.Since(start).Milliseconds(),
            "model", c.GetString("model"),     // set by handlers
            "provider", c.GetString("provider"),
        )
    }
}