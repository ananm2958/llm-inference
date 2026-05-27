package telemetry

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

const meterName = "llm.gateway"

var (
	meter = otel.Meter(meterName)

	RequestCounter, _ = meter.Int64Counter(
		"llm.gateway.requests.total",
		metric.WithDescription("Total number of LLM requests"),
	)

	LatencyHistogram, _ = meter.Int64Histogram(
		"llm.gateway.request.latency_ms",
		metric.WithDescription("End-to-end request latency in milliseconds"),
		metric.WithUnit("ms"),
	)

	PromptTokenCounter, _ = meter.Int64Counter(
		"llm.gateway.tokens.prompt",
		metric.WithDescription("Total prompt tokens consumed"),
	)

	CompletionTokenCounter, _ = meter.Int64Counter(
		"llm.gateway.tokens.completion",
		metric.WithDescription("Total completion tokens produced"),
	)

	CacheHitCounter, _ = meter.Int64Counter(
		"llm.gateway.cache.hits",
		metric.WithDescription("Cache hits by cache type"),
	)

	CacheMissCounter, _ = meter.Int64Counter(
		"llm.gateway.cache.misses",
		metric.WithDescription("Cache misses"),
	)

	CostMicrodollarsCounter, _ = meter.Int64Counter(
		"llm.gateway.cost.microdollars",
		metric.WithDescription("Estimated cost in microdollars (divide by 1,000,000 for USD)"),
	)

	ProviderErrorCounter, _ = meter.Int64Counter(
		"llm.gateway.provider.errors",
		metric.WithDescription("Provider errors by provider and error type"),
	)

	CircuitBreakerOpen, _ = meter.Int64Gauge(
		"llm.gateway.circuit_breaker.open",
		metric.WithDescription("1 if circuit breaker is open for a provider, 0 otherwise"),
	)
)

func AttrTenant(id string) attribute.KeyValue {
	return attribute.String("tenant_id", id)
}
func AttrModel(m string) attribute.KeyValue {
	return attribute.String("model", m)
}
func AttrProvider(p string) attribute.KeyValue {
	return attribute.String("provider", p)
}
func AttrStatus(s string) attribute.KeyValue {
	return attribute.String("status", s)
}
func AttrCacheType(t string) attribute.KeyValue {
	return attribute.String("cache_type", t)
}

func RecordRequest(ctx context.Context, tenantID, model, provider, status, cacheType string,
	latencyMs, promptTokens, completionTokens int64, costUSD float64) {

	attrs := metric.WithAttributes(
		AttrTenant(tenantID),
		AttrModel(model),
		AttrProvider(provider),
		AttrStatus(status),
	)

	RequestCounter.Add(ctx, 1, attrs)
	LatencyHistogram.Record(ctx, latencyMs, attrs)
	PromptTokenCounter.Add(ctx, promptTokens, attrs)
	CompletionTokenCounter.Add(ctx, completionTokens, attrs)
	CostMicrodollarsCounter.Add(ctx, int64(costUSD*1_000_000), attrs)

	if cacheType != "" {
		CacheHitCounter.Add(ctx, 1, metric.WithAttributes(
			AttrTenant(tenantID), AttrModel(model), AttrCacheType(cacheType),
		))
	} else {
		CacheMissCounter.Add(ctx, 1, metric.WithAttributes(
			AttrTenant(tenantID), AttrModel(model),
		))
	}
}
