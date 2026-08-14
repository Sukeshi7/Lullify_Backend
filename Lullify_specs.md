# 📋 Cahier des Charges — Lullify
> Projet Semestriel 5A TL — S2 — Bloc 3  
> École .decode — 2025/2026

---

## Table des matières

1. [Présentation du projet](#1-présentation-du-projet)
2. [Objectifs pédagogiques](#2-objectifs-pédagogiques)
3. [Périmètre fonctionnel](#3-périmètre-fonctionnel)
4. [Architecture technique](#4-architecture-technique)
5. [Stack technologique](#5-stack-technologique)
6. [Modèle de données](#6-modèle-de-données)
7. [API — Endpoints principaux](#7-api--endpoints-principaux)
8. [Observabilité & SRE](#8-observabilité--sre)
9. [Sécurité](#9-sécurité)
10. [Infrastructure & Déploiement](#10-infrastructure--déploiement)
11. [Organisation du projet](#11-organisation-du-projet)
12. [Livrables attendus](#12-livrables-attendus)
13. [Critères de validation RNCP](#13-critères-de-validation-rncp)

---

## 1. Présentation du projet

### 1.1 Contexte

L'industrie du streaming audio en direct connaît une croissance exponentielle. Radios numériques, podcasts live, sessions lo-fi en continu — la maîtrise d'une chaîne de transmission audio temps réel est devenue une compétence critique pour tout Tech Lead.

**Lullify** est une plateforme de radio streaming en temps réel, à l'esthétique lo-fi et onirique, permettant à des diffuseurs de streamer leur musique ou playlists, et à des auditeurs de les écouter en direct depuis une application mobile multi-plateforme.

### 1.2 Identité du projet

| Attribut | Valeur |
|---|---|
| Nom du projet | **Lullify** |
| Univers visuel | Lo-fi, cozy, onirique — inspiré du jeu *Melatonine* |
| Palette | Tons pastel désaturés, teintes nocturnes douces |
| Ambiance | Entre veille et sommeil, flottant, apaisant |

### 1.3 Équipe

| Membre | Rôle principal |
|---|---|
| Membre 1 | Backend Go (lead) + Flutter |
| Membre 2 | Flutter (lead) + Backend Go |

> Les deux membres touchent à l'ensemble du stack. La répartition ci-dessus reflète une dominante, non une exclusivité.

---

## 2. Objectifs pédagogiques

Ce projet vise la validation du **Bloc 3 du RNCP 38822** — *Piloter la mise en production des solutions logicielles et leur évolution* — à travers la conception d'une architecture complexe, son industrialisation et sa supervision.

Les compétences cibles sont :

- **C3.1** — Piloter le système d'intégration continue (CI/CD)
- **C3.2** — Organiser l'élaboration d'un plan de tests itératifs
- **C3.3** — Surveiller les automatisations visant les mises à jour logicielles
- **C3.4** — Piloter le déploiement continu de la solution logicielle
- **C3.5** — Piloter l'optimisation des applications et des environnements (DevOps)
- **C3.6** — Organiser la rédaction de la documentation technique

---

## 3. Périmètre fonctionnel

### 3.1 Rôles utilisateurs

| Rôle | Description |
|---|---|
| **Anonyme** | Consultation limitée : liste des streams actifs, aperçu public |
| **User** | Écoute des streams, gestion de favoris, création de playlists |
| **Diffuseur** | Création et gestion de flux live, upload de sources audio |
| **Admin** | Gestion des utilisateurs, accès aux métriques globales |

### 3.2 Fonctionnalités principales *(notées /15)*

#### 🔐 Authentification & Gestion des utilisateurs
- Inscription / Connexion sécurisée via **JWT** (access token + refresh token)
- Gestion des rôles (RBAC)
- Profil utilisateur éditable
- Endpoints protégés selon le rôle

#### 🎙️ Moteur de Streaming (Go)
- Diffusion audio en temps réel vers N auditeurs simultanés
- Protocole : **HTTP Live Streaming (HLS)** ou **ICY/Icecast-compatible**
- Gestion concurrente des flux via **goroutines + channels**
- Gestion propre des déconnexions et des timeouts via **Context**
- CRUD complet des playlists avec logique de file d'attente (Queue)
- Configuration 100% via variables d'environnement (**12-Factor App**)

#### 📱 Application Mobile Flutter
- **Lecteur audio avancé** : streaming, barre de progression, contrôle du volume, lecture en arrière-plan (Background Services)
- **Interface Diffuseur** : dashboard pour lancer / arrêter un flux live
- **Interface Auditeur** : liste des streams actifs, lecteur intégré, favoris
- Gestion des interruptions (appels entrants, notifications)
- UI/UX moderne, responsive iOS & Android, accessibilité WCAG AA
- State management via **Bloc** ou **Riverpod**

#### 📊 Observabilité & Supervision
- Logs structurés au format **JSON** (sortie stdout, compatible Loki)
- Instrumentation **OpenTelemetry (OTEL)** : traces distribuées + métriques
- Dashboard **Grafana** : utilisateurs en ligne, débit de streaming, taux d'erreurs, temps de réponse API
- Distinction métriques techniques (erreurs 500) vs métriques métier (déconnexions brutales)

### 3.3 Fonctionnalités bonus *(notées /5)*

| Fonctionnalité | Description technique |
|---|---|
| **Mode Offline** | Mise en cache de segments de playlists pour écoute sans réseau |
| **Chat live** | WebSockets — chat entre auditeurs d'un même flux |
| **Kubernetes** | Déploiement sur cluster K8s avec limits/requests |
| **Recommandations** | Algorithme simple basé sur l'historique d'écoute |
| **Transcodage à la volée** | Adaptation du bitrate selon la bande passante (ABR) |

---

## 4. Architecture technique

### 4.1 Vue d'ensemble

```
┌─────────────────────────────────────────────────┐
│                 Flutter App                     │
│         (iOS / Android / Web)                   │
└──────────────────┬──────────────────────────────┘
                   │ HTTPS / HLS / WebSocket
┌──────────────────▼──────────────────────────────┐
│                API Gateway (Go)                 │
│           HTTP REST + gRPC (interne)            │
├─────────────────────────────────────────────────┤
│  Auth Service  │  Stream Engine  │  Playlist    │
│   (JWT/RBAC)   │ (goroutines)    │   Service    │
├─────────────────────────────────────────────────┤
│       PostgreSQL   │   Redis (cache/queue)      │
├─────────────────────────────────────────────────┤
│  OpenTelemetry Collector → Prometheus → Grafana │
│   Logs (JSON stdout) → Loki → Grafana           │
└─────────────────────────────────────────────────┘
```

### 4.2 Clean Architecture (Go)

Le backend suit les principes de **Clean Architecture** et **Domain Driven Design (DDD)** :

```
backend/
├── cmd/
│   └── server/          # Point d'entrée
├── internal/
│   ├── domain/          # Entités métier, interfaces (ports)
│   │   ├── user/
│   │   ├── stream/
│   │   └── playlist/
│   ├── application/     # Use cases
│   ├── infrastructure/  # Implémentations concrètes (DB, cache, OTEL)
│   └── transport/       # HTTP handlers, middlewares, gRPC
├── pkg/                 # Utilitaires partagés
├── config/              # Chargement de config (env vars)
└── docker/
    └── Dockerfile
```

### 4.3 Architecture Flutter

```
mobile/
├── lib/
│   ├── core/            # Config, constantes, thème
│   ├── data/            # Repositories, datasources, modèles
│   ├── domain/          # Entités, use cases
│   ├── presentation/    # Pages, widgets, blocs/cubits
│   └── main.dart
├── android/
├── ios/
└── Dockerfile           # Build CI
```

---

## 5. Stack technologique

### Backend

| Composant | Technologie | Justification |
|---|---|---|
| Langage | **Go 1.23+** | Performance, concurrence native (goroutines), faible empreinte mémoire |
| Framework HTTP | **Chi** ou **Fiber** | Léger, idiomatique Go, middleware-friendly |
| Base de données | **PostgreSQL 16** | Fiabilité, support JSON, transactions ACID |
| Cache / Queue | **Redis 7** | File d'attente des playlists, sessions, pub/sub pour le streaming |
| ORM / Query | **sqlc** + **pgx** | Type-safe, performances supérieures à GORM |
| Auth | **JWT** (golang-jwt) | Stateless, standard industrie |
| Observabilité | **OpenTelemetry Go SDK** | Standard ouvert, vendor-neutral |
| Métriques | **Prometheus** | Compatible OTEL, large adoption |
| Logs | **zerolog** | JSON natif, zéro allocation |
| Traces | **Jaeger** ou **Tempo** | Visualisation des traces distribuées |
| Dashboard | **Grafana** | Visualisation unifiée métriques + logs + traces |

### Mobile

| Composant | Technologie | Justification |
|---|---|---|
| Framework | **Flutter 3.x** | Multi-plateforme iOS/Android depuis un seul codebase |
| Langage | **Dart** | Typé, performant, compilé AOT |
| State Management | **Riverpod** ou **Bloc** | Robuste, testable, séparation claire UI/logique |
| Audio | **just_audio** + **audio_service** | Lecture en arrière-plan, gestion des interruptions |
| HTTP | **Dio** | Intercepteurs, gestion d'erreurs, refresh token |
| WebSocket | **web_socket_channel** | Chat live (bonus) |
| Navigation | **GoRouter** | Déclaratif, deep linking |

### Infrastructure

| Composant | Technologie |
|---|---|
| Conteneurisation | **Docker** (multi-stage, image Alpine/Distroless) |
| Orchestration | **Docker Compose** (dev) / **Kubernetes** (bonus) |
| CI/CD | **GitHub Actions** |
| Registry | **GitHub Container Registry (GHCR)** |
| Reverse Proxy | **Traefik** ou **Nginx** |
| TLS | **Let's Encrypt** (via Traefik) |

---

## 6. Modèle de données

### Entités principales

```sql
-- Utilisateurs
CREATE TABLE users (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email       VARCHAR(255) UNIQUE NOT NULL,
    username    VARCHAR(100) UNIQUE NOT NULL,
    password    VARCHAR(255) NOT NULL,         -- bcrypt hash
    role        VARCHAR(20) NOT NULL DEFAULT 'user', -- user | broadcaster | admin
    created_at  TIMESTAMPTZ DEFAULT NOW(),
    updated_at  TIMESTAMPTZ DEFAULT NOW()
);

-- Streams live
CREATE TABLE streams (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id     UUID REFERENCES users(id) ON DELETE CASCADE,
    title        VARCHAR(255) NOT NULL,
    description  TEXT,
    status       VARCHAR(20) DEFAULT 'offline', -- offline | live | ended
    mount_point  VARCHAR(100) UNIQUE NOT NULL,   -- ex: /streams/chill-vibes
    listener_count INT DEFAULT 0,
    created_at   TIMESTAMPTZ DEFAULT NOW(),
    started_at   TIMESTAMPTZ,
    ended_at     TIMESTAMPTZ
);

-- Playlists
CREATE TABLE playlists (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id    UUID REFERENCES users(id) ON DELETE CASCADE,
    title       VARCHAR(255) NOT NULL,
    is_public   BOOLEAN DEFAULT true,
    created_at  TIMESTAMPTZ DEFAULT NOW()
);

-- Tracks
CREATE TABLE tracks (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title       VARCHAR(255) NOT NULL,
    artist      VARCHAR(255),
    duration    INT,                            -- secondes
    file_url    VARCHAR(500) NOT NULL,
    uploaded_by UUID REFERENCES users(id),
    created_at  TIMESTAMPTZ DEFAULT NOW()
);

-- Playlist items (ordre de lecture)
CREATE TABLE playlist_tracks (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    playlist_id  UUID REFERENCES playlists(id) ON DELETE CASCADE,
    track_id     UUID REFERENCES tracks(id) ON DELETE CASCADE,
    position     INT NOT NULL,
    UNIQUE(playlist_id, position)
);

-- Favoris
CREATE TABLE favorites (
    user_id    UUID REFERENCES users(id) ON DELETE CASCADE,
    stream_id  UUID REFERENCES streams(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, stream_id)
);

-- Historique d'écoute
CREATE TABLE listening_history (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID REFERENCES users(id) ON DELETE CASCADE,
    stream_id   UUID REFERENCES streams(id) ON DELETE SET NULL,
    track_id    UUID REFERENCES tracks(id) ON DELETE SET NULL,
    listened_at TIMESTAMPTZ DEFAULT NOW(),
    duration    INT                                           -- secondes écoutées
);
```

---

## 7. API — Endpoints principaux

### Authentification

| Méthode | Endpoint | Rôle requis | Description |
|---|---|---|---|
| POST | `/api/v1/auth/register` | — | Inscription |
| POST | `/api/v1/auth/login` | — | Connexion, retourne JWT |
| POST | `/api/v1/auth/refresh` | — | Refresh du token |
| POST | `/api/v1/auth/logout` | User | Révocation du token |

### Utilisateurs

| Méthode | Endpoint | Rôle requis | Description |
|---|---|---|---|
| GET | `/api/v1/users/me` | User | Profil courant |
| PATCH | `/api/v1/users/me` | User | Mise à jour du profil |
| GET | `/api/v1/users` | Admin | Liste tous les utilisateurs |
| DELETE | `/api/v1/users/:id` | Admin | Suppression d'un utilisateur |

### Streams

| Méthode | Endpoint | Rôle requis | Description |
|---|---|---|---|
| GET | `/api/v1/streams` | — | Liste des streams actifs |
| GET | `/api/v1/streams/:id` | — | Détails d'un stream |
| POST | `/api/v1/streams` | Diffuseur | Créer un stream |
| PATCH | `/api/v1/streams/:id` | Diffuseur | Mettre à jour |
| DELETE | `/api/v1/streams/:id` | Diffuseur / Admin | Supprimer |
| POST | `/api/v1/streams/:id/start` | Diffuseur | Démarrer le live |
| POST | `/api/v1/streams/:id/stop` | Diffuseur | Arrêter le live |
| GET | `/streams/:mount` | — | Point de streaming audio (HLS/ICY) |

### Playlists

| Méthode | Endpoint | Rôle requis | Description |
|---|---|---|---|
| GET | `/api/v1/playlists` | User | Mes playlists |
| POST | `/api/v1/playlists` | User | Créer une playlist |
| GET | `/api/v1/playlists/:id` | User | Détails + tracks |
| PATCH | `/api/v1/playlists/:id` | User | Modifier |
| DELETE | `/api/v1/playlists/:id` | User | Supprimer |
| POST | `/api/v1/playlists/:id/tracks` | User | Ajouter un track |
| DELETE | `/api/v1/playlists/:id/tracks/:trackId` | User | Retirer un track |
| PATCH | `/api/v1/playlists/:id/tracks/reorder` | User | Réordonner |

### Admin & Métriques

| Méthode | Endpoint | Rôle requis | Description |
|---|---|---|---|
| GET | `/api/v1/admin/metrics` | Admin | Métriques globales |
| GET | `/metrics` | Interne | Endpoint Prometheus |
| GET | `/health` | — | Health check |
| GET | `/ready` | — | Readiness check |

---

## 8. Observabilité & SRE

### 8.1 Logs structurés (JSON)

Chaque log émis par le backend Go est au format JSON via **zerolog** :

```json
{
  "level": "info",
  "timestamp": "2025-10-01T22:14:00Z",
  "service": "lullify-backend",
  "trace_id": "abc123",
  "span_id": "def456",
  "event": "stream.started",
  "stream_id": "uuid",
  "user_id": "uuid",
  "listener_count": 42
}
```

### 8.2 Traces distribuées (OpenTelemetry)

- Chaque requête HTTP est instrumentée avec un **span** OTEL
- Le `trace_id` est propagé de l'app Flutter jusqu'à la base de données
- Les traces sont exportées vers **Jaeger** ou **Grafana Tempo**

### 8.3 Métriques Prometheus

Métriques techniques :
- `http_requests_total` (par route, status, méthode)
- `http_request_duration_seconds`
- `go_goroutines` (santé du runtime)

Métriques métier :
- `lullify_active_streams` — nombre de streams live
- `lullify_active_listeners` — auditeurs connectés
- `lullify_stream_disconnections_total` — déconnexions brutales
- `lullify_audio_bitrate_bytes` — débit audio

### 8.4 Dashboard Grafana

Panels attendus :
- Utilisateurs connectés en temps réel
- Débit de streaming agrégé
- Taux d'erreurs HTTP (4xx / 5xx)
- Temps de réponse API (p50 / p95 / p99)
- Déconnexions brutales vs déconnexions normales
- Goroutines actives

---

## 9. Sécurité

| Mesure | Détail |
|---|---|
| **TLS** | HTTPS obligatoire en production (Let's Encrypt) |
| **JWT** | Access token court (15 min) + refresh token (7j) en cookie HttpOnly |
| **RBAC** | Middleware vérifiant le rôle sur chaque route protégée |
| **Validation** | Validation stricte des inputs (pas d'injection SQL via sqlc) |
| **Rate limiting** | Middleware de rate limiting sur les endpoints publics |
| **CORS** | Origines autorisées explicites |
| **Secrets** | Zéro hardcoding — 100% variables d'environnement |
| **Images Docker** | Distroless ou Alpine — surface d'attaque minimale |
| **Endpoints sensibles** | `/metrics`, `/admin/*` restreints (IP allowlist ou auth) |

---

## 10. Infrastructure & Déploiement

### 10.1 Structure des repositories

```
# Repo 1
lullify-backend/
├── cmd/server/
├── internal/
├── pkg/
├── config/
├── migrations/
├── docker/
│   └── Dockerfile          # Multi-stage build
├── docker-compose.yml      # Dev local complet
├── docker-compose.prod.yml
├── .github/
│   └── workflows/
│       ├── ci.yml          # Tests + lint + build
│       └── cd.yml          # Push image GHCR + deploy
├── .env.example
└── README.md

# Repo 2
lullify-mobile/
├── lib/
├── android/
├── ios/
├── test/
├── Dockerfile              # Build CI Flutter
├── .github/
│   └── workflows/
│       └── ci.yml          # Tests + build APK/IPA
└── README.md
```

### 10.2 Dockerfile Backend (multi-stage)

```dockerfile
# Stage 1 — Build
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o lullify ./cmd/server

# Stage 2 — Runtime
FROM gcr.io/distroless/static-debian12
COPY --from=builder /app/lullify /lullify
EXPOSE 8080
ENTRYPOINT ["/lullify"]
```

### 10.3 Docker Compose (dev)

```yaml
services:
  api:
    build: .
    ports: ["8080:8080"]
    env_file: .env
    depends_on: [postgres, redis, otel-collector]

  postgres:
    image: postgres:16-alpine
    volumes: [postgres_data:/var/lib/postgresql/data]
    environment:
      POSTGRES_DB: lullify
      POSTGRES_USER: lullify
      POSTGRES_PASSWORD: ${DB_PASSWORD}

  redis:
    image: redis:7-alpine

  otel-collector:
    image: otel/opentelemetry-collector-contrib:latest
    volumes: [./config/otel-collector.yml:/etc/otel/config.yml]

  prometheus:
    image: prom/prometheus:latest
    volumes: [./config/prometheus.yml:/etc/prometheus/prometheus.yml]

  grafana:
    image: grafana/grafana:latest
    ports: ["3000:3000"]
    volumes: [grafana_data:/var/lib/grafana]

  loki:
    image: grafana/loki:latest

volumes:
  postgres_data:
  grafana_data:
```

### 10.4 GitHub Actions — CI Backend

```yaml
name: CI
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with: { go-version: '1.23' }
      - run: go test ./... -coverprofile=coverage.out
      - run: go tool cover -func=coverage.out  # doit atteindre 80%

  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: golangci/golangci-lint-action@v6

  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: docker/build-push-action@v6
        with:
          push: false
          tags: ghcr.io/${{ github.repository }}:${{ github.sha }}
```

### 10.5 Variables d'environnement requises

```env
# Server
PORT=8080
ENV=development

# Database
DATABASE_URL=postgres://lullify:password@postgres:5432/lullify?sslmode=disable

# Redis
REDIS_URL=redis://redis:6379

# Auth
JWT_SECRET=changeme
JWT_ACCESS_EXPIRY=15m
JWT_REFRESH_EXPIRY=168h

# OTEL
OTEL_EXPORTER_OTLP_ENDPOINT=http://otel-collector:4318
OTEL_SERVICE_NAME=lullify-backend

# Storage (pour les fichiers audio)
STORAGE_PROVIDER=local         # local | s3
STORAGE_PATH=/data/audio
```

---

## 11. Organisation du projet

### 11.1 Méthodologie

- **Git Flow** : branches `main`, `develop`, `feature/*`, `fix/*`
- **Commits signés** (GPG) et conventionnels (`feat:`, `fix:`, `chore:`, etc.)
- **Pull Requests** obligatoires avant merge sur `develop`
- **Code review** croisée (chaque PR relue par l'autre membre)
- **Issues GitHub** pour le suivi des tâches

### 11.2 Planning indicatif

| Phase | Contenu | Durée estimée |
|---|---|---|
| **Setup** | Repos, CI, Docker Compose, structure du code | S1 |
| **Auth & Users** | JWT, RBAC, endpoints utilisateurs | S2 |
| **Streaming Engine** | Goroutines, HLS, gestion des flux | S3-S4 |
| **Playlists** | CRUD, queue, upload | S3-S4 |
| **Flutter — base** | Navigation, auth, lecteur audio | S3-S4 |
| **Flutter — streaming** | Lecteur live, arrière-plan, diffuseur | S5-S6 |
| **Observabilité** | OTEL, Prometheus, Grafana, Loki | S5-S6 |
| **Tests** | Unitaires (80%), intégration | S5-S7 |
| **Bonus** | Chat WS, offline, K8s | S7-S8 |
| **Soutenance** | Doc finale, polish, démo | S8 |

### 11.3 Architecture Decision Records (ADR)

Chaque choix technique significatif doit être documenté sous `docs/adr/` :

```
docs/adr/
├── 001-choix-langage-backend.md
├── 002-protocol-streaming-hls.md
├── 003-state-management-flutter.md
├── 004-observabilite-otel.md
└── 005-base-de-donnees-postgresql.md
```

Format ADR :
```markdown
# ADR-XXX : [Titre]
**Date** : YYYY-MM-DD
**Statut** : Accepté

## Contexte
...

## Décision
...

## Alternatives considérées
...

## Conséquences
...
```

---

## 12. Livrables attendus

### Commun aux deux repos

- [ ] Repository GitHub public avec commits signés
- [ ] README complet (setup, architecture, how to run)
- [ ] Documentation technique (ce CDC) — FR & EN
- [ ] Application déployée et accessible en production
- [ ] Archive finale (code + doc + livrables)

### Backend (`lullify-backend`)

- [ ] API REST documentée (Swagger/OpenAPI)
- [ ] Moteur de streaming fonctionnel
- [ ] Pipeline CI/CD GitHub Actions
- [ ] Images Docker multi-stage publiées sur GHCR
- [ ] Stack d'observabilité complète (OTEL + Prometheus + Grafana + Loki)
- [ ] Tests unitaires ≥ 80% de couverture
- [ ] ADR documentés

### Mobile (`lullify-mobile`)

- [ ] Application Flutter buildable iOS & Android
- [ ] Lecteur audio avec lecture en arrière-plan
- [ ] Interface diffuseur fonctionnelle
- [ ] Tests widgets et unitaires
- [ ] Pipeline CI Flutter (build APK)

---

## 13. Critères de validation RNCP

Rappel de la règle : **un seul critère "Non acquis" invalide l'ensemble du bloc.**

| Critère | Ce qui sera vérifié |
|---|---|
| **Ce3.1.1** | Dépôt Git avec gestion des branches, historique des commits |
| **Ce3.1.2** | Pipeline CI fonctionnel (build + tests automatiques) |
| **Ce3.1.3** | CI résout les erreurs rapidement, cahier de recette documenté |
| **Ce3.1.4** | Contraintes RGPD respectées (données personnelles) |
| **Ce3.2.1** | Plan de tests couvrant unitaire, intégration, sécurité |
| **Ce3.2.2** | Tests en parallèle du développement |
| **Ce3.2.3** | Tests automatisés identifient et corrigent les bugs |
| **Ce3.2.4** | Tests vérifient le bon fonctionnement selon les attentes utilisateurs |
| **Ce3.3.1** | Outils de surveillance avancés (OTEL, Grafana) opérationnels |
| **Ce3.3.2** | Surveillance oriente la roadmap de développement |
| **Ce3.3.3** | Actions de surveillance guident l'adaptation du code |
| **Ce3.3.4** | Surveillance minimise les vulnérabilités et protège l'intégrité des données |
| **Ce3.4.1** | Mise en production automatisée (CD) |
| **Ce3.4.2** | Distribution automatique sur toutes les plateformes |
| **Ce3.4.3** | Déploiement continu permet le retour utilisateur |
| **Ce3.4.4** | Mises à jour fréquentes de la solution |
| **Ce3.5.1** | Opérations continues permettent ajustements rapides |
| **Ce3.5.2** | Alertes et notifications détectent les anomalies |
| **Ce3.5.3** | Chaîne d'outils intégrée automatise le travail collaboratif |
| **Ce3.5.4** | Performances garanties des applications et environnements |
| **Ce3.6.1** | Documentation technique claire (user stories, schéma DB, sécurité, UML) |
| **Ce3.6.2** | Documentation en français et en anglais (niveau B2) |
| **Ce3.6.3** | Doc adaptée aux différentes versions de la solution |
| **Ce3.6.4** | Documentation inclusive (accessibilité) |
| **Ce3.6.5** | Plan de formation utilisateurs adapté à la diversité du public |

---

*Document rédigé dans le cadre du Projet Semestriel 5A TL — S2 — Bloc 3*  
*École .decode — contact@ecole-decode.fr*  
*Dernière mise à jour : Avril 2026*
