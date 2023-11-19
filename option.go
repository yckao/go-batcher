package batcher

type option[REQ any, RES any] func(*batcher[REQ, RES])

func WithMaxBatchSize[REQ any, RES any](maxBatchSize int) option[REQ, RES] {
	return func(l *batcher[REQ, RES]) {
		l.maxBatchSize = maxBatchSize
	}
}

func WithScheduleFn[REQ any, RES any](scheduleFn BatchScheduleFn) option[REQ, RES] {
	return func(l *batcher[REQ, RES]) {
		l.scheduleFn = scheduleFn
	}
}

func WithConcurrencyControl[REQ any, RES any](concurrencyControl BatchConcurrencyControl) option[REQ, RES] {
	return func(l *batcher[REQ, RES]) {
		l.concurrencyControl = concurrencyControl
	}
}
