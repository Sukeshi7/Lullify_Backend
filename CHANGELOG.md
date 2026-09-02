# Changelog

All notable changes to Lullify Backend are documented in this file.  
Format: [Keep a Changelog](https://keepachangelog.com/en/1.0.0/)  
Versioning: [Semantic Versioning](https://semver.org/spec/v2.0.0.html)

---

## [1.0.0] — 2026-09-02

### Added
- JWT authentication (register, login, refresh, logout) with double-secret access/refresh tokens
- Role-based access control: `user`, `broadcaster`, `admin`
- Broadcaster registration via `want_broadcaster` flag at signup
- Live stream CRUD with HLS segmentation and sliding window playlist
- Redis queue (LPUSH/BRPOP) for playlist track feeding
- HLS transcoder: reads audio files and writes `.ts` segments + `.m3u8` playlist
- `GET /streams/mine` endpoint to restore broadcaster state after reconnect
- One active stream per broadcaster constraint
- Playlist and track management (CRUD + multipart upload)
- Local and MinIO/S3 storage with auto-provider switching via `STORAGE_PROVIDER`
- Listening history (record + list, JWT-scoped)
- Favorites (add, remove, list, unique constraint)
- Admin panel: list users, delete user, platform stats
- `/metrics` endpoint secured with admin middleware
- Secret metrics path (`METRICS_SECRET_PATH`) for Prometheus scraping without auth
- OpenTelemetry traces on all HTTP handlers with trace_id + span_id in logs
- Prometheus custom metrics: `lullify_active_streams`, `lullify_active_listeners`, `lullify_stream_disconnections_total`, `lullify_http_request_duration_seconds`
- Zerolog structured JSON logging with trace context
- Grafana dashboard provisioned as code (8 panels: streams, listeners, goroutines, HTTP rate, latency p99, error rate, memory)
- Grafana Cloud integration with Railway-hosted Prometheus
- 3 Grafana alert rules: API down, high error rate, high memory usage
- Health endpoints: `/healthz` (liveness) + `/readyz` (readiness with DB + Redis checks)
- Rate limiting on auth endpoints
- CORS whitelist middleware
- Docker multi-stage build (alpine runtime)
- Docker Compose full observability stack (API + PostgreSQL + Redis + OTEL Collector + Prometheus + Loki + Grafana)
- Railway CD auto-deploy on `develop` branch
- OpenAPI 3.0 specification (`docs/openapi.yaml`)
- Architecture Decision Records ADR-001 to ADR-005
- Unit test coverage: 82.9% on testable packages
- Integration tests: auth flows, stream lifecycle

### Security
- JWT typ claim validation (access vs refresh token type enforcement)
- Path traversal protection in local storage
- Admin-only endpoints enforced via Bearer token middleware
- Secret metrics scraping path (not documented in public API)
- Rate limiting on `/auth/register` and `/auth/login`

---

## [0.3.0] — 2026-08-24 — Sprint 6

### Added
- Admin handler: list users, delete user, platform stats
- Health check handler (liveness + readiness)
- `/metrics` admin-protected endpoint
- Broadcaster role enforcement on stream create/start
- Rate limiter middleware on auth routes

---

## [0.2.0] — 2026-08-15 — Sprint 5

### Added
- OpenTelemetry SDK integration (traces on all HTTP handlers)
- Prometheus custom metrics (active streams, listeners, disconnections, HTTP duration)
- Zerolog structured JSON logging with trace_id + span_id
- OTEL Collector + Prometheus + Loki + Grafana in Docker Compose
- Grafana dashboard provisioned as code
- Favorites domain + application + HTTP handler
- Listening history domain + application + HTTP handler

---

## [0.1.0] — 2026-08-07 — Sprint 1-4

### Added
- Go project structure with Clean Architecture (domain / application / infrastructure)
- PostgreSQL repositories (users, streams, playlists, tracks, history, favorites)
- JWT authentication (register + login + refresh + logout)
- Stream engine with goroutines and channels (HLS segmentation)
- Redis queue for playlist track feeding
- Local and MinIO storage
- GitHub Actions CI (lint + test + build)
- Railway deployment on `develop` branch
- Flutter mobile app (authentication, stream list, HLS player, broadcaster dashboard)