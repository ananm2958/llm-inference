# LLM Inference Gateway

A unified OpenAI-compatible API gateway over multiple model backends. Provides policy-based routing, fallback logic, caching, and end-to-end observability — enabling downstream teams to switch models without changing their integration code.

---

## Features

- **OpenAI-compatible API** — drop-in replacement for `/v1/chat/completions`, `/v1/completions`, and `/v1/models`
- **Policy-based routing** — route requests by tenant, model, cost tier, or latency target
- **Fallback chains** — automatically retry across providers on failure with circuit breaker protection
- **Two-layer caching** — exact-match (Redis) and semantic similarity (pgvector) to eliminate redundant LLM calls
- **Full observability** — distributed tracing via OpenTelemetry, metrics via Prometheus, dashboards via Grafana
- **Per-tenant cost tracking** — token usage and USD spend recorded per request, queryable for billing exports

---

## Tech Stack

| Layer | Technology |
|---|---|
| Gateway API | Go (Gin) |
| Local inference | vLLM / Ollama |
| Exact-match cache | Redis |
| Semantic cache | PostgreSQL + pgvector |
| Embeddings | Ollama (`nomic-embed-text`) |
| Tracing | OpenTelemetry + OTel Collector |
| Metrics | Prometheus |
| Dashboards | Grafana |

---

## Project Structure

```
llm-gateway/
├── cmd/
│   └── gateway/
│       └── main.go
├── internal/
│   ├── api/
│   │   ├── router.go
│   │   ├── middleware/
│   │   │   ├── auth.go
│   │   │   ├── logging.go
│   │   │   └── telemetry.go
│   │   └── handlers/
│   │       ├── chat.go
│   │       ├── completions.go
│   │       └── models.go
│   ├── cache/
│   │   ├── cache.go
│   │   ├── exact/
│   │   │   └── exact.go
│   │   ├── semantic/
│   │   │   └── semantic.go
│   │   └── keygen/
│   │       └── keygen.go
│   ├── config/
│   │   └── config.go
│   ├── embedding/
│   │   └── embedder.go
│   ├── providers/
│   │   ├── provider.go
│   │   ├── registry.go
│   │   ├── openai/
│   │   │   └── openai.go
│   │   └── vllm/
│   │       └── vllm.go
│   ├── routing/
│   │   ├── router.go
│   │   ├── policy.go
│   │   └── fallback.go
│   ├── telemetry/
│   │   ├── otel.go
│   │   ├── metrics.go
│   │   └── cost.go
│   └── usage/
│       └── recorder.go
├── schema.sql
├── config.yaml
├── docker-compose.yml
├── otel-collector-config.yaml
├── prometheus.yml
├── .gitignore
├── go.mod
└── go.sum
```

---

## Prerequisites

- Windows with WSL2 (Ubuntu)
- Docker Desktop with WSL2 integration enabled
- Go 1.22+
- Python 3.11+
- NVIDIA GPU (for vLLM) — or Ollama for CPU-based local dev

---

## Getting Started

### 1. Clone the repository

```bash
git clone https://github.com/yourorg/llm-gateway.git
cd llm-gateway
```

### 2. Start infrastructure

```bash
docker compose up -d
```

This starts PostgreSQL, Redis, OTel Collector, Prometheus, and Grafana.

### 3. Run the database schema

```bash
psql -h localhost -U gateway -d llm_gateway -f schema.sql
```

### 4. Pull a local model (if using Ollama)

```bash
ollama pull mistral
ollama pull nomic-embed-text   # required for semantic cache
ollama serve
```

### 5. Configure the gateway

Copy and edit the config:

```bash
cp config.yaml config.local.yaml
# Edit providers, API keys, and endpoints to match your environment
```

### 6. Run the gateway

```bash
go run cmd/gateway/main.go
```

The gateway starts on `http://localhost:8080`.

---

## Configuration

`config.yaml` controls all runtime behaviour:

```yaml
server:
  port: 8080
  api_keys:
    - "sk-gateway-dev-key-1"

providers:
  - name: vllm-local
    type: vllm
    base_url: "http://localhost:11434/v1"
    models:
      - "mistral"
    timeout_seconds: 60

  - name: openai
    type: openai
    base_url: "https://api.openai.com/v1"
    api_key: "${OPENAI_API_KEY}"
    models:
      - "gpt-4o"
    timeout_seconds: 30

routing:
  default_policy:
    provider_chain:
      - "vllm-local"
      - "openai"
    strategy: "priority"   # priority | round_robin

embedding:
  base_url: "http://localhost:11434/v1"
  model: "nomic-embed-text"

database:
  conn_str: "postgres://gateway:gateway@localhost:5432/llm_gateway"

telemetry:
  otel_endpoint: "localhost:4317"
```

