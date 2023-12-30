package batcher

import (
	"context"
)

// ConcurrencyControl is an interface for controlling the concurrency of batch operations.
type ConcurrencyControl interface {
	// Acquire acquires a concurrency token.
	Acquire(ctx context.Context) (ConcurrencyToken, error)
}

// ConcurrencyToken is an interface for a token that controls concurrency.
type ConcurrencyToken interface {
	// Release releases the concurrency token.
	Release()
}

// concurrencyToken is an implementation of the ConcurrencyToken interface.
type concurrencyToken struct {
	released chan bool
	release  func()
}

// NewConcurrencyToken creates a new ConcurrencyToken with the provided release function.
func NewConcurrencyToken(release func()) ConcurrencyToken {
	token := &concurrencyToken{
		released: make(chan bool, 1),
		release:  release,
	}
	token.released <- false
	return token
}

// Release releases the concurrency token.
func (c *concurrencyToken) Release() {
	if !<-c.released {
		c.release()
		c.released <- true
	}
}

// unlimitedConcurrencyControl is a ConcurrencyControl that allows unlimited concurrency.
type unlimitedConcurrencyControl struct{}

// NewUnlimitedConcurrencyControl creates a new unlimitedConcurrencyControl.
func NewUnlimitedConcurrencyControl() ConcurrencyControl {
	return &unlimitedConcurrencyControl{}
}

// Acquire acquires a concurrency token.
func (u unlimitedConcurrencyControl) Acquire(ctx context.Context) (ConcurrencyToken, error) {
	return NewConcurrencyToken(func() {}), nil
}

// limitedConcurrencyControl is a ConcurrencyControl that limits concurrency.
type limitedConcurrencyControl struct {
	sem   chan struct{}
	queue chan chan struct{}
}

// NewLimitedConcurrencyControl creates a new limitedConcurrencyControl with the provided concurrency limit and options.
func NewLimitedConcurrencyControl(concurrency int, option ...limitedConcurrencyControlOption) ConcurrencyControl {
	cc := &limitedConcurrencyControl{
		sem:   make(chan struct{}, concurrency),
		queue: make(chan chan struct{}, concurrency*2),
	}

	for _, opt := range option {
		opt(cc)
	}

	return cc
}

// Acquire acquires a concurrency token.
func (l *limitedConcurrencyControl) Acquire(ctx context.Context) (ConcurrencyToken, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		select {
		case l.sem <- struct{}{}:
			return NewConcurrencyToken(l.release), nil
		default:
			releaseChan := make(chan struct{})
			l.queue <- releaseChan
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-releaseChan:
				return NewConcurrencyToken(l.release), nil
			}
		}
	}
}

// release releases a concurrency token.
func (l *limitedConcurrencyControl) release() {
	select {
	case releaseChan := <-l.queue:
		releaseChan <- struct{}{}
	default:
		<-l.sem
	}
}

// LimitedConcurrencyControlOption is a function that configures a limitedConcurrencyControl.
type limitedConcurrencyControlOption func(*limitedConcurrencyControl)

// WithLimitedConcurrencyControlQueueSize returns an option that sets the queue size for a limitedConcurrencyControl.
func WithLimitedConcurrencyControlQueueSize(size int) limitedConcurrencyControlOption {
	return func(l *limitedConcurrencyControl) {
		l.queue = make(chan chan struct{}, size)
	}
}
