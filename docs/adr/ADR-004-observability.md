# ADR-004 — OpenTelemetry + Prometheus + Grafana + Loki

**Date:** 2025-01-01  
**Status:** Accepted

## Context

The project requires a complete observability strategy: distributed traces, structured logs, and business metrics on a Grafana dashboard.

## Decision

We use the OpenTelemetry standard for traces, Prometheus for metrics scraping, Loki for log aggregation, and Grafana for visualization — all orchestrated via Docker Compose.

## Rationale

- **OTEL** is the industry standard for distributed tracing — vendor-neutral, supported by all major backends
- **Prometheus** pull model fits our architecture: API exposes `/metrics`, Prometheus scrapes on schedule
- **zerolog** produces structured JSON logs at near-zero allocation cost
- **Loki** integrates natively with Grafana, avoiding Elasticsearch complexity
- The full stack runs locally via Docker Compose for development

## Alternatives Considered

- **Datadog**: Full-featured but expensive and vendor lock-in
- **Jaeger standalone**: Good for traces but no metrics or logs
- **ELK Stack**: More powerful log analysis but heavier resource usage

## Consequences

- Local stack requires ~500MB RAM (Prometheus + Grafana + Loki + OTEL Collector)
- `/metrics` endpoint must be secured in production (admin-only middleware added)
- Grafana dashboards are provisioned as code (`grafana/provisioning/`) for reproducibility