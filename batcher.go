package batcher

import (
	"context"
	"sync"
	"time"
)

// Batcher is an interface for a batcher that batches operations.
type Batcher[REQ any, RES any] interface {
	// Do adds a request to the batcher and returns a Thunk that will be filled with the result.
	Do(context.Context, REQ) Thunk[RES]
	// Shutdown will dispatch pending batchers and waits for all operations to complete.
	Shutdown() error
}

// Response is a struct that holds a response and an error.
type Response[RES any] struct {
	Response RES
	Error    error
}

// Batch is an interface for a batch of operations.
type Batch interface {
	// Full returns a channel that is closed when the batch is full.
	Full() <-chan struct{}
	// Dispatch returns a channel that is closed when the batch is dispatched.
	Dispatch() <-chan struct{}
}

// Action is an interface for an action that can be performed on a batch of requests.
type Action[REQ any, RES any] interface {
	// Perform performs the action on the batch of requests and returns a slice of Responses.
	Perform(context.Context, []REQ) []Response[RES]
}

// action is a concrete implementation of the Action interface.
type action[REQ any, RES any] struct {
	fn func(context.Context, []REQ) []Response[RES]
}

// NewAction creates a new Action with the provided function.
func NewAction[REQ any, RES any](fn func(context.Context, []REQ) []Response[RES]) Action[REQ, RES] {
	return &action[REQ, RES]{
		fn: fn,
	}
}

// Perform performs the action on the batch of requests and returns a slice of Responses.
func (b *action[REQ, RES]) Perform(ctx context.Context, requests []REQ) []Response[RES] {
	return b.fn(ctx, requests)
}

// New creates a new Batcher with the provided context, action, and options.
func New[REQ any, RES any](ctx context.Context, action Action[REQ, RES], options ...option[REQ, RES]) Batcher[REQ, RES] {
	b := &batcher[REQ, RES]{
		ctx:                ctx,
		closed:             make(chan bool),
		batches:            make(chan []*batch[REQ, RES], 1),
		action:             action,
		scheduler:          NewTimeWindowScheduler(2 * time.Second),
		maxBatchSize:       100,
		concurrencyControl: NewUnlimitedConcurrencyControl(),
	}

	b.batches <- []*batch[REQ, RES]{}

	for _, option := range options {
		option(b)
	}

	if b.metrics == nil {
		b.metrics = NewMetricSet("go", "batcher", nil)
	}

	return b
}

// batcher is a concrete implementation of the Batcher interface.
type batcher[REQ any, RES any] struct {
	ctx     context.Context
	closed  chan bool
	wg      sync.WaitGroup
	metrics *MetricSet

	batches            chan []*batch[REQ, RES]
	action             Action[REQ, RES]
	scheduler          Scheduler
	concurrencyControl ConcurrencyControl
	maxBatchSize       int
}

// batch is a concrete implementation of the Batch interface.
type batch[REQ any, RES any] struct {
	full      chan struct{}
	dispatch  chan struct{}
	requests  []REQ
	thunks    []Thunk[RES]
	createdAt time.Time
}

// Full returns a channel that is closed when the batch is full.
func (b *batch[K, V]) Full() <-chan struct{} {
	return b.full
}

// Dispatch returns a channel that is closed when the batch is dispatched.
func (b *batch[K, V]) Dispatch() <-chan struct{} {
	return b.dispatch
}

// Do adds a request to the batcher and returns a Thunk that will be filled with the result.
func (b *batcher[REQ, RES]) Do(ctx context.Context, request REQ) Thunk[RES] {
	b.metrics.DoActionCounter.Inc()

	b.metrics.ThunkCreatedCounter.Inc()
	thunk := NewThunk[RES]()

	batches := <-b.batches

	select {
	case <-b.closed:
		b.metrics.ThunkErrorCounter.Inc()
		thunk.Error(ctx, context.Canceled)
		b.batches <- batches
		return thunk
	default:
	}

	if len(batches) == 0 || len(batches[len(batches)-1].requests) >= b.maxBatchSize {
		b.metrics.BatchCreatedCounter.Inc()
		bat := &batch[REQ, RES]{
			full:      make(chan struct{}),
			dispatch:  make(chan struct{}),
			requests:  []REQ{},
			thunks:    []Thunk[RES]{},
			createdAt: time.Now(),
		}

		batches = append(batches, bat)
		b.wg.Add(1)

		b.metrics.SchedulerScheduleCounter.Inc()
		go b.scheduler.Schedule(b.ctx, bat, NewSchedulerCallback(b.dispatch))
	}

	bat := batches[len(batches)-1]
	bat.requests = append(bat.requests, request)
	bat.thunks = append(bat.thunks, thunk)

	if len(batches) != 0 && len(batches[len(batches)-1].requests) >= b.maxBatchSize {
		b.metrics.BatchFullCounter.Inc()
		close(batches[len(batches)-1].full)
	}

	b.batches <- batches

	return thunk
}

// Shutdown will dispatch pending batchers and waits for all operations to complete.
func (b *batcher[REQ, RES]) Shutdown() error {
	close(b.closed)
	b.flushall()
	b.wg.Wait()
	return nil
}

// flushall flushes all batches in the batcher.
func (b *batcher[REQ, RES]) flushall() {
	batches := <-b.batches
	for _, batch := range batches {
		close(batch.dispatch)
	}
	b.batches <- batches
}

// dispatch dispatches the first batch in the batcher.
func (b *batcher[REQ, RES]) dispatch() {
	b.metrics.SchedulerCallbackCounter.Inc()
	ctx := b.ctx
	batches := <-b.batches

	if len(batches) == 0 {
		b.batches <- batches
		return
	}
	batch := batches[0]

	b.metrics.BatchStartedCounter.Inc()
	b.batches <- batches[1:]

	b.metrics.BatchSizeHistogram.Observe(float64(len(batch.requests)))
	b.metrics.CouncurrencyControlAcquireCounter.Inc()
	token, err := b.concurrencyControl.Acquire(ctx)

	if err != nil {
		b.metrics.ConcurrencyControlErrorCounter.Inc()
		for _, thunk := range batch.thunks {
			b.metrics.ThunkErrorCounter.Inc()
			thunk.Error(ctx, err)
		}

		b.metrics.BatchDoneCounter.Inc()
		b.metrics.BatchLifetimeHistogram.Observe(time.Since(batch.createdAt).Seconds())
		b.wg.Done()
		return
	}

	b.metrics.ConcurrencyControlTokenCounter.Inc()
	b.metrics.BatchActionPerformCounter.Inc()
	results := b.action.Perform(ctx, batch.requests)

	b.metrics.ConcurrencyControlReleaseCounter.Inc()
	token.Release()

	for index, res := range results {
		if res.Error != nil {
			b.metrics.ThunkErrorCounter.Inc()
			batch.thunks[index].Error(ctx, res.Error)
		} else {
			b.metrics.ThunkSuccessCounter.Inc()
			batch.thunks[index].Set(ctx, res.Response)
		}
	}

	b.metrics.BatchDoneCounter.Inc()
	b.metrics.BatchLifetimeHistogram.Observe(time.Since(batch.createdAt).Seconds())
	b.wg.Done()
}
