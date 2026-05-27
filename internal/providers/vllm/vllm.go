package vllm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/ananm2958/llm-gateway/internal/providers"
)

type VLLMProvider struct {
	name    string
	baseURL string
	apiKey  string
	models  []string
	client  *http.Client
}

func New(name, baseURL, apiKey string, models []string, timeoutSecs int) *VLLMProvider {
	return &VLLMProvider{
		name:    name,
		baseURL: baseURL,
		apiKey:  apiKey,
		models:  models,
		client:  &http.Client{Timeout: time.Duration(timeoutSecs) * time.Second},
	}
}

func (p *VLLMProvider) Name() string              { return p.name }
func (p *VLLMProvider) SupportedModels() []string { return p.models }

func (p *VLLMProvider) Chat(ctx context.Context, req *providers.ChatRequest) (*providers.ChatResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if p.apiKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)
	}

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("provider %s returned %d: %s", p.name, resp.StatusCode, string(b))
	}

	var chatResp providers.ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return nil, err
	}
	return &chatResp, nil
}

func (p *VLLMProvider) ChatStream(ctx context.Context, req *providers.ChatRequest) (io.ReadCloser, error) {
	req.Stream = true
	body, _ := json.Marshal(req)

	httpReq, _ := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/chat/completions", bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")
	if p.apiKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)
	}

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("provider %s stream returned %d", p.name, resp.StatusCode)
	}
	return resp.Body, nil
}

func (p *VLLMProvider) Healthy(ctx context.Context) bool {
	req, _ := http.NewRequestWithContext(ctx, "GET", p.baseURL+"/models", nil)
	resp, err := p.client.Do(req)
	if err != nil {
		return false
	}
	resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}
