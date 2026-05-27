package cache

import (
	"context"
	"crypto/sha256"
	"fmt"
	"log/slog"

	"github.com/ananm2958/llm-gateway/internal/cache/exact"
	"github.com/ananm2958/llm-gateway/internal/cache/keygen"
	"github.com/ananm2958/llm-gateway/internal/cache/semantic"
	"github.com/ananm2958/llm-gateway/internal/embedding"
	"github.com/ananm2958/llm-gateway/internal/providers"
)

type Manager struct {
	exact    *exact.Cache
	semantic *semantic.Cache
	embedder *embedding.Embedder
}

func NewManager(e *exact.Cache, s *semantic.Cache, emb *embedding.Embedder) *Manager {
	return &Manager{exact: e, semantic: s, embedder: emb}
}

type Result struct {
	Response   *providers.ChatResponse
	CacheType  string
	Similarity float64
}

func (m *Manager) Get(ctx context.Context, tenantID string, req *providers.ChatRequest) (*Result, error) {
	exactKey, err := keygen.ExactKey(tenantID, req)
	if err != nil {
		return nil, err
	}

	resp, err := m.exact.Get(ctx, exactKey)
	if err != nil {
		slog.Warn("exact cache error", "error", err)
	}
	if resp != nil {
		return &Result{Response: resp, CacheType: "exact"}, nil
	}

	promptText := keygen.PromptText(req)
	embedding, err := m.embedder.Embed(ctx, promptText)
	if err != nil {
		slog.Warn("embedder error, skipping semantic cache", "error", err)
		return nil, nil
	}

	resp, similarity, err := m.semantic.Get(ctx, tenantID, req.Model, embedding)
	if err != nil {
		slog.Warn("semantic cache error", "error", err)
	}
	if resp != nil {
		return &Result{Response: resp, CacheType: "semantic", Similarity: similarity}, nil
	}

	return nil, nil
}

func (m *Manager) WriteBack(tenantID string, req *providers.ChatRequest, resp *providers.ChatResponse) {
	go func() {
		ctx := context.Background()

		exactKey, err := keygen.ExactKey(tenantID, req)
		if err == nil {
			if err := m.exact.Set(ctx, exactKey, resp); err != nil {
				slog.Error("exact cache write-back failed", "error", err)
			}
		}

		promptText := keygen.PromptText(req)
		emb, err := m.embedder.Embed(ctx, promptText)
		if err != nil {
			slog.Error("embedder write-back failed", "error", err)
			return
		}

		hash := sha256.Sum256([]byte(promptText))
		promptHash := fmt.Sprintf("%x", hash)

		if err := m.semantic.Set(ctx, tenantID, req.Model, promptHash, promptText, emb, resp); err != nil {
			slog.Error("semantic cache write-back failed", "error", err)
		}
	}()
}
