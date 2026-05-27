package middleware

import (
    "net/http"
    "strings"

    "github.com/gin-gonic/gin"
)

func APIKeyAuth(validKeys []string) gin.HandlerFunc {
    keySet := make(map[string]struct{}, len(validKeys))
    for _, k := range validKeys {
        keySet[k] = struct{}{}
    }

    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        if !strings.HasPrefix(authHeader, "Bearer ") {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing or invalid Authorization header"})
            return
        }

        key := strings.TrimPrefix(authHeader, "Bearer ")
        if _, ok := keySet[key]; !ok {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid API key"})
            return
        }

        c.Next()
    }
}