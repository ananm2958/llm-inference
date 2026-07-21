package config
import (
 "os"
 "path/filepath"
 "testing"
 "github.com/ananm2958/llm-gateway/internal/routing"
)
func TestLoadAndBuildRoutingPolicy(t *testing.T) { t.Setenv("TEST_GATEWAY_KEY","secret"); path:=filepath.Join(t.TempDir(),"config.yaml"); data:=[]byte("server:\n  api_keys: [${TEST_GATEWAY_KEY}]\nrouting:\n  default_policy:\n    provider_chain: [first, second]\n    strategy: round_robin\n    model_overrides:\n      requested:\n        provider_name: second\n        model_name: actual\n");if err:=os.WriteFile(path,data,0600);err!=nil{t.Fatal(err)};cfg,err:=Load(path);if err!=nil{t.Fatal(err)};if cfg.Server.APIKeys[0]!="secret"{t.Fatal("environment was not expanded")};p:=cfg.BuildRoutingPolicy();if p.Strategy!=routing.StrategyRoundRobin||p.ModelOverrides["requested"].ModelName!="actual"{t.Fatalf("unexpected policy: %#v",p)} }
