module github.com/ananm2958/llm-gateway

go 1.22

require (
	github.com/gin-gonic/gin v1.10.0
	github.com/jackc/pgx/v5 v5.7.1
	github.com/pgvector/pgvector-go v0.2.0
	github.com/prometheus/client_golang v1.20.5
	github.com/redis/go-redis/v9 v9.7.0
	github.com/sony/gobreaker v1.0.0
	go.opentelemetry.io/otel v1.31.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.31.0
	go.opentelemetry.io/otel/exporters/prometheus v0.55.0
	go.opentelemetry.io/otel/sdk v1.31.0
	google.golang.org/grpc v1.67.1
	gopkg.in/yaml.v3 v3.0.1
)
