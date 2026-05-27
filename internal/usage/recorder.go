package usage

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Event struct {
	TenantID         string
	RequestID        string
	Model            string
	Provider         string
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
	CacheType        string
	LatencyMs        int
	Status           string
	ErrorMessage     string
	CostUSD          float64
}

type Recorder struct {
	pool *pgxpool.Pool
}

func NewRecorder(pool *pgxpool.Pool) *Recorder {
	return &Recorder{pool: pool}
}

func (r *Recorder) Record(event Event) {
	go func() {
		ctx := context.Background()

		var cacheTypePtr *string
		if event.CacheType != "" {
			cacheTypePtr = &event.CacheType
		}

		var errMsgPtr *string
		if event.ErrorMessage != "" {
			errMsgPtr = &event.ErrorMessage
		}

		_, err := r.pool.Exec(ctx, `
            INSERT INTO usage_events (
                tenant_id, request_id, model, provider,
                prompt_tokens, completion_tokens, total_tokens,
                cache_type, latency_ms, status, error_message, cost_usd
            ) VALUES (
                $1, $2, $3, $4,
                $5, $6, $7,
                $8, $9, $10, $11, $12
            )
        `,
			event.TenantID, event.RequestID, event.Model, event.Provider,
			event.PromptTokens, event.CompletionTokens, event.TotalTokens,
			cacheTypePtr, event.LatencyMs, event.Status, errMsgPtr, event.CostUSD,
		)

		if err != nil {
			slog.Error("usage event write failed", "error", err, "request_id", event.RequestID)
		}
	}()
}
