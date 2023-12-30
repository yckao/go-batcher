package batcher

import (
	"context"
)

type Thunk[V any] interface {
	Await(context.Context) (V, error)
	Set(context.Context, V) (V, error)
	Error(context.Context, error) (V, error)
}

type thunk[V any] struct {
	pending chan bool
	data    chan *thunkData[V]
}

type thunkData[V any] struct {
	value V
	err   error
}

func NewThunk[V any]() Thunk[V] {
	thunk := &thunk[V]{
		pending: make(chan bool, 1),
		data:    make(chan *thunkData[V], 1),
	}

	thunk.pending <- true

	return thunk
}

func (t *thunk[V]) Await(ctx context.Context) (V, error) {
	select {
	case <-ctx.Done():
		return *new(V), ctx.Err()
	case v := <-t.data:
		t.data <- v
		return v.value, v.err
	}
}

func (t *thunk[V]) Set(ctx context.Context, value V) (V, error) {
	select {
	case <-t.data:
	case <-t.pending:
	}

	t.data <- &thunkData[V]{value: value}

	return t.Await(ctx)
}

func (t *thunk[V]) Error(ctx context.Context, err error) (V, error) {
	select {
	case <-t.data:
	case <-t.pending:
	}

	t.data <- &thunkData[V]{err: err}

	return t.Await(ctx)
}
