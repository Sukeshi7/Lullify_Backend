# Lullify Backend

> Real-time lo-fi radio streaming backend — Go API with Clean Architecture, HLS streaming, and full observability stack.

[![CI](https://github.com/Sukeshi7/Lullify_Backend/actions/workflows/ci.yml/badge.svg)](https://github.com/Sukeshi7/Lullify_Backend/actions)
[![Go](https://img.shields.io/badge/Go-1.25-00ADD8?logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-blue)](LICENSE)

**Production API**: https://lullify.up.railway.app  
**Grafana Dashboard**: https://epickayak2184.grafana.net  
**OpenAPI Spec**: [`docs/openapi.yaml`](docs/openapi.yaml)

---

## Table of Contents

- [Overview](#overview)
- [Architecture](#architecture)
- [Tech Stack](#tech-stack)
- [Getting Started](#getting-started)
- [Environment Variables](#environment-variables)
- [API Reference](#api-reference)
- [Streaming Pipeline](#streaming-pipeline)
- [Observability](#observability)
- [Testing](#testing)
- [Deployment](#deployment)
- [Database Migrations](#database-migrations)
- [Project Structure](#project-structure)
- [Architecture Decision Records](#architecture-decision-records)
- [Known Limitations](#known-limitations)

---

## Overview

Lullify is a real-time audio streaming platform for lo-fi and vaporwave radio. The backend is a Go API that handles:

- **Live HLS streaming** — broadcasters upload audio tracks, the engine segments them into `.ts` chunks and serves `.m3u8` playlists to listeners
- **Multi-role auth** — users can register as listeners or broadcasters; admins manage the platform
- **Full observability** — distributed traces (OTEL), structured logs (zerolog), custom Prometheus metrics, Grafana dashboards and alerts

---

## Architecture

The project follows **Clean Architecture** with Domain-Driven Design, enforcing strict separation between layers:

```
┌─────────────────────────────────────────────────────────┐
│                    HTTP Handlers                         │
│          (infrastructure/http — no business logic)       │
└──────────────────────┬──────────────────────────────────┘
                       │
┌──────────────────────▼──────────────────────────────────┐
│                   Use Cases                              │
│        (application — orchestrates domain logic)         │
└──────────────────────┬──────────────────────────────────┘
                       │
┌──────────────────────▼──────────────────────────────────┐
│                    Domain                                │
│   (entities, errors, repository interfaces — pure Go)    │
└──────────────────────┬──────────────────────────────────┘
                       │
┌──────────────────────▼──────────────────────────────────┐
│                 Infrastructure                           │
│    (postgres, redis, storage, stream engine, token)      │
└─────────────────────────────────────────────────────────┘
```

Each layer depends only on the layer below it. The domain layer has **zero external dependencies** — it is fully unit-testable without any database or network.

---

## Tech Stack

| Component | Technology |
|---|---|
| Language | Go 1.25 |
| HTTP server | `net/http` (stdlib) |
| Database | PostgreSQL 16 |
| Cache / Queue | Redis 7 |
| Auth | JWT (HS256, double-secret access/refresh) |
| Streaming | HLS (`.m3u8` + `.ts` segments) |
| Storage | Local filesystem or MinIO (S3-compatible) |
| Traces | OpenTelemetry SDK + OTLP HTTP exporter |
| Metrics | Prometheus (`client_golang`) |
| Logs | zerolog (structured JSON) |
| Dashboards | Grafana (provisioned as code) |
| Containerization | Docker multi-stage (alpine runtime) |
| CI/CD | GitHub Actions + Railway |

---

## Getting Started

### Prerequisites

- Go 1.25+
- Docker & Docker Compose
- Make (optional)

### Local development

```bash
# Clone the repository
git clone https://github.com/Sukeshi7/Lullify_Backend.git
cd Lullify_Backend

# Copy and edit environment variables
cp .env.example .env

# Start all services (API + PostgreSQL + Redis + observability stack)
docker compose up -d

# The API is available at http://localhost:8080
# Grafana at http://localhost:3000 (admin / lullify_admin)
# Prometheus at http://localhost:9090
```

### Running without Docker

```bash
# Start only the dependencies
docker compose up -d postgres redis

# Run the API
go run main.go
```

---

## Environment Variables

| Variable | Default | Description |
|---|---|---|
| `PORT` | `8080` | HTTP server port |
| `ENV` | `development` | Environment name |
| `DATABASE_URL` | `postgres://lullify:password@localhost:5432/lullify` | PostgreSQL connection string |
| `REDIS_URL` | `redis://localhost:6379` | Redis connection string |
| `JWT_ACCESS_SECRET` | `changeme-access` | JWT signing secret for access tokens |
| `JWT_REFRESH_SECRET` | `changeme-refresh` | JWT signing secret for refresh tokens |
| `JWT_ACCESS_EXPIRY` | `15m` | Access token TTL |
| `JWT_REFRESH_EXPIRY` | `168h` | Refresh token TTL |
| `STORAGE_PROVIDER` | `local` | Storage backend: `local`, `minio`, `s3` |
| `STORAGE_PATH` | `/data/audio` | Local storage base path |
| `MINIO_ENDPOINT` | `localhost:9000` | MinIO endpoint |
| `MINIO_ACCESS_KEY` | — | MinIO access key |
| `MINIO_SECRET_KEY` | — | MinIO secret key |
| `MINIO_BUCKET` | `lullify-audio` | MinIO bucket name |
| `MINIO_USE_SSL` | `false` | Enable TLS for MinIO |
| `MAX_UPLOAD_SIZE_BYTES` | `52428800` | Max audio file size (50 MB) |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | `http://localhost:4318` | OTEL Collector endpoint (leave empty for no-op) |
| `OTEL_SERVICE_NAME` | `lullify-backend` | Service name in traces |
| `CORS_ALLOWED_ORIGINS` | `http://localhost:*` | Comma-separated allowed CORS origins |
| `RATE_LIMIT_RPS` | `10` | Rate limit requests per second on auth routes |
| `RATE_LIMIT_BURST` | `20` | Rate limit burst size |
| `METRICS_SECRET_PATH` | — | Secret path for Prometheus scraping (e.g. `/metrics-xk9p2q7m`) |

---

## API Reference

Full spec available in [`docs/openapi.yaml`](docs/openapi.yaml) — open it at [editor.swagger.io](https://editor.swagger.io) for interactive docs.

### Authentication

| Method | Path | Auth | Description |
|---|---|---|---|
| `POST` | `/api/v1/auth/register` | — | Register (set `want_broadcaster: true` for broadcaster role) |
| `POST` | `/api/v1/auth/login` | — | Login |
| `POST` | `/api/v1/auth/refresh` | — | Refresh access token |
| `POST` | `/api/v1/auth/logout` | Bearer | Logout |

### Streams

| Method | Path | Auth | Description |
|---|---|---|---|
| `GET` | `/api/v1/streams` | — | List active streams |
| `GET` | `/api/v1/streams/mine` | Bearer | List my active streams (broadcaster) |
| `POST` | `/api/v1/streams` | Bearer (broadcaster) | Create a stream |
| `POST` | `/api/v1/streams/{id}/start` | Bearer (broadcaster) | Start streaming |
| `POST` | `/api/v1/streams/{id}/stop` | Bearer (broadcaster) | Stop streaming |
| `GET` | `/streams/{id}/playlist.m3u8` | — | HLS playlist |
| `GET` | `/streams/{id}/{segment}` | — | HLS segment |

### Playlists & Tracks

| Method | Path | Auth | Description |
|---|---|---|---|
| `POST` | `/api/v1/playlists` | Bearer | Create playlist |
| `GET` | `/api/v1/playlists` | Bearer | List my playlists |
| `POST` | `/api/v1/playlists/{id}/tracks` | Bearer | Upload track (multipart) |

### Favorites & History

| Method | Path | Auth | Description |
|---|---|---|---|
| `GET` | `/api/v1/favorites` | Bearer | List favorites |
| `POST` | `/api/v1/favorites` | Bearer | Add favorite |
| `DELETE` | `/api/v1/favorites/{streamId}` | Bearer | Remove favorite |
| `GET` | `/api/v1/history` | Bearer | Listening history |
| `POST` | `/api/v1/history` | Bearer | Record listening event |

### Admin

| Method | Path | Auth | Description |
|---|---|---|---|
| `GET` | `/api/v1/admin/users` | Bearer (admin) | List all users |
| `DELETE` | `/api/v1/admin/users/{id}` | Bearer (admin) | Delete user |
| `GET` | `/api/v1/admin/stats` | Bearer (admin) | Platform statistics |
| `GET` | `/metrics` | Bearer (admin) | Prometheus metrics |

### Health

| Method | Path | Description |
|---|---|---|
| `GET` | `/healthz` | Liveness probe |
| `GET` | `/readyz` | Readiness probe (checks DB + Redis) |

---

## Streaming Pipeline

```
Broadcaster uploads audio file
        │
        ▼
POST /playlists/{id}/tracks
        │
        ▼
Track stored on filesystem / MinIO
        │
        ▼
POST /streams/{id}/start
        │
        ├── Redis BRPOP → fetch track file path from queue
        │
        ├── StreamEngine.Start(streamID, filePath)
        │       │
        │       └── goroutine: Transcoder.TranscodeFile()
        │               │
        │               ├── Read audio file in 32KB chunks
        │               ├── Write .ts segments to /tmp/lullify/{streamID}/
        │               └── Update playlist.m3u8 (sliding window of 6 segments)
        │
        ▼
Listeners connect via Flutter
        │
        └── GET /streams/{id}/playlist.m3u8
                └── GET /streams/{id}/segment{N}.ts
```

The HLS sliding window keeps the last 6 segments (12 seconds of audio) on disk. Older segments are deleted automatically to limit disk usage.

---

## Observability

### Local stack

```bash
docker compose up -d
```

| Service | URL | Credentials |
|---|---|---|
| Grafana | http://localhost:3000 | `admin` / `lullify_admin` |
| Prometheus | http://localhost:9090 | — |
| Loki | http://localhost:3100 | — |

### Production (Grafana Cloud)

The production stack uses Railway-hosted Prometheus scraping the API via a secret path, with metrics forwarded to Grafana Cloud.

**Dashboard**: https://epickayak2184.grafana.net/d/lullify-overview

### Custom metrics

| Metric | Type | Description |
|---|---|---|
| `lullify_active_streams` | Gauge | Number of live streams |
| `lullify_active_listeners` | Gauge | Connected listeners |
| `lullify_stream_disconnections_total` | Counter | Disconnections by type (`normal`/`abrupt`) |
| `lullify_http_request_duration_seconds` | Histogram | Request duration by method/path/status |

### Alert rules

| Alert | Condition | Severity |
|---|---|---|
| Lullify API Down | `up{job="lullify-api"} < 1` for 2m | Critical |
| High Error Rate | 5xx rate > 5% for 2m | Warning |
| High Memory | Heap > 50 MB for 5m | Warning |

---

## Testing

### Unit tests (business logic only)

```bash
go test ./config/... ./internal/application/... ./internal/domain/... \
  ./internal/infrastructure/observability/... ./internal/infrastructure/redis/... \
  ./internal/infrastructure/stream/... ./internal/infrastructure/token/... \
  ./internal/infrastructure/storage/... -coverprofile=coverage

go tool cover -func coverage | grep total
# total: 82.8%
```

### Integration tests (requires PostgreSQL + Redis)

```bash
docker compose up -d postgres redis

go test ./internal/integration/... -v
```

Integration tests cover: full auth flow (register → login → token validation), stream lifecycle (create → start → stop), and duplicate constraint enforcement.

### Coverage notes

| Scope | Coverage |
|---|---|
| Unit tests (testable packages) | **82.8%** |
| Global (all packages) | 36.6% |

The lower global figure reflects that `infrastructure/postgres` (PostgreSQL repositories) and `infrastructure/http` (HTTP handlers) are not covered by unit tests — these packages require a live database and are exercised by integration tests. This is standard practice in Clean Architecture Go projects where the domain and application layers carry the business logic.

---

## Deployment

### Railway (current)

The API is automatically deployed to Railway on every push to `develop`.

**Services on Railway:**
- `api` — Go backend (this repo)
- `postgres` — PostgreSQL 16 database
- `redis` — Redis 7 cache
- `prometheus` — Prometheus scraper (custom Dockerfile)

### Docker

```bash
# Build production image
docker build -f docker/Dockerfile -t lullify-backend .

# Run
docker run -p 8080:8080 \
  -e DATABASE_URL=postgres://... \
  -e REDIS_URL=redis://... \
  -e JWT_ACCESS_SECRET=... \
  -e JWT_REFRESH_SECRET=... \
  -e STORAGE_PROVIDER=local \
  lullify-backend
```

### CD pipeline

GitHub Actions runs on every push:
1. `golangci-lint` — static analysis
2. `go test ./...` — unit + integration tests  
3. `docker build` — validate Dockerfile compiles

Railway auto-deploys on successful CI on `develop`.

---

## Database Migrations

Migrations are plain SQL files in `migrations/`. They run automatically when the postgres container starts (mounted in `/docker-entrypoint-initdb.d/`).

For production (Railway), run them manually via the Railway PostgreSQL query console.

| File | Description |
|---|---|
| `001_create_users.sql` | Users table with role column |
| `002_create_streams.sql` | Streams table |
| `003_create_playlists.sql` | Playlists and tracks tables |
| `004_add_upload_columns_to_tracks.sql` | Format, size, uploaded_by columns |
| `005_create_listening_history.sql` | Listening history table |
| `006_create_favorites.sql` | Favorites table with unique constraint |

---

## Project Structure

```
.
├── main.go                          # Entry point — dependency injection
├── config/                          # 12-Factor configuration (env vars)
├── docs/
│   ├── openapi.yaml                 # OpenAPI 3.0 specification
│   └── adr/                         # Architecture Decision Records
├── docker/
│   ├── Dockerfile                   # Multi-stage build (Go → alpine)
│   └── Dockerfile.prometheus        # Prometheus for Railway
├── migrations/                      # SQL migrations (001–006)
├── grafana/provisioning/            # Grafana dashboards and datasources as code
├── internal/
│   ├── domain/                      # Entities, errors, repository interfaces
│   │   ├── user/
│   │   ├── stream/
│   │   ├── playlist/
│   │   ├── history/
│   │   └── favorite/
│   ├── application/                 # Use cases (business logic)
│   │   ├── user/                    # Register, Login
│   │   ├── stream/                  # Create, Start, Stop
│   │   ├── track/                   # Upload
│   │   ├── history/                 # Record, List
│   │   └── favorite/                # Add, Remove, List
│   ├── infrastructure/
│   │   ├── http/                    # HTTP handlers + router + rate limiter
│   │   ├── postgres/                # PostgreSQL repositories
│   │   ├── redis/                   # Redis client + queue (LPUSH/BRPOP)
│   │   ├── storage/                 # Local + MinIO storage
│   │   ├── stream/                  # HLS engine + segmenter + transcoder
│   │   ├── token/                   # JWT service
│   │   └── observability/           # OTEL + Prometheus + zerolog
│   └── integration/                 # Integration tests
├── otel-collector-config.yml
├── prometheus.yml
├── prometheus-cloud.yml
└── loki-config.yml
```

---

## Architecture Decision Records

| ADR | Decision |
|---|---|
| [ADR-001](docs/adr/ADR-001-go-backend.md) | Go as backend language |
| [ADR-002](docs/adr/ADR-002-clean-architecture.md) | Clean Architecture with DDD |
| [ADR-003](docs/adr/ADR-003-hls-streaming.md) | HLS for audio streaming |
| [ADR-004](docs/adr/ADR-004-observability.md) | OpenTelemetry + Prometheus + Grafana + Loki |
| [ADR-005](docs/adr/ADR-005-flutter-riverpod.md) | Flutter with Riverpod for state management |

---

## Known Limitations

| Item | Status | Notes |
|---|---|---|
| Playlist use case layer | Missing | Create playlist logic lives directly in the HTTP handler — no `CreateUseCase`. Functionally complete, architecturally inconsistent with other domains. Planned refactor post-v1. |
| Role upgrade flow | Not implemented | Users cannot upgrade from `listener` to `broadcaster` after registration. Role is set at signup via `want_broadcaster`. A role-change request flow is planned. |
| Global test coverage | 36.6% | PostgreSQL repositories and HTTP handlers are not unit-tested (require a live DB). Business logic coverage is **82.8%**. Integration tests cover the remaining flows. |
| HLS latency | ~4 seconds | Inherent to HLS segmentation (2s segments × 2 buffer). Not a bug. |
| Anonymous role | Declared, not enforced | The `anonymous` role exists in the spec but is not implemented in auth middleware. |
