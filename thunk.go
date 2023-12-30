package batcher

import (
	"context"
	"reflect"
)

// Thunk is a javascript promise-like interface that can be awaited.
// It has setter, getter and await to interact with data.
// And pending fulfilled and rejected to check the status of the thunk.
// Just like a promise.
type Thunk[V any] interface {
	// Await will block current goroutine until the value or error is set.
	Await(context.Context) (V, error)
	// Set sets the result of the computation.
	Set(context.Context, V) (V, error)
	// Error sets an error for the computation.
	Error(context.Context, error) (V, error)
	// Pending checks if the computation is still pending.
	Pending() bool
	// Fulfilled checks if the computation has a result.
	Fulfilled() bool
	// Rejected checks if the computation has an error.
	Rejected() bool
}

// thunk is an implementation of the Thunk interface.
type thunk[V any] struct {
	pending chan bool
	data    chan *thunkData[V]
}

// thunkData holds the result or error of a computation.
type thunkData[V any] struct {
	value V
	err   error
}

// NewThunk creates a new Thunk with generic type V represent any kind of value.
func NewThunk[V any]() Thunk[V] {
	thunk := &thunk[V]{
		pending: make(chan bool, 1),
		data:    make(chan *thunkData[V], 1),
	}

	thunk.pending <- true

	return thunk
}

// Await implements the Thunk interface.
func (t *thunk[V]) Await(ctx context.Context) (V, error) {
	select {
	case <-ctx.Done():
		return *new(V), ctx.Err()
	case v := <-t.data:
		t.data <- v
		return v.value, v.err
	}
}

// Set implements the Thunk interface.
func (t *thunk[V]) Set(ctx context.Context, value V) (V, error) {
	select {
	case <-t.data:
	case <-t.pending:
	}

	t.data <- &thunkData[V]{value: value}

	return t.Await(ctx)
}

// Error implements the Thunk interface.
func (t *thunk[V]) Error(ctx context.Context, err error) (V, error) {
	select {
	case <-t.data:
	case <-t.pending:
	}

	t.data <- &thunkData[V]{err: err}

	return t.Await(ctx)
}

// Pending implements the Thunk interface.
func (t *thunk[V]) Pending() bool {
	select {
	case <-t.pending:
		t.pending <- true
		return true
	default:
		return false
	}
}

// Fulfilled implements the Thunk interface.
func (t *thunk[V]) Fulfilled() bool {
	select {
	case data := <-t.data:
		t.data <- data
		return !reflect.ValueOf(data.value).IsZero()
	default:
		return false
	}
}

// Rejected implements the Thunk interface.
func (t *thunk[V]) Rejected() bool {
	select {
	case data := <-t.data:
		t.data <- data
		return data.err != nil
	default:
		return false
	}
}
