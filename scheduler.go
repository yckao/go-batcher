package batcher

import (
	"context"
	"time"

	"k8s.io/utils/clock"
)

// Scheduler is an interface for scheduling batch operations.
type Scheduler interface {
	// Schedule schedules a batch operation and calls the provided callback when it's time to dispatch the batch.
	Schedule(ctx context.Context, batch Batch, callback SchedulerCallback)
}

// SchedulerCallback is an interface for callbacks that are called when it's time to dispatch a batch.
type SchedulerCallback interface {
	Call()
}

// schedulerCallback is an implementation of the SchedulerCallback interface.
type schedulerCallback struct {
	callback func()
}

// NewSchedulerCallback creates a new SchedulerCallback with the provided function.
func NewSchedulerCallback(callback func()) SchedulerCallback {
	return &schedulerCallback{
		callback: callback,
	}
}

// Call calls the callback function.
func (c *schedulerCallback) Call() {
	c.callback()
}

// TimeWindowScheduler is a Scheduler that dispatches batches after a certain time window.
type TimeWindowScheduler struct {
	clock      clock.Clock
	timeWindow time.Duration
}

// NewTimeWindowScheduler creates a new TimeWindowScheduler with the provided time window.
func NewTimeWindowScheduler(timeWindow time.Duration) Scheduler {
	return &TimeWindowScheduler{
		clock:      clock.RealClock{},
		timeWindow: timeWindow,
	}
}

// Schedule schedules a batch operation and calls the provided callback when it's time to dispatch the batch.
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

// InstantScheduler is a Scheduler that dispatches batches instantly.
type InstantScheduler struct{}

// NewInstantScheduler creates a new InstantScheduler.
func NewInstantScheduler() Scheduler {
	return &InstantScheduler{}
}

// Schedule calls the provided callback instantly to dispatch the batch.
func (i *InstantScheduler) Schedule(ctx context.Context, batch Batch, callback SchedulerCallback) {
	callback.Call()
}
