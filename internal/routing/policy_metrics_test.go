package routing

import (
	"context"
	"errors"
	"io"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/ananm2958/llm-gateway/internal/providers"
)

// measuringProvider records every attempted LLM call for routing assertions.
type measuringProvider struct {
	name  string
	delay time.Duration
	fail  bool
	mu    sync.Mutex
	calls int
}

func (p *measuringProvider) Name() string              { return p.name }
func (p *measuringProvider) SupportedModels() []string { return []string{"model"} }
func (p *measuringProvider) Chat(ctx context.Context, _ *providers.ChatRequest) (*providers.ChatResponse, error) {
	p.mu.Lock()
	p.calls++
	p.mu.Unlock()
	if p.delay > 0 {
		select {
		case <-time.After(p.delay):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	if p.fail {
		return nil, errors.New("provider unavailable")
	}
	return &providers.ChatResponse{Model: "model"}, nil
}
func (p *measuringProvider) ChatStream(context.Context, *providers.ChatRequest) (io.ReadCloser, error) {
	return nil, nil
}
func (p *measuringProvider) Healthy(context.Context) bool { return true }
func (p *measuringProvider) CallCount() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.calls
}

func TestPriorityPolicyFallbackMaintainsSuccessRateAndCountsLLMCalls(t *testing.T) {
	primary := &measuringProvider{name: "primary", fail: true}
	fallback := &measuringProvider{name: "fallback"}
	router := NewRouter(providers.NewRegistry([]providers.Provider{primary, fallback}), &RoutingPolicy{ProviderChain: []string{"primary", "fallback"}, Strategy: StrategyPriority})
	const requests = 3 // Keep below the circuit breaker's opening threshold.
	successes := 0
	for range requests {
		response, provider, err := router.Route(context.Background(), &providers.ChatRequest{Model: "model"})
		if err != nil || response == nil || provider != "fallback" {
			t.Fatalf("Route() = (%v, %q, %v), want successful fallback", response, provider, err)
		}
		successes++
	}
	if got := float64(successes) / requests; got != 1 {
		t.Errorf("request success rate = %.2f, want 1.00", got)
	}
	if got, want := primary.CallCount()+fallback.CallCount(), 2*requests; got != want {
		t.Errorf("total LLM calls = %d, want %d", got, want)
	}
}

func TestPriorityPolicyFallbackIncludesFailedAttemptInTailLatency(t *testing.T) {
	const requests = 5 // Keep below the circuit breaker's opening threshold.
	const primaryDelay = 20 * time.Millisecond
	primary := &measuringProvider{name: "primary", delay: primaryDelay, fail: true}
	fallback := &measuringProvider{name: "fallback", delay: 2 * time.Millisecond}
	router := NewRouter(providers.NewRegistry([]providers.Provider{primary, fallback}), &RoutingPolicy{ProviderChain: []string{"primary", "fallback"}, Strategy: StrategyPriority})
	latencies := make([]time.Duration, 0, requests)
	for range requests {
		started := time.Now()
		response, provider, err := router.Route(context.Background(), &providers.ChatRequest{Model: "model"})
		latencies = append(latencies, time.Since(started))
		if err != nil || response == nil || provider != "fallback" {
			t.Fatalf("Route() = (%v, %q, %v), want successful fallback", response, provider, err)
		}
	}
	sort.Slice(latencies, func(i, j int) bool { return latencies[i] < latencies[j] })
	p95 := latencies[(95*len(latencies)+99)/100-1]
	if p95 < primaryDelay {
		t.Errorf("p95 request latency = %s, want at least %s", p95, primaryDelay)
	}
	if got, want := primary.CallCount()+fallback.CallCount(), 2*requests; got != want {
		t.Errorf("total LLM calls = %d, want %d", got, want)
	}
}
