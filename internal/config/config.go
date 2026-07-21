package config

import (
    "os"
    "github.com/ananm2958/llm-gateway/internal/routing"
    "gopkg.in/yaml.v3"
)
type ServerConfig struct { Port int `yaml:"port"`; APIKeys []string `yaml:"api_keys"` }
type ProviderConfig struct { Name string `yaml:"name"`; Type string `yaml:"type"`; BaseURL string `yaml:"base_url"`; APIKey string `yaml:"api_key"`; Models []string `yaml:"models"`; TimeoutSeconds int `yaml:"timeout_seconds"` }
type RoutingConfig struct { DefaultPolicy DefaultPolicyConfig `yaml:"default_policy"` }
type DefaultPolicyConfig struct { ProviderChain []string `yaml:"provider_chain"`; Strategy string `yaml:"strategy"`; ModelOverrides map[string]ModelOverrideConfig `yaml:"model_overrides"` }
type ModelOverrideConfig struct { ProviderName string `yaml:"provider_name"`; ModelName string `yaml:"model_name"` }
type TelemetryConfig struct { OTelEndpoint string `yaml:"otel_endpoint"` }
type DatabaseConfig struct { ConnStr string `yaml:"conn_str"` }
type EmbeddingConfig struct { BaseURL string `yaml:"base_url"`; Model string `yaml:"model"` }
type RedisConfig struct { Addr string `yaml:"addr"`; Password string `yaml:"password"`; DB int `yaml:"db"` }
type Config struct { Server ServerConfig `yaml:"server"`; Providers []ProviderConfig `yaml:"providers"`; Routing RoutingConfig `yaml:"routing"`; Telemetry TelemetryConfig `yaml:"telemetry"`; Database DatabaseConfig `yaml:"database"`; Embedding EmbeddingConfig `yaml:"embedding"`; Redis RedisConfig `yaml:"redis"` }
func Load(path string) (*Config, error) { data, err := os.ReadFile(path); if err != nil{return nil,err}; var cfg Config; if err=yaml.Unmarshal([]byte(os.ExpandEnv(string(data))), &cfg);err != nil{return nil,err}; return &cfg,nil }
func (c *Config) BuildRoutingPolicy() *routing.RoutingPolicy {
    p := &routing.RoutingPolicy{TenantID:"default", ProviderChain: append([]string(nil), c.Routing.DefaultPolicy.ProviderChain...), Strategy:routing.RoutingStrategy(c.Routing.DefaultPolicy.Strategy), ModelOverrides: map[string]routing.ModelOverride{}}
    if p.Strategy == "" { p.Strategy = routing.StrategyPriority }
    for model, override := range c.Routing.DefaultPolicy.ModelOverrides { p.ModelOverrides[model]=routing.ModelOverride{ProviderName:override.ProviderName, ModelName:override.ModelName} }
    return p
}
