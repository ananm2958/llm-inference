import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

func (f *FallbackExecutor) Execute(ctx context.Context, req *providers.ChatRequest, chain []string) (*providers.ChatResponse, string, error) {
	tracer := otel.Tracer("llm.gateway.routing")
	var lastErr error

	for _, providerName := range chain {
		provider, err := f.registry.GetByName(providerName)
		if err != nil {
			continue
		}

		provCtx, span := tracer.Start(ctx, "provider.call",
			trace.WithAttributes(attribute.String("provider", providerName)),
		)

		breaker := f.getBreaker(providerName)
		result, err := breaker.Execute(func() (interface{}, error) {
			return provider.Chat(provCtx, req)
		})

		if err != nil {
			span.SetStatus(codes.Error, err.Error())
			span.End()
			lastErr = err
			continue
		}

		span.SetStatus(codes.Ok, "")
		span.End()
		return result.(*providers.ChatResponse), providerName, nil
	}

	return nil, "", fmt.Errorf("all providers failed: %w", lastErr)
}