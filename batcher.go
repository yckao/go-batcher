package batcher

import (
	"context"
	"sync"
	"time"
)

type Batcher[REQ any, RES any] interface {
	Do(context.Context, REQ) Thunk[RES]
	Dispatch()
	Shutdown() error
}

type Response[RES any] struct {
	Response RES
	Error    error
}

type Batch interface {
	Full() <-chan struct{}
	Dispatch() <-chan struct{}
}

type BatchDoFn[REQ any, RES any] func(context.Context, []REQ) []Response[RES]
type Scheduler interface {
	Schedule(ctx context.Context, batch Batch, callback SchedulerCallback)
}

type BatchConcurrencyControl interface {
	Acquire(ctx context.Context) (func(), error)
}

func New[REQ any, RES any](ctx context.Context, doFn BatchDoFn[REQ, RES], options ...option[REQ, RES]) Batcher[REQ, RES] {
	b := &batcher[REQ, RES]{
		ctx:                ctx,
		batches:            make(chan []*batch[REQ, RES], 1),
		doFn:               doFn,
		scheduler:          NewTimeWindowScheduler(2 * time.Second),
		maxBatchSize:       100,
		concurrencyControl: &UnlimitedConcurrencyControl{},
	}

	b.batches <- []*batch[REQ, RES]{}

	for _, option := range options {
		option(b)
	}

	return b
}

type batcher[REQ any, RES any] struct {
	ctx                context.Context
	closed             bool
	wg                 sync.WaitGroup
	batches            chan []*batch[REQ, RES]
	doFn               BatchDoFn[REQ, RES]
	scheduler          Scheduler
	concurrencyControl BatchConcurrencyControl
	maxBatchSize       int
}

type batch[REQ any, RES any] struct {
	full     chan struct{}
	dispatch chan struct{}
	requests []REQ
	thunks   []Thunk[RES]
}

func (b *batch[K, V]) Full() <-chan struct{} {
	return b.full
}

func (b *batch[K, V]) Dispatch() <-chan struct{} {
	return b.dispatch
}

func (b *batcher[REQ, RES]) Do(ctx context.Context, request REQ) Thunk[RES] {
	thunk := NewThunk[RES]()
	if b.closed {
		thunk.Error(ctx, context.Canceled)
		return thunk
	}

	batches := <-b.batches
	if len(batches) == 0 || len(batches[len(batches)-1].requests) >= b.maxBatchSize {
		bat := &batch[REQ, RES]{
			full:     make(chan struct{}),
			dispatch: make(chan struct{}),
			requests: []REQ{},
			thunks:   []Thunk[RES]{},
		}

		batches = append(batches, bat)
		b.wg.Add(1)

		go b.scheduler.Schedule(b.ctx, bat, NewSchedulerCallback(b.dispatch))
	}

	bat := batches[len(batches)-1]
	bat.requests = append(bat.requests, request)
	bat.thunks = append(bat.thunks, thunk)

	if len(batches) != 0 && len(batches[len(batches)-1].requests) >= b.maxBatchSize {
		close(batches[len(batches)-1].full)
	}

	b.batches <- batches

	return thunk
}

func (b *batcher[REQ, RES]) Shutdown() error {
	b.closed = true
	b.Dispatch()
	b.wg.Wait()
	return nil
}

func (b *batcher[REQ, RES]) Dispatch() {
	batches := <-b.batches
	for _, batch := range batches {
		select {
		case <-batch.full:
		case <-batch.dispatch:
		default:
			close(batch.dispatch)
		}
	}
	b.batches <- batches
}

func (b *batcher[REQ, RES]) dispatch() {
	ctx := b.ctx
	batches := <-b.batches

	if len(batches) == 0 {
		b.batches <- batches
		return
	}
	batch := batches[0]

	b.batches <- batches[1:]
	release, err := b.concurrencyControl.Acquire(ctx)
	if err != nil {
		for _, thunk := range batch.thunks {
			thunk.Error(ctx, err)
		}
	}

	results := b.doFn(ctx, batch.requests)

	release()

	for index, res := range results {
		if res.Error != nil {
			batch.thunks[index].Error(ctx, res.Error)
		} else {
			batch.thunks[index].Set(ctx, res.Response)
		}
	}

	b.wg.Done()
}
