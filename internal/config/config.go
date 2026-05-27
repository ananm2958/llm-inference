package config

import (
    "os"
    "gopkg.in/yaml.v3"
)

type ServerConfig struct {
    Port    int      `yaml:"port"`
    APIKeys []string `yaml:"api_keys"`
}

type ProviderConfig struct {
    Name           string   `yaml:"name"`
    Type           string   `yaml:"type"`
    BaseURL        string   `yaml:"base_url"`
    APIKey         string   `yaml:"api_key"`
    Models         []string `yaml:"models"`
    TimeoutSeconds int      `yaml:"timeout_seconds"`
}

type Config struct {
    Server    ServerConfig     `yaml:"server"`
    Providers []ProviderConfig `yaml:"providers"`
}

func Load(path string) (*Config, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, err
    }

 
    expanded := os.ExpandEnv(string(data))

    var cfg Config
    if err := yaml.Unmarshal([]byte(expanded), &cfg); err != nil {
        return nil, err
    }
    return &cfg, nil
}