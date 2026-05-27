package handlers

import (
    "net/http"

    "github.com/gin-gonic/gin"
    "github.com/ananm2958/llm-gateway/internal/cache"
    "github.com/ananm2958/llm-gateway/internal/providers"
    "github.com/ananm2958/llm-gateway/internal/routing"
)

type ChatHandler struct {
    router       *routing.Router
    cacheManager *cache.Manager
}

func NewChatHandler(router *routing.Router, cm *cache.Manager) *ChatHandler {
    return &ChatHandler{router: router, cacheManager: cm}
}

func (h *ChatHandler) Handle(c *gin.Context) {
    var req providers.ChatRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    bypassCache := c.GetHeader("X-Cache-Bypass") == "true"

    tenantID := c.GetString("tenant_id") 
    if tenantID == "" {
        tenantID = "default"
    }

    c.Set("model", req.Model)

    if !bypassCache {
        result, err := h.cacheManager.Get(c.Request.Context(), tenantID, &req)
        if err == nil && result != nil {
            c.Header("X-Cache", result.CacheType)
            c.JSON(http.StatusOK, result.Response)
            return
        }
    }

    resp, err := h.router.Route(c.Request.Context(), &req)
    if err != nil {
        c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
        return
    }

    if !bypassCache {
        h.cacheManager.WriteBack(tenantID, &req, resp)
    }

    c.Header("X-Cache", "miss")
    c.JSON(http.StatusOK, resp)
}


// On a cache hit:
c.Set("cache_type", result.CacheType)
c.Set("provider", "cache")
c.Set("prompt_tokens", 0)
c.Set("completion_tokens", 0)

c.Set("provider", resp.Model)   
c.Set("prompt_tokens", resp.Usage.PromptTokens)
c.Set("completion_tokens", resp.Usage.CompletionTokens)