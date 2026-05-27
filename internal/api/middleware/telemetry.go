package middleware

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/ananm2958/llm-gateway/internal/telemetry"
	"github.com/ananm2958/llm-gateway/internal/usage"
)

func Telemetry(recorder *usage.Recorder, costCalc *telemetry.CostCalculator) gin.HandlerFunc {
	tracer := otel.Tracer("llm.gateway.http")

	return func(c *gin.Context) {
		start := time.Now()

		ctx, span := tracer.Start(c.Request.Context(), fmt.Sprintf("%s %s", c.Request.Method, c.FullPath()),
			trace.WithAttributes(
				attribute.String("http.method", c.Request.Method),
				attribute.String("http.path", c.FullPath()),
			),
		)
		defer span.End()

		c.Request = c.Request.WithContext(ctx)

		span.SpanContext().TraceID()
		c.Header("X-Trace-ID", span.SpanContext().TraceID().String())

		c.Next()

		latencyMs := time.Since(start).Milliseconds()

		tenantID := c.GetString("tenant_id")
		if tenantID == "" {
			tenantID = "default"
		}
		model := c.GetString("model")
		provider := c.GetString("provider")
		cacheType := c.GetString("cache_type")

		promptTokens := c.GetInt("prompt_tokens")
		completionTokens := c.GetInt("completion_tokens")
		status := "success"
		if c.Writer.Status() >= 400 {
			status = "error"
			span.SetStatus(codes.Error, fmt.Sprintf("HTTP %d", c.Writer.Status()))
		}

		span.SetAttributes(
			attribute.String("tenant_id", tenantID),
			attribute.String("model", model),
			attribute.String("provider", provider),
			attribute.String("cache_type", cacheType),
			attribute.Int("prompt_tokens", promptTokens),
			attribute.Int("completion_tokens", completionTokens),
			attribute.Int64("latency_ms", latencyMs),
		)

		costUSD, _ := costCalc.Calculate(ctx, provider, model, promptTokens, completionTokens)

		telemetry.RecordRequest(ctx,
			tenantID, model, provider, status, cacheType,
			latencyMs, int64(promptTokens), int64(completionTokens), costUSD,
		)

		recorder.Record(usage.Event{
			TenantID:         tenantID,
			RequestID:        span.SpanContext().TraceID().String(),
			Model:            model,
			Provider:         provider,
			PromptTokens:     promptTokens,
			CompletionTokens: completionTokens,
			TotalTokens:      promptTokens + completionTokens,
			CacheType:        cacheType,
			LatencyMs:        int(latencyMs),
			Status:           status,
			CostUSD:          costUSD,
		})
	}
}
