package routing

import (
	"context"
	"sync/atomic"

	"github.com/ananm2958/llm-gateway/internal/providers"
	"honnef.co/go/tools/config"
)

type Router struct {
	registry  *providers.Registry
	executor  *FallbackExecutor
	policy    *RoutingPolicy
	rrCounter atomic.Uint64
}

func NewRouter(registry *providers.Registry, policy *RoutingPolicy) *Router {
	return &Router{
		registry: registry,
		executor: NewFallbackExecutor(registry),
		policy:   policy,
	}
}

func (r *Router) Route(ctx context.Context, req *providers.ChatRequest) (*providers.ChatResponse, error) {
	chain := r.buildChain(req)
	return r.executor.Execute(ctx, req, chain)
}

func (r *Router) buildChain(req *providers.ChatRequest) []string {
	if override, ok := r.policy.ModelOverrides[req.Model]; ok {
		req.Model = override.ModelName
		chain := []string{override.ProviderName}
		for _, p := range r.policy.ProviderChain {
			if p != override.ProviderName {
				chain = append(chain, p)
			}
		}
		return chain
	}

	if r.policy.Strategy == StrategyRoundRobin {
		return r.roundRobinChain()
	}

	return r.policy.ProviderChain
}

func (r *Router) roundRobinChain() []string {
	n := r.rrCounter.Add(1)
	chain := make([]string, len(r.policy.ProviderChain))
	copy(chain, r.policy.ProviderChain)
	offset := int(n) % len(chain)
	return append(chain[offset:], chain[:offset]...)
}

func NewRouter(cfg *config.Config, router *routing.Router,
	providerList []providers.Provider,
	recorder *usage.Recorder,
	costCalc *telemetry.CostCalculator,
) *gin.Engine {

	r := gin.New()
	r.Use(middleware.StructuredLogger())
	r.Use(middleware.Telemetry(recorder, costCalc))
	r.Use(gin.Recovery())

	r.GET("/metrics", gin.WrapH(promhttp.Handler()))
	r.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })

	v1 := r.Group("/v1", middleware.APIKeyAuth(cfg.Server.APIKeys))
	{
		chatHandler := handlers.NewChatHandler(router, cacheManager)
		v1.POST("/chat/completions", chatHandler.Handle)
		v1.GET("/models", handlers.ModelsHandler(providerList))
	}

	return r
}
