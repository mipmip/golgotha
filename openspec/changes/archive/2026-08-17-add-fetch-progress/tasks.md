## 1. Progress event model

- [x] 1.1 Define the event types (FetchStarted/PageDone/FetchDone/FetchFailed/Canceled/Warning) and an emit seam
- [x] 1.2 Fixed worker cap constant (6); a small bounded-pool helper
- [x] 1.3 Unit tests for the pool and event ordering

## 2. Page-aware provider fetch

- [x] 2.1 Refactor each client to fetch page 1, derive total pages (Link rel=last / X-Total-Pages / X-Total-Count)
- [x] 2.2 Fan out pages 2..N bounded-parallel; merge + dedupe; emit PageDone per page
- [x] 2.3 Sequential fallback + indeterminate progress when no total is exposed
- [x] 2.4 Keep plain `ListRepos` as a wrapper over the event-emitting fetch
- [x] 2.5 Tests against mocked HTTP (multi-page, totals, fallback, cancel)

## 3. Cache safety

- [x] 3.1 Commit an owner only on complete fetch; cancel/failure leaves it unfetched
- [x] 3.2 Tests: complete commits; cancelled/failed does not mutate the owner

## 4. TUI progress + cancel

- [x] 4.1 Stream events via a channel-backed tea.Cmd → progressMsg; keep Update pure
- [x] 4.2 Spinner (bubbles/spinner) until total known, then bar (bubbles/progress) + text line
- [x] 4.3 Esc cancels current fetch, backs out, no partial cache
- [x] 4.4 Update-driven tests (progress advances; cancel path; transient partials)

## 5. CLI progress

- [x] 5.1 sync/refresh consume the event stream → line-oriented per-owner output
- [x] 5.2 Bounded-parallel owner fetch (cap 6); attribute lines per owner
- [x] 5.3 Tests for printed progress + concurrency attribution

## 6. Verify

- [x] 6.1 `gofmt -l .`, `go vet`, `go build`, `go test ./...`, `nix flake check` (coverage gate) all pass
