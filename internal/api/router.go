package api
import (
    "net/http"
    "github.com/ananm2958/llm-gateway/internal/api/handlers"
    "github.com/ananm2958/llm-gateway/internal/api/middleware"
    "github.com/ananm2958/llm-gateway/internal/cache"
    "github.com/ananm2958/llm-gateway/internal/config"
    "github.com/ananm2958/llm-gateway/internal/providers"
    "github.com/ananm2958/llm-gateway/internal/routing"
    "github.com/ananm2958/llm-gateway/internal/telemetry"
    "github.com/ananm2958/llm-gateway/internal/usage"
    "github.com/gin-gonic/gin"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)
func NewRouter(cfg *config.Config, router *routing.Router, providerList []providers.Provider, cacheManager *cache.Manager, recorder *usage.Recorder, costCalc *telemetry.CostCalculator) *gin.Engine {
    r:=gin.New(); r.Use(middleware.StructuredLogger(), middleware.Telemetry(recorder,costCalc), gin.Recovery())
    r.GET("/health",func(c *gin.Context){c.JSON(http.StatusOK,gin.H{"status":"ok"})}); r.GET("/metrics",gin.WrapH(promhttp.Handler()))
    v1:=r.Group("/v1",middleware.APIKeyAuth(cfg.Server.APIKeys)); chat:=handlers.NewChatHandler(router,cacheManager)
    v1.POST("/chat/completions",chat.Handle); v1.POST("/completions",handlers.CompletionsHandler()); v1.GET("/models",handlers.ModelsHandler(providerList)); return r
}
