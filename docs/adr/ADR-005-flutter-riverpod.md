# ADR-005 — Flutter with Riverpod for State Management

**Date:** 2025-01-01  
**Status:** Accepted

## Context

The mobile app needs robust state management for: audio playback (background + interruptions), authentication with JWT refresh, real-time stream list updates, and broadcaster dashboard.

## Decision

We use Flutter with Riverpod (`flutter_riverpod`) as the state management solution, combined with `just_audio` and `audio_service` for background audio.

## Rationale

- **Riverpod** over BLoC: less boilerplate, compile-time safe providers, easier testing with `ProviderScope` overrides
- **Riverpod** over Provider: no `BuildContext` dependency, works outside widget tree
- `StateNotifierProvider` gives reactive UI updates with sealed class states (Loading/Loaded/Error)
- `audio_service` handles Android/iOS background audio, lock screen controls, and audio focus interruptions natively
- `AuthInterceptor` with `QueuedInterceptorsWrapper` handles JWT refresh transparently across concurrent requests

## Alternatives Considered

- **BLoC**: More verbose, requires more files per feature
- **GetX**: Opinionated, harder to test, mixes concerns
- **Redux**: Over-engineered for this scale

## Consequences

- Riverpod 2.x requires understanding of `ref.watch` vs `ref.read` vs `ref.listen`
- `audio_service` requires Android manifest configuration and iOS background modes
- All providers are lazily initialized — performance benefit but requires careful lifecycle management