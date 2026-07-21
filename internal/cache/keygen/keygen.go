package keygen

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/ananm2958/llm-gateway/internal/providers"
)

type NormalizedRequest struct {
	Model       string              `json:"model"`
	Messages    []providers.Message `json:"messages"`
	Temperature float64             `json:"temperature"`
	MaxTokens   int                 `json:"max_tokens"`
}

func Normalize(req *providers.ChatRequest) NormalizedRequest {
	return NormalizedRequest{
		Model:       req.Model,
		Messages:    req.Messages,
		Temperature: req.Temperature,
		MaxTokens:   req.MaxTokens,
	}
}

func ExactKey(tenantID string, req *providers.ChatRequest) (string, error) {
	norm := Normalize(req)

	data, err := json.Marshal(norm)
	if err != nil {
		return "", err
	}

	hash := sha256.Sum256(data)
	return fmt.Sprintf("llm:exact:%s:%x", tenantID, hash), nil
}

func PromptText(req *providers.ChatRequest) string {
	order := map[string]int{"system": 0, "user": 1, "assistant": 2}
	msgs := make([]providers.Message, len(req.Messages))
	copy(msgs, req.Messages)

	sort.SliceStable(msgs, func(i, j int) bool {
		return order[msgs[i].Role] < order[msgs[j].Role]
	})

	var result string
	for _, m := range msgs {
		result += fmt.Sprintf("[%s]: %s\n", m.Role, m.Content)
	}
	return result
}
