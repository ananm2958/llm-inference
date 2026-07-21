package routing

import (
    "context"
    "fmt"
    "sync"

    "github.com/ananm2958/llm-gateway/internal/providers"
    "github.com/ananm2958/llm-gateway/internal/telemetry"
    "github.com/sony/gobreaker"
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/attribute"
    "go.opentelemetry.io/otel/codes"
    "go.opentelemetry.io/otel/trace"
)

type FallbackExecutor struct { registry *providers.Registry; breakers map[string]*gobreaker.CircuitBreaker; mu sync.Mutex }
func NewFallbackExecutor(registry *providers.Registry) *FallbackExecutor { return &FallbackExecutor{registry: registry, breakers: make(map[string]*gobreaker.CircuitBreaker)} }
func (f *FallbackExecutor) getBreaker(name string) *gobreaker.CircuitBreaker {
    f.mu.Lock(); defer f.mu.Unlock()
    if b := f.breakers[name]; b != nil { return b }
    b := gobreaker.NewCircuitBreaker(gobreaker.Settings{Name: name, OnStateChange: func(_ string, _ gobreaker.State, to gobreaker.State) { telemetry.RecordCircuitBreakerState(context.Background(), name, to) }})
    f.breakers[name] = b; return b
}
func (f *FallbackExecutor) Execute(ctx context.Context, req *providers.ChatRequest, chain []string) (*providers.ChatResponse, string, error) {
    tracer := otel.Tracer("llm.gateway.routing"); var lastErr error
    for _, name := range chain {
        provider, err := f.registry.GetByName(name); if err != nil { lastErr = err; continue }
        provCtx, span := tracer.Start(ctx, "provider.call", trace.WithAttributes(attribute.String("provider", name)))
        result, err := f.getBreaker(name).Execute(func() (interface{}, error) { return provider.Chat(provCtx, req) })
        if err != nil { span.SetStatus(codes.Error, err.Error()); span.End(); lastErr = err; continue }
        span.SetStatus(codes.Ok, ""); span.End(); return result.(*providers.ChatResponse), name, nil
    }
    if lastErr == nil { lastErr = fmt.Errorf("provider chain is empty") }; return nil, "", fmt.Errorf("all providers failed: %w", lastErr)
}
