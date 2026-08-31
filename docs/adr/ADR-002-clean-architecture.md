# ADR-002 — Clean Architecture with DDD

**Date:** 2025-01-01  
**Status:** Accepted

## Context

The codebase needs to be testable at 80%+ unit coverage without requiring a live database or external services.

## Decision

We adopted Clean Architecture with Domain-Driven Design, separating the codebase into three layers: domain, application, and infrastructure.

## Rationale

- Domain layer (entities, errors, repository interfaces) has zero external dependencies — fully unit testable
- Application layer (use cases) depends only on interfaces, enabling mock-based testing without DB
- Infrastructure layer (postgres, redis, HTTP handlers) is the only layer touching external services
- This separation allows 82%+ unit coverage on business logic while infrastructure is covered by integration tests

## Alternatives Considered

- **MVC**: Simpler but tightly couples business logic to HTTP handlers, making unit testing harder
- **Hexagonal Architecture**: Equivalent approach with different naming conventions

## Consequences

- More boilerplate (interfaces, separate packages per layer)
- Clear boundaries make onboarding and code review easier
- Adding a new feature requires touching domain → application → infrastructure in sequence