---

## API Reference

All routes under `/v1` require an `Authorization: Bearer <key>` header.

### Chat completions

```
POST /v1/chat/completions
```

Fully OpenAI-compatible. Any client using the OpenAI SDK can point at this gateway with no code changes.

```bash
curl http://localhost:8080/v1/chat/completions \
  -H "Authorization: Bearer sk-gateway-dev-key-1" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "mistral",
    "messages": [{"role": "user", "content": "Hello"}]
  }'
```

**Response headers:**

| Header | Description |
|---|---|
| `X-Cache` | `exact`, `semantic`, or `miss` |
| `X-Trace-ID` | OTel trace ID for log correlation |

**Bypass cache for a single request:**

```bash
-H "X-Cache-Bypass: true"
```

### List models

```
GET /v1/models
```

Returns all models across all configured providers.

### Health check

```
GET /health
```

No auth required. Returns `{"status": "ok"}` when the gateway is running.

### Prometheus metrics

```
GET /metrics
```

Exposes all gateway metrics in Prometheus text format.

---

## Caching

Requests pass through two cache layers before hitting a provider:

1. **Exact-match (Redis)** — SHA-256 hash of the normalized request. TTL 24 hours. Sub-millisecond lookup.
2. **Semantic (pgvector)** — cosine similarity search over prompt embeddings. Configurable similarity threshold (default `0.92`). Catches paraphrased or slightly reworded prompts.

On a cache miss the response is written back to both caches asynchronously so the response path is never blocked.

---

## Routing & Fallback

The routing engine selects a provider chain per request based on the active policy for the tenant. Policies are stored in PostgreSQL and support:

- **Priority** — try providers in order, use the first healthy one
- **Round-robin** — rotate across providers evenly
- **Model overrides** — rewrite the requested model to a different provider/model for a specific tenant

Each provider is protected by a circuit breaker. After a configurable failure rate threshold is exceeded the circuit opens and that provider is skipped until it recovers.

### Routing policy tests

The routing tests use deterministic provider doubles and cover priority-policy
fallback behavior:

- `TestPriorityPolicyFallbackMaintainsSuccessRateAndCountsLLMCalls` verifies
  that requests succeed through the fallback when the primary fails and records
  the total number of LLM calls (one failed primary attempt plus one fallback
  attempt per request).
- `TestPriorityPolicyFallbackIncludesFailedAttemptInTailLatency` verifies that
  p95 request latency includes the time spent on a failed primary call before
  the fallback returns successfully.

Run the focused routing test suite with:

```bash
go test ./internal/routing
```

---

## Observability

| Signal | Where to view |
|---|---|
| Traces | OTel Collector → `http://localhost:4317` |
| Metrics | Prometheus → `http://localhost:9090` |
| Dashboards | Grafana → `http://localhost:3000` (admin / admin) |

### Key metrics

| Metric | Description |
|---|---|
| `llm_gateway_requests_total` | Total requests by tenant, model, provider, status |
| `llm_gateway_request_latency_ms` | End-to-end latency histogram |
| `llm_gateway_tokens_prompt` | Prompt tokens consumed |
| `llm_gateway_tokens_completion` | Completion tokens produced |
| `llm_gateway_cache_hits_total` | Cache hits by type |
| `llm_gateway_cost_microdollars_total` | Estimated spend (divide by 1,000,000 for USD) |
| `llm_gateway_circuit_breaker_open` | Circuit breaker state per provider |

---

## Cost Tracking

Every request writes a row to `usage_events` in Postgres. Costs are calculated at insert time from the `model_pricing` table.

**Spend by tenant this month:**

```sql
SELECT
    tenant_id,
    SUM(cost_usd)     AS total_cost_usd,
    SUM(total_tokens) AS total_tokens,
    COUNT(*)          AS total_requests
FROM  usage_events
WHERE created_at >= date_trunc('month', NOW())
GROUP BY tenant_id
ORDER BY total_cost_usd DESC;
```

**Cache savings:**

```sql
SELECT
    tenant_id,
    cache_type,
    COUNT(*) AS requests_saved
FROM  usage_events
WHERE cache_type IS NOT NULL
  AND created_at >= date_trunc('month', NOW())
GROUP BY tenant_id, cache_type;
```

