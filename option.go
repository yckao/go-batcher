package batcher

type option[REQ any, RES any] func(*batcher[REQ, RES])

func WithMaxBatchSize[REQ any, RES any](maxBatchSize int) option[REQ, RES] {
	return func(l *batcher[REQ, RES]) {
		l.maxBatchSize = maxBatchSize
	}
}

func WithScheduler[REQ any, RES any](scheduler Scheduler) option[REQ, RES] {
	return func(l *batcher[REQ, RES]) {
		l.scheduler = scheduler
	}
}

func WithConcurrencyControl[REQ any, RES any](concurrencyControl ConcurrencyControl) option[REQ, RES] {
	return func(l *batcher[REQ, RES]) {
		l.concurrencyControl = concurrencyControl
	}
}
