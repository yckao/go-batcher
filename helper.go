package batcher

import (
	"context"
	"time"
)

func NewTimeWindowScheduler(t time.Duration) BatchScheduleFn {
	return func(ctx context.Context, batch Batch, callback func()) {
		timer := time.NewTimer(t)
		defer timer.Stop()
		select {
		case <-ctx.Done():
			return
		case <-batch.Dispatch():
			callback()
		case <-batch.Full():
			callback()
		case <-timer.C:
			callback()
		}
	}
}

type UnlimitedConcurrencyControl struct{}

func (u UnlimitedConcurrencyControl) Acquire(ctx context.Context) (func(), error) {
	return func() {}, nil
}

type defaultConcurrencyControl struct {
	sem   chan struct{}
	queue chan chan struct{}
}

func (d *defaultConcurrencyControl) Acquire(ctx context.Context) (func(), error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		select {
		case d.sem <- struct{}{}:
			return func() { <-d.sem }, nil
		default:
			releaseChan := make(chan struct{})
			d.queue <- releaseChan
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-releaseChan:
				return func() {
					<-d.sem
					d.releaseNext()
				}, nil
			}
		}
	}
}

func (c *defaultConcurrencyControl) releaseNext() {
	select {
	case releaseChan := <-c.queue:
		releaseChan <- struct{}{}
	default:
	}
}

func NewDefaultConcurrencyControl(concurrency int) BatchConcurrencyControl {
	return &defaultConcurrencyControl{
		sem:   make(chan struct{}, concurrency),
		queue: make(chan chan struct{}, concurrency),
	}
}
