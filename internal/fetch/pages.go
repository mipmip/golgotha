package fetch

import (
	"context"
	"sort"
	"sync"
)

// Page is the result of fetching one page of an owner's repositories. Item is
// the provider-agnostic element type (the caller supplies provider.Repo).
type Page[T any] struct {
	// Items are the repositories on this page.
	Items []T
	// TotalPages is the total page count derived from the provider's paging
	// metadata, or 0 when the provider does not expose a total. It only needs to
	// be set on the first page; later pages may leave it 0.
	TotalPages int
}

// PageFunc fetches page `page` (1-based) of an owner's repositories. It must
// honor ctx for cancellation. The returned Page carries the items and, for
// page 1, the total page count when known.
type PageFunc[T any] func(ctx context.Context, page int) (Page[T], error)

// KeyFunc returns the dedupe key for an item (e.g. owner/name).
type KeyFunc[T any] func(T) string

// Pages fetches all pages for one owner and returns the merged, deduped items,
// emitting progress events through emit:
//
//   - Started once at the beginning.
//   - PageDone per page as it completes (with TotalPages when known and the
//     running ReposSoFar count).
//   - Done with the final count on success, Failed on error, or Canceled when
//     ctx is canceled.
//
// It fetches page 1 first to learn the total. When the total is known and > 1
// it fans out pages 2..N with at most WorkerCap in flight (bounded via Pool).
// When the total is unknown it falls back to sequential pagination, stopping on
// the first short/empty page. Results are merged in page order and deduped by
// key, preserving first-seen order.
func Pages[T any](
	ctx context.Context,
	emit Emit,
	provider, owner string,
	perPage int,
	fetchPage PageFunc[T],
	key KeyFunc[T],
) ([]T, error) {
	emit.Started(provider, owner)

	// Page 1 (sequential) to learn the total.
	first, err := fetchPage(ctx, 1)
	if err != nil {
		if ctx.Err() != nil {
			emit.Canceled(provider, owner)
			return nil, ctx.Err()
		}
		emit.Failed(provider, owner, err)
		return nil, err
	}

	total := first.TotalPages

	// byPage collects each page's items keyed by page number so the final merge
	// is deterministic regardless of completion order.
	var mu sync.Mutex
	byPage := map[int][]T{1: first.Items}
	reposSoFar := len(first.Items)
	emit.PageDone(provider, owner, 1, total, reposSoFar)

	if total > 1 {
		// Determinate: fan out pages 2..N bounded-parallel.
		pool := NewPool(ctx, WorkerCap)
		var (
			ferr   error
			ferrMu sync.Mutex
		)
		for page := 2; page <= total; page++ {
			page := page
			pool.Go(func(ctx context.Context) {
				pg, perr := fetchPage(ctx, page)
				if perr != nil {
					ferrMu.Lock()
					if ferr == nil {
						ferr = perr
					}
					ferrMu.Unlock()
					return
				}
				mu.Lock()
				byPage[page] = pg.Items
				reposSoFar += len(pg.Items)
				soFar := reposSoFar
				mu.Unlock()
				emit.PageDone(provider, owner, page, total, soFar)
			})
		}
		pool.Wait()

		if ctx.Err() != nil {
			emit.Canceled(provider, owner)
			return nil, ctx.Err()
		}
		if ferr != nil {
			emit.Failed(provider, owner, ferr)
			return nil, ferr
		}
	} else if total == 0 && len(first.Items) >= perPage {
		// Indeterminate fallback: sequential pagination until a short/empty page.
		// A page shorter than perPage (or empty) is the last one. Page 1 was
		// already a full page, so there may be more.
		for page := 2; ; page++ {
			if ctx.Err() != nil {
				emit.Canceled(provider, owner)
				return nil, ctx.Err()
			}
			pg, perr := fetchPage(ctx, page)
			if perr != nil {
				if ctx.Err() != nil {
					emit.Canceled(provider, owner)
					return nil, ctx.Err()
				}
				emit.Failed(provider, owner, perr)
				return nil, perr
			}
			byPage[page] = pg.Items
			reposSoFar += len(pg.Items)
			emit.PageDone(provider, owner, page, 0, reposSoFar)
			if len(pg.Items) < perPage {
				break
			}
		}
	}

	merged := mergeDedupe(byPage, key)
	emit.Done(provider, owner, len(merged))
	return merged, nil
}

// mergeDedupe flattens byPage in ascending page order and drops later duplicates
// by key, preserving first-seen order.
func mergeDedupe[T any](byPage map[int][]T, key KeyFunc[T]) []T {
	pages := make([]int, 0, len(byPage))
	for p := range byPage {
		pages = append(pages, p)
	}
	sort.Ints(pages)

	seen := make(map[string]struct{})
	out := make([]T, 0)
	for _, p := range pages {
		for _, it := range byPage[p] {
			k := key(it)
			if _, dup := seen[k]; dup {
				continue
			}
			seen[k] = struct{}{}
			out = append(out, it)
		}
	}
	return out
}
