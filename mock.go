package batcher

//go:generate mockgen -destination=clock_mock_test.go -package=batcher k8s.io/utils/clock Clock,PassiveClock,Timer
//go:generate mockgen -destination=batcher_mock_test.go -source=batcher.go -package=batcher github.com/yckao/go-batcher Batch,Action
//go:generate mockgen -destination=scheduler_mock_test.go -source=scheduler.go -package=batcher github.com/yckao/go-batcher SchedulerCallback,Scheduler
//go:generate mockgen -destination=concurrency_mock_test.go -source=concurrency.go -package=batcher github.com/yckao/go-batcher ConcurrencyControl
