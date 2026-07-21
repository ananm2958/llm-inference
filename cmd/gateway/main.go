package main

import (
    "context"
    "fmt"
    "log"
    "time"
    "github.com/ananm2958/llm-gateway/internal/api"
    "github.com/ananm2958/llm-gateway/internal/cache"
    "github.com/ananm2958/llm-gateway/internal/cache/embedding"
    "github.com/ananm2958/llm-gateway/internal/cache/exact"
    "github.com/ananm2958/llm-gateway/internal/cache/semantic"
    "github.com/ananm2958/llm-gateway/internal/config"
    "github.com/ananm2958/llm-gateway/internal/providers"
    "github.com/ananm2958/llm-gateway/internal/routing"
    "github.com/ananm2958/llm-gateway/internal/telemetry"
    "github.com/ananm2958/llm-gateway/internal/usage"
    "github.com/jackc/pgx/v5/pgxpool"
)
func main() {
 ctx:=context.Background(); cfg,err:=config.Load("config.yaml");if err!=nil{log.Fatalf("config load failed: %v",err)}
 shutdown,err:=telemetry.Init(ctx,"llm-gateway",cfg.Telemetry.OTelEndpoint);if err!=nil {log.Printf("telemetry disabled: %v",err)} else {defer func(){shutdownCtx,cancel:=context.WithTimeout(context.Background(),5*time.Second);defer cancel();_ = shutdown(shutdownCtx)}()}
 pool,err:=pgxpool.New(ctx,cfg.Database.ConnStr);if err!=nil{log.Fatalf("database pool init failed: %v",err)};defer pool.Close()
 providerList,err:=providers.FromConfig(cfg.Providers);if err!=nil{log.Fatalf("provider config failed: %v",err)}
 registry:=providers.NewRegistry(providerList); router:=routing.NewRouter(registry,cfg.BuildRoutingPolicy())
 cm:=cache.NewManager(exact.New(cfg.Redis.Addr,cfg.Redis.Password,cfg.Redis.DB),semantic.New(pool),embedding.New(cfg.Embedding.BaseURL,cfg.Embedding.Model))
 ginRouter:=api.NewRouter(cfg,router,providerList,cm,usage.NewRecorder(pool),telemetry.NewCostCalculator(pool)); log.Printf("Gateway listening on :%d",cfg.Server.Port);if err:=ginRouter.Run(fmt.Sprintf(":%d",cfg.Server.Port));err!=nil{log.Fatal(err)}
}
