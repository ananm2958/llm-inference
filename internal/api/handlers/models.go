package handlers

import (
	"net/http"
	"time"

	"github.com/ananm2958/llm-gateway/internal/providers"
	"github.com/gin-gonic/gin"
)

func ModelsHandler(providerList []providers.Provider) gin.HandlerFunc {
	return func(c *gin.Context) {
		type modelEntry struct {
			ID      string `json:"id"`
			Object  string `json:"object"`
			Created int64  `json:"created"`
			OwnedBy string `json:"owned_by"`
		}

		var models []modelEntry
		for _, p := range providerList {
			for _, m := range p.SupportedModels() {
				models = append(models, modelEntry{
					ID:      m,
					Object:  "model",
					Created: time.Now().Unix(),
					OwnedBy: p.Name(),
				})
			}
		}

		c.JSON(http.StatusOK, gin.H{"object": "list", "data": models})
	}
}
