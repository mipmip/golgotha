// Package cache persists per-provider repository metadata as JSON under
// ~/.cache/skull2/<provider>.json, acting as the single source of truth for
// the TUI and sync.
//
// Implemented in milestone "02 Provider clients & cache", epic "02d JSON cache
// & refresh". See BRIEFING.md section 7.
package cache
