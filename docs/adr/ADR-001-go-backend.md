# ADR-001 — Go as Backend Language

**Date:** 2025-01-01  
**Status:** Accepted

## Context

We need a backend capable of handling real-time audio streaming to multiple concurrent listeners with minimal memory footprint and low latency.

## Decision

We chose Go as the backend language.

## Rationale

- Native goroutines and channels enable efficient multiplexing of audio streams without blocking the HTTP server
- Go's memory model and garbage collector are well-suited for long-lived streaming connections
- Strong standard library (net/http, context) covers most needs without heavy dependencies
- Single binary deployment simplifies containerization

## Alternatives Considered

- **Node.js**: Good async I/O but single-threaded event loop limits CPU-bound transcoding
- **Python**: Too slow for real-time audio processing at scale
- **Rust**: Better performance but steeper learning curve and slower development

## Consequences

- Team must learn Go concurrency patterns (goroutines, channels, select)
- Deployment is a single binary, no runtime required