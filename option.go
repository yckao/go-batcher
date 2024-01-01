package batcher

type batcherConfig struct {
	maxBatchSize       int
	scheduler          Scheduler
	concurrencyControl ConcurrencyControl
	metrics            *MetricSet
}

// option is a function that configures a Batcher.
type option func(conf *batcherConfig)

// WithMaxBatchSize returns an option that sets the maximum batch size.
func WithMaxBatchSize(maxBatchSize int) option {
	return func(conf *batcherConfig) {
		conf.maxBatchSize = maxBatchSize
	}
}

// WithScheduler returns an option that sets the scheduler.
func WithScheduler(scheduler Scheduler) option {
	return func(conf *batcherConfig) {
		conf.scheduler = scheduler
	}
}

// WithConcurrencyControl returns an option that sets the concurrency control.
func WithConcurrencyControl(concurrencyControl ConcurrencyControl) option {
	return func(conf *batcherConfig) {
		conf.concurrencyControl = concurrencyControl
	}
}

// WithMetricSet returns an option that sets the metric set.
func WithMetricSet(metrics *MetricSet) option {
	return func(conf *batcherConfig) {
		conf.metrics = metrics
	}
}
