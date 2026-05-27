func main() {
    ctx := context.Background()

    cfg, err := config.Load("config.yaml")
    if err != nil {
        log.Fatalf("config load failed: %v", err)
    }

    shutdown, err := telemetry.Init(ctx, "llm-gateway", cfg.Telemetry.OTelEndpoint)
    if err != nil {
        log.Fatalf("telemetry init failed: %v", err)
    }
    defer func() {
        shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()
        shutdown(shutdownCtx)
    }()

   
    pool, err := pgxpool.New(ctx, cfg.Database.ConnStr)
    if err != nil {
        log.Fatalf("db init failed: %v", err)
    }
    defer pool.Close()

  
    recorder  := usage.NewRecorder(pool)
    costCalc  := telemetry.NewCostCalculator(pool)

    ginRouter := api.NewRouter(cfg, router, providerList, recorder, costCalc)

    log.Printf("Gateway listening on :%d", cfg.Server.Port)
    ginRouter.Run(fmt.Sprintf(":%d", cfg.Server.Port))
}