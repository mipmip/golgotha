package main

import (
	"context"
	"fmt"
	"io"
	"sync"

	"github.com/mipmip/golgotha/internal/config"
	"github.com/mipmip/golgotha/internal/fetch"
	"github.com/mipmip/golgotha/internal/provider"
)

// ownerResult is one owner's fetch outcome from the bounded-parallel sweep.
type ownerResult struct {
	Owner string
	Repos []provider.Repo
	Err   error
}

// cliProgressPrinter serializes fetch progress into line-oriented, cron-friendly
// output. Each event becomes one complete line attributed to its owner, guarded
// by a mutex so concurrent owners never interleave mid-line.
type cliProgressPrinter struct {
	mu sync.Mutex
	w  io.Writer
}

// emitter returns a fetch.Emit that prints one line per meaningful event for the
// given provider/owner.
func (p *cliProgressPrinter) emitter(providerName, owner string) fetch.Emit {
	return func(ev fetch.Event) {
		p.mu.Lock()
		defer p.mu.Unlock()
		who := displayOwner(owner)
		switch ev.Kind {
		case fetch.KindStarted:
			fmt.Fprintf(p.w, "%s: %s: fetching...\n", providerName, who)
		case fetch.KindDone:
			fmt.Fprintf(p.w, "%s: %s: fetched %d repos\n", providerName, who, ev.Count)
		case fetch.KindFailed:
			fmt.Fprintf(p.w, "%s: %s: fetch failed: %v\n", providerName, who, ev.Err)
		case fetch.KindCanceled:
			fmt.Fprintf(p.w, "%s: %s: fetch canceled\n", providerName, who)
		case fetch.KindWarning:
			fmt.Fprintf(p.w, "%s: %s: warning: %s\n", providerName, who, ev.Msg)
		}
		// PageDone is intentionally not printed on the CLI to stay cron-friendly;
		// the TUI shows per-page progress instead.
	}
}

// fetchOwnersProgress fetches each owner's repositories bounded-parallel (at most
// fetch.WorkerCap owners in flight), printing per-owner progress through printer.
// It prefers the client's OwnerFetcher (page-aware, event-emitting); when the
// client does not implement it, it falls back to a single ListRepos per owner
// framed by start/done events. Results are returned per owner (including errors)
// so the caller can commit only the owners that fetched completely.
func fetchOwnersProgress(
	ctx context.Context,
	client provider.Provider,
	p *config.Provider,
	owners []string,
	printer *cliProgressPrinter,
) []ownerResult {
	of, hasFetcher := client.(provider.OwnerFetcher)

	results := make([]ownerResult, len(owners))
	pool := fetch.NewPool(ctx, fetch.WorkerCap)
	for i, owner := range owners {
		i, owner := i, owner
		pool.Go(func(ctx context.Context) {
			emit := printer.emitter(p.Name, owner)
			if hasFetcher {
				repos, err := of.FetchOwner(ctx, emit, owner)
				results[i] = ownerResult{Owner: owner, Repos: repos, Err: err}
				return
			}
			// Fallback: no page-aware fetch; frame a single ListRepos with events.
			emit.Started(p.Name, owner)
			var reqOwners []string
			if owner != config.SelfOwner {
				reqOwners = []string{owner}
			}
			repos, err := client.ListRepos(ctx, reqOwners)
			if err != nil {
				emit.Failed(p.Name, owner, err)
				results[i] = ownerResult{Owner: owner, Err: err}
				return
			}
			emit.Done(p.Name, owner, len(repos))
			results[i] = ownerResult{Owner: owner, Repos: repos}
		})
	}
	pool.Wait()
	return results
}
