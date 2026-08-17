package fetch

import (
	"context"
	"sync"
)

// Pool runs tasks with at most cap in flight. It is a thin helper over a
// semaphore + WaitGroup, kept tiny and dependency-free so both the page fan-out
// (provider clients) and the owner fan-out (CLI) share one bounded-parallel
// primitive.
//
// Tasks receive the pool's context; when it is canceled Go stops scheduling new
// tasks and Wait returns promptly. Already-running tasks are expected to observe
// ctx themselves (all provider HTTP calls do).
type Pool struct {
	sem chan struct{}
	wg  sync.WaitGroup
	ctx context.Context
}

// NewPool returns a Pool bound to ctx with at most limit tasks in flight. A
// limit <= 0 is treated as 1.
func NewPool(ctx context.Context, limit int) *Pool {
	if limit < 1 {
		limit = 1
	}
	return &Pool{sem: make(chan struct{}, limit), ctx: ctx}
}

// Go schedules task to run when a worker slot is free. If the pool's context is
// already canceled it returns without scheduling. task receives the pool
// context. Go blocks only until a slot is acquired (or the context is canceled),
// not until the task completes; use Wait to await all scheduled tasks.
func (p *Pool) Go(task func(ctx context.Context)) {
	// Fast path: don't schedule once the context is canceled. Checked before the
	// select because select would otherwise pick a ready send case at random.
	if p.ctx.Err() != nil {
		return
	}
	select {
	case <-p.ctx.Done():
		return
	case p.sem <- struct{}{}:
	}
	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		defer func() { <-p.sem }()
		task(p.ctx)
	}()
}

// Wait blocks until every scheduled task has returned.
func (p *Pool) Wait() { p.wg.Wait() }
