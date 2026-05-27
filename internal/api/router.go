package api

import (
    "github.com/gin-gonic/gin"
    "github.com/ananm2958/llm-gateway/internal/api/handlers"
    "github.com/ananm2958/llm-gateway/internal/api/middleware"
    "github.com/ananm2958/llm-gateway/internal/config"
    "github.com/ananm2958/llm-gateway/internal/providers"
)

func NewRouter(cfg *config.Config, providerList []providers.Provider) *gin.Engine {
    r := gin.New()
    r.Use(middleware.StructuredLogger())
    r.Use(gin.Recovery())


    r.GET("/health", func(c *gin.Context) {
        c.JSON(200, gin.H{"status": "ok"})
    })

  
    v1 := r.Group("/v1", middleware.APIKeyAuth(cfg.Server.APIKeys))
    {
        chatHandler := handlers.NewChatHandler(providerList[0]) 
        v1.POST("/chat/completions", chatHandler.Handle)
        v1.GET("/models", handlers.ModelsHandler(providerList))
    }

    return r
}

func (r *Router) Route(ctx context.Context, req *providers.ChatRequest) (*providers.ChatResponse, string, error) {
    chain := r.buildChain(req)
    return r.executor.Execute(ctx, req, chain)
}


return result.(*providers.ChatResponse), providerName, nil