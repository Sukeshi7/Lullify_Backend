# ADR-003 — HLS for Audio Streaming

**Date:** 2025-01-01  
**Status:** Accepted

## Context

We need to stream audio from a broadcaster to multiple listeners simultaneously. The protocol must work across mobile platforms (iOS, Android) and support background playback.

## Decision

We use HTTP Live Streaming (HLS) with `.m3u8` playlists and `.ts` segments served over standard HTTP.

## Rationale

- HLS is natively supported by `just_audio` on Flutter without additional plugins
- Works over standard HTTP — no WebSocket or custom protocol needed
- Sliding window of segments (6 × 2s = 12s buffer) limits disk usage
- iOS requires HLS for background audio — no viable alternative
- Segments are served as static files, enabling CDN caching at scale

## Alternatives Considered

- **WebRTC**: Lower latency but complex signaling server and poor mobile background support
- **Icecast/Shoutcast**: Proven for radio but requires a separate server process
- **WebSocket raw audio**: No buffering, no seek, fragile on network changes

## Consequences

- ~2-4 second latency inherent to HLS segmentation
- Disk I/O for segment writing (mitigated by `/tmp` and segment cleanup)
- Listeners can join mid-stream seamlessly via playlist sliding window