package routing

import (
    "context"
    "fmt"
    "sync/atomic"

    "github.com/ananm2958/llm-gateway/internal/providers"
)

type Router struct {
    registry *providers.Registry
    executor *FallbackExecutor
    policy *RoutingPolicy
    rrCounter atomic.Uint64
}

func NewRouter(registry *providers.Registry, policy *RoutingPolicy) *Router {
    return &Router{registry: registry, executor: NewFallbackExecutor(registry), policy: policy}
}

func (r *Router) Route(ctx context.Context, req *providers.ChatRequest) (*providers.ChatResponse, string, error) {
    chain, err := r.buildChain(req)
    if err != nil { return nil, "", err }
    return r.executor.Execute(ctx, req, chain)
}

func (r *Router) buildChain(req *providers.ChatRequest) ([]string, error) {
    if r.policy == nil || len(r.policy.ProviderChain) == 0 { return nil, fmt.Errorf("routing policy has no providers") }
    if override, ok := r.policy.ModelOverrides[req.Model]; ok {
        req.Model = override.ModelName
        chain := []string{override.ProviderName}
        for _, p := range r.policy.ProviderChain { if p != override.ProviderName { chain = append(chain, p) } }
        return chain, nil
    }
    if r.policy.Strategy == StrategyRoundRobin { return r.roundRobinChain(), nil }
    return append([]string(nil), r.policy.ProviderChain...), nil
}

func (r *Router) roundRobinChain() []string {
    chain := append([]string(nil), r.policy.ProviderChain...)
    offset := int(r.rrCounter.Add(1)-1) % len(chain)
    return append(chain[offset:], chain[:offset]...)
}
