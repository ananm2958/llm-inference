package providers

import (
    "fmt"
    "strings"
    "github.com/ananm2958/llm-gateway/internal/config"
    "github.com/ananm2958/llm-gateway/internal/providers/openai"
    "github.com/ananm2958/llm-gateway/internal/providers/vllm"
)
func FromConfig(cfgs []config.ProviderConfig) ([]Provider, error) {
    list := make([]Provider, 0, len(cfgs))
    for _, cfg := range cfgs { switch strings.ToLower(cfg.Type) { case "vllm": list=append(list, vllm.New(cfg.Name,cfg.BaseURL,cfg.APIKey,cfg.Models,cfg.TimeoutSeconds)); case "openai": list=append(list, openai.New(cfg.Name,cfg.BaseURL,cfg.APIKey,cfg.Models,cfg.TimeoutSeconds)); default: return nil,fmt.Errorf("unsupported provider type %q",cfg.Type) } }
    return list,nil
}
