

CREATE EXTENSION IF NOT EXISTS vector;




CREATE TABLE routing_policies (
    id              BIGSERIAL PRIMARY KEY,
    tenant_id       TEXT        NOT NULL DEFAULT 'default',
    provider_chain  TEXT[]      NOT NULL,           
    strategy        TEXT        NOT NULL DEFAULT 'priority',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (tenant_id)
);

CREATE TABLE model_overrides (
    id              BIGSERIAL PRIMARY KEY,
    tenant_id       TEXT        NOT NULL DEFAULT 'default',
    requested_model TEXT        NOT NULL,           
    provider_name   TEXT        NOT NULL,           
    actual_model    TEXT        NOT NULL,          
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (tenant_id, requested_model)
);

INSERT INTO routing_policies (tenant_id, provider_chain, strategy)
VALUES ('default', ARRAY['vllm-local', 'openai'], 'priority');

CREATE TABLE semantic_cache (
    id              BIGSERIAL PRIMARY KEY,
    tenant_id       TEXT        NOT NULL DEFAULT 'default',
    model           TEXT        NOT NULL,
    prompt_hash     TEXT        NOT NULL,
    prompt_text     TEXT        NOT NULL,
    embedding       vector(1536),
    response_json   JSONB       NOT NULL,
    hit_count       INT         NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_hit_at     TIMESTAMPTZ
);

CREATE INDEX ON semantic_cache
    USING ivfflat (embedding vector_cosine_ops)
    WITH (lists = 100);

CREATE INDEX ON semantic_cache (tenant_id, model);

CREATE TABLE cache_stats (
    id          BIGSERIAL PRIMARY KEY,
    tenant_id   TEXT        NOT NULL DEFAULT 'default',
    cache_type  TEXT        NOT NULL,             
    hit         BOOLEAN     NOT NULL,
    model       TEXT        NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);




CREATE TABLE usage_events (
    id                BIGSERIAL PRIMARY KEY,
    tenant_id         TEXT        NOT NULL DEFAULT 'default',
    request_id        TEXT        NOT NULL,
    model             TEXT        NOT NULL,
    provider          TEXT        NOT NULL,
    prompt_tokens     INT         NOT NULL DEFAULT 0,
    completion_tokens INT         NOT NULL DEFAULT 0,
    total_tokens      INT         NOT NULL DEFAULT 0,
    cache_type        TEXT,                       
    latency_ms        INT         NOT NULL DEFAULT 0,
    status            TEXT        NOT NULL DEFAULT 'success',
    error_message     TEXT,
    cost_usd          NUMERIC(12,8),
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX ON usage_events (tenant_id, created_at DESC);
CREATE INDEX ON usage_events (model, created_at DESC);
CREATE INDEX ON usage_events (provider, created_at DESC);

CREATE TABLE model_pricing (
    id                      BIGSERIAL PRIMARY KEY,
    provider                TEXT           NOT NULL,
    model                   TEXT           NOT NULL,
    prompt_cost_per_1k      NUMERIC(10,6)  NOT NULL,
    completion_cost_per_1k  NUMERIC(10,6)  NOT NULL,
    effective_from          TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    UNIQUE (provider, model, effective_from)
);

INSERT INTO model_pricing (provider, model, prompt_cost_per_1k, completion_cost_per_1k) VALUES
    ('openai',     'gpt-4o',         0.005000, 0.015000),
    ('openai',     'gpt-4o-mini',    0.000150, 0.000600),
    ('openai',     'gpt-3.5-turbo',  0.000500, 0.001500),
    ('vllm-local', 'mistral',        0.000000, 0.000000),
    ('vllm-local', 'llama3',         0.000000, 0.000000),
    ('ollama',     'mistral',        0.000000, 0.000000);




CREATE TABLE tenants (
    id            BIGSERIAL PRIMARY KEY,
    tenant_id     TEXT        NOT NULL UNIQUE,
    name          TEXT        NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    active        BOOLEAN     NOT NULL DEFAULT TRUE
);

CREATE TABLE api_keys (
    id            BIGSERIAL PRIMARY KEY,
    tenant_id     TEXT        NOT NULL REFERENCES tenants(tenant_id),
    key_hash      TEXT        NOT NULL UNIQUE,      
    name          TEXT,                           
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at    TIMESTAMPTZ,
    active        BOOLEAN     NOT NULL DEFAULT TRUE
);

CREATE INDEX ON api_keys (key_hash);

CREATE TABLE tenant_quotas (
    id                  BIGSERIAL PRIMARY KEY,
    tenant_id           TEXT           NOT NULL REFERENCES tenants(tenant_id) UNIQUE,
    monthly_budget_usd  NUMERIC(10,2),              
    rpm_limit           INT,                        
    tpm_limit           INT,                        
    updated_at          TIMESTAMPTZ    NOT NULL DEFAULT NOW()
);

INSERT INTO tenants (tenant_id, name) VALUES ('default', 'Default Tenant');
INSERT INTO tenant_quotas (tenant_id, monthly_budget_usd, rpm_limit)
VALUES ('default', NULL, NULL);