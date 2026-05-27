package semantic

import (
    "context"
    "encoding/json"
    "log/slog"

    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/pgvector/pgvector-go"
    "github.com/yourorg/llm-gateway/internal/providers"
)

// SimilarityThreshold controls how close two prompts must be to count as a cache hit.
// 1.0 = identical, 0.0 = completely different.
// 0.92 is a reasonable starting point — tune based on your use case.
const SimilarityThreshold = 0.92

type Cache struct {
    pool *pgxpool.Pool
}

func New(ctx context.Context, connStr string) (*Cache, error) {
    pool, err := pgxpool.New(ctx, connStr)
    if err != nil {
        return nil, err
    }
    return &Cache{pool: pool}, nil
}

// Get searches for a semantically similar cached response.
// Returns nil if no match above the similarity threshold is found.
func (c *Cache) Get(
    ctx context.Context,
    tenantID, model string,
    embedding []float32,
) (*providers.ChatResponse, float64, error) {

    vec := pgvector.NewVector(embedding)

    // cosine similarity: 1 - cosine_distance
    // We filter by tenant and model first to keep the vector scan small
    row := c.pool.QueryRow(ctx, `
        SELECT response_json,
               1 - (embedding <=> $1) AS similarity
        FROM   semantic_cache
        WHERE  tenant_id = $2
          AND  model     = $3
          AND  1 - (embedding <=> $1) >= $4
        ORDER  BY embedding <=> $1
        LIMIT  1
    `, vec, tenantID, model, SimilarityThreshold)

    var responseJSON []byte
    var similarity float64

    if err := row.Scan(&responseJSON, &similarity); err != nil {
        // pgx returns pgx.ErrNoRows on miss — treat as cache miss
        return nil, 0, nil
    }

    var resp providers.ChatResponse
    if err := json.Unmarshal(responseJSON, &resp); err != nil {
        return nil, 0, err
    }

    return &resp, similarity, nil
}

// Set stores a new entry in the semantic cache.
func (c *Cache) Set(
    ctx context.Context,
    tenantID, model, promptHash, promptText string,
    embedding []float32,
    resp *providers.ChatResponse,
) error {
    responseJSON, err := json.Marshal(resp)
    if err != nil {
        return err
    }

    vec := pgvector.NewVector(embedding)

    _, err = c.pool.Exec(ctx, `
        INSERT INTO semantic_cache
            (tenant_id, model, prompt_hash, prompt_text, embedding, response_json)
        VALUES ($1, $2, $3, $4, $5, $6)
        ON CONFLICT DO NOTHING
    `, tenantID, model, promptHash, promptText, vec, responseJSON)

    if err != nil {
        slog.Error("semantic cache write failed", "error", err)
    }
    return err
}