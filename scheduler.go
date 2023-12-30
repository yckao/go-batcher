package batcher

import (
	"context"
	"time"

	"k8s.io/utils/clock"
)

type SchedulerCallback interface {
	Call()
}

type schedulerCallback struct {
	callback func()
}

func NewSchedulerCallback(callback func()) SchedulerCallback {
	return &schedulerCallback{
		callback: callback,
	}
}

func (c *schedulerCallback) Call() {
	c.callback()
}

type TimeWindowScheduler struct {
	clock      clock.Clock
	timeWindow time.Duration
}

func NewTimeWindowScheduler(timeWindow time.Duration) Scheduler {
	return &TimeWindowScheduler{
		clock:      clock.RealClock{},
		timeWindow: timeWindow,
	}
}

func (t *TimeWindowScheduler) Schedule(ctx context.Context, batch Batch, callback SchedulerCallback) {
	timer := t.clock.NewTimer(t.timeWindow)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return
	case <-batch.Dispatch():
		callback.Call()
	case <-batch.Full():
		callback.Call()
	case <-timer.C():
		callback.Call()
	}
}

type InstantScheduler struct{}

func NewInstantScheduler() Scheduler {
	return &InstantScheduler{}
}

func (i *InstantScheduler) Schedule(ctx context.Context, batch Batch, callback SchedulerCallback) {
	callback.Call()
}
