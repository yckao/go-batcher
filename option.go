package batcher

// option is a function that configures a Batcher.
type option[REQ any, RES any] func(*batcher[REQ, RES])

// WithMaxBatchSize returns an option that sets the maximum batch size.
func WithMaxBatchSize[REQ any, RES any](maxBatchSize int) option[REQ, RES] {
	return func(l *batcher[REQ, RES]) {
		l.maxBatchSize = maxBatchSize
	}
}

// WithScheduler returns an option that sets the scheduler.
func WithScheduler[REQ any, RES any](scheduler Scheduler) option[REQ, RES] {
	return func(l *batcher[REQ, RES]) {
		l.scheduler = scheduler
	}
}

// WithConcurrencyControl returns an option that sets the concurrency control.
func WithConcurrencyControl[REQ any, RES any](concurrencyControl ConcurrencyControl) option[REQ, RES] {
	return func(l *batcher[REQ, RES]) {
		l.concurrencyControl = concurrencyControl
	}
}

// WithMetricSet returns an option that sets the metric set.
func WithMetricSet[REQ any, RES any](metrics *MetricSet) option[REQ, RES] {
	return func(l *batcher[REQ, RES]) {
		l.metrics = metrics
	}
}
