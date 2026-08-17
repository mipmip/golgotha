// Package fetch defines the progress-event model shared by the repository fetch
// pipeline and its frontends (the TUI and the CLI), plus a small bounded worker
// pool. The event model decouples the work (paginated provider fetches) from
// presentation: a fetch emits typed Event values through an emit seam, and each
// frontend renders them however it likes (a spinner/bar in the TUI, printed
// lines on the CLI). Tests observe the raw event stream.
package fetch

import "fmt"

// WorkerCap is the fixed maximum number of concurrent workers used both for
// fanning out an owner's pages (provider clients) and for fanning out owners
// (the CLI). It is intentionally not configurable for now.
const WorkerCap = 6

// Kind enumerates the progress event kinds.
type Kind int

const (
	// KindStarted marks the beginning of an owner's fetch.
	KindStarted Kind = iota
	// KindPageDone marks the completion of one page of an owner's fetch.
	KindPageDone
	// KindDone marks a successful, complete fetch of an owner.
	KindDone
	// KindFailed marks a fetch that ended in an error.
	KindFailed
	// KindCanceled marks a fetch cut short by context cancellation.
	KindCanceled
	// KindWarning carries a non-fatal warning for an owner.
	KindWarning
)

// String renders the kind for logs and tests.
func (k Kind) String() string {
	switch k {
	case KindStarted:
		return "started"
	case KindPageDone:
		return "page_done"
	case KindDone:
		return "done"
	case KindFailed:
		return "failed"
	case KindCanceled:
		return "canceled"
	case KindWarning:
		return "warning"
	default:
		return fmt.Sprintf("kind(%d)", int(k))
	}
}

// Event is a single progress event emitted by a fetch. The Kind selects which
// of the remaining fields are meaningful:
//
//   - Started:  Provider, Owner
//   - PageDone: Provider, Owner, Page, TotalPages (0 = unknown), ReposSoFar
//   - Done:     Provider, Owner, Count
//   - Failed:   Provider, Owner, Err
//   - Canceled: Provider, Owner
//   - Warning:  Provider, Owner, Msg
type Event struct {
	Kind     Kind
	Provider string
	Owner    string

	// Page is the 1-based page number a PageDone event refers to.
	Page int
	// TotalPages is the known total page count, or 0 when the provider does not
	// expose a total (indeterminate / sequential fallback).
	TotalPages int
	// ReposSoFar is the number of repositories accumulated so far at a PageDone.
	ReposSoFar int
	// Count is the final repository count on a Done event.
	Count int
	// Err carries the failure on a Failed event.
	Err error
	// Msg carries free-form text on a Warning event.
	Msg string
}

// Emit is the seam a fetch uses to publish events. A nil Emit is a valid no-op
// consumer, so fetch code may call emit unconditionally.
type Emit func(Event)

// call invokes e when non-nil, letting callers avoid nil checks at every site.
func (e Emit) call(ev Event) {
	if e != nil {
		e(ev)
	}
}

// Started emits a KindStarted event.
func (e Emit) Started(provider, owner string) {
	e.call(Event{Kind: KindStarted, Provider: provider, Owner: owner})
}

// PageDone emits a KindPageDone event.
func (e Emit) PageDone(provider, owner string, page, totalPages, reposSoFar int) {
	e.call(Event{
		Kind: KindPageDone, Provider: provider, Owner: owner,
		Page: page, TotalPages: totalPages, ReposSoFar: reposSoFar,
	})
}

// Done emits a KindDone event.
func (e Emit) Done(provider, owner string, count int) {
	e.call(Event{Kind: KindDone, Provider: provider, Owner: owner, Count: count})
}

// Failed emits a KindFailed event.
func (e Emit) Failed(provider, owner string, err error) {
	e.call(Event{Kind: KindFailed, Provider: provider, Owner: owner, Err: err})
}

// Canceled emits a KindCanceled event.
func (e Emit) Canceled(provider, owner string) {
	e.call(Event{Kind: KindCanceled, Provider: provider, Owner: owner})
}

// Warning emits a KindWarning event.
func (e Emit) Warning(provider, owner, msg string) {
	e.call(Event{Kind: KindWarning, Provider: provider, Owner: owner, Msg: msg})
}
