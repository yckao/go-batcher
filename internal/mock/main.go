package mock

//go:generate mockgen -destination=clock_mock.go -package=mock k8s.io/utils/clock Clock,PassiveClock,Timer
//go:generate mockgen -destination=batcher_mock.go -package=mock github.com/yckao/go-batcher Callback,Batch
