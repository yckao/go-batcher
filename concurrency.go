package batcher

import (
	"context"
)

type ConcurrencyControl interface {
	Acquire(ctx context.Context) (ConcurrencyToken, error)
}

type ConcurrencyToken interface {
	Release()
}

type concurrencyToken struct {
	released chan bool
	release  func()
}

func NewConcurrencyToken(release func()) ConcurrencyToken {
	token := &concurrencyToken{
		released: make(chan bool, 1),
		release:  release,
	}
	token.released <- false
	return token
}

func (c *concurrencyToken) Release() {
	if !<-c.released {
		c.release()
		c.released <- true
	}
}

type UnlimitedConcurrencyControl struct{}

func NewUnlimitedConcurrencyControl() ConcurrencyControl {
	return &UnlimitedConcurrencyControl{}
}

func (u UnlimitedConcurrencyControl) Acquire(ctx context.Context) (ConcurrencyToken, error) {
	return NewConcurrencyToken(func() {}), nil
}

type limitedConcurrencyControl struct {
	sem   chan struct{}
	queue chan chan struct{}
}

func NewLimitedConcurrencyControl(concurrency int, option ...LimitedConcurrencyControlOption) ConcurrencyControl {
	cc := &limitedConcurrencyControl{
		sem:   make(chan struct{}, concurrency),
		queue: make(chan chan struct{}, concurrency*2),
	}

	for _, opt := range option {
		opt(cc)
	}

	return cc
}

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

func (l *limitedConcurrencyControl) release() {
	select {
	case releaseChan := <-l.queue:
		releaseChan <- struct{}{}
	default:
		<-l.sem
	}
}

type LimitedConcurrencyControlOption func(*limitedConcurrencyControl)

func WithLimitedConcurrencyControlQueueSize(size int) LimitedConcurrencyControlOption {
	return func(l *limitedConcurrencyControl) {
		l.queue = make(chan chan struct{}, size)
	}
}
