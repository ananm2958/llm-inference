package handlers
import ("net/http"; "github.com/gin-gonic/gin")
func CompletionsHandler() gin.HandlerFunc { return func(c *gin.Context) { c.JSON(http.StatusNotImplemented,gin.H{"error":gin.H{"message":"/v1/completions is not implemented; use /v1/chat/completions","type":"not_implemented"}}) } }
