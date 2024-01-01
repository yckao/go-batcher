package batcher

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gleak"
	"github.com/prometheus/client_golang/prometheus"
	gomock "go.uber.org/mock/gomock"
)

var _ = Describe("Action", func() {
	var (
		a Action[string, string]

		called bool
	)

	BeforeEach(func() {
		a = NewAction[string, string](func(ctx context.Context, requests []string) []Response[string] {
			called = true
			return nil
		})
	})

	JustBeforeEach(func() {
		a.Perform(context.Background(), nil)
	})

	It("should call callback", func() {
		Expect(called).To(BeTrue())
	})
})

var _ = Describe("Batcher", func() {
	BeforeEach(func() {
		goods := Goroutines()
		DeferCleanup(func() {
			Eventually(Goroutines).ShouldNot(HaveLeaked(goods))
		})
	})

	var (
		ctx        context.Context
		cancelFunc context.CancelFunc
	)

	BeforeEach(func() {
		ctx, cancelFunc = context.WithCancel(context.TODO())
	})

	AfterEach(func() {
		cancelFunc()
	})

	Describe("Batcher[string, string]", func() {
		var (
			wg      *sync.WaitGroup
			action  *MockAction[string, string]
			b       *batcher[string, string]
			options []option
		)

		BeforeEach(func() {
			wg = &sync.WaitGroup{}
			action = NewMockAction[string, string](ctrl)
		})

		JustBeforeEach(func() {
			b = New[string, string](ctx, action, options...).(*batcher[string, string])
		})

		Describe("can batch do request", func() {
			var (
				actionCount int
				metrics     *MetricSet
				requests    []string
				responses   []Response[string]
			)

			BeforeEach(func() {
				actionCount = gofakeit.Number(3, 5)
				metrics = NewMetricSet("go", "batcher", nil)
				metrics.Register(prometheus.DefaultRegisterer)
				options = append(options, WithMaxBatchSize(actionCount), WithMetricSet(metrics))
				requests = make([]string, actionCount)
				responses = make([]Response[string], actionCount)

				for i := 0; i < actionCount; i++ {
					requests[i] = fmt.Sprintf("req: #%d", i)
					responses[i] = Response[string]{
						Response: fmt.Sprintf("res: #%d", i),
					}
				}

				action.EXPECT().Perform(ctx, requests).Times(1).Return(responses)
			})

			It("should batch do request", func() {
				wg.Add(actionCount)
				for i := 0; i < actionCount; i++ {
					go func(i int) {
						defer GinkgoRecover()
						defer wg.Done()
						val, err := b.Do(ctx, requests[i]).Await(ctx)
						Expect(err).To(BeNil())
						Expect(val).To(Equal(responses[i].Response))
					}(i)
					<-time.After(50 * time.Millisecond)
				}

				wg.Wait()
			})
		})

		Describe("can handle mix value and error response", func() {
			var (
				actionCount int
				requests    []string
				responses   []Response[string]
			)

			BeforeEach(func() {
				actionCount = gofakeit.Number(5, 30)
				options = append(options, WithMaxBatchSize(actionCount))
				requests = make([]string, actionCount)
				responses = make([]Response[string], actionCount)

				for i := 0; i < actionCount; i++ {
					requests[i] = fmt.Sprintf("req: #%d", i)
					if i%2 == 0 {
						responses[i] = Response[string]{
							Response: fmt.Sprintf("res: #%d", i),
						}
					} else {
						responses[i] = Response[string]{
							Error: fmt.Errorf("err: #%d", i),
						}
					}
				}

				action.EXPECT().Perform(ctx, requests).Times(1).Return(responses)
			})

			It("should have corrsponding value and error", func() {
				wg.Add(actionCount)
				for i := 0; i < actionCount; i++ {
					go func(i int) {
						defer GinkgoRecover()
						defer wg.Done()
						val, err := b.Do(ctx, requests[i]).Await(ctx)
						if err != nil {
							Expect(err).To(Equal(responses[i].Error))
						} else {
							Expect(err).ToNot(HaveOccurred())
						}
						Expect(val).To(Equal(responses[i].Response))
					}(i)
					<-time.After(50 * time.Millisecond)
				}

				wg.Wait()
			})
		})

		Describe("split batch if over max batch size", func() {
			var (
				batchSize   int
				actionCount int
				requests    []string
				responses   []Response[string]
			)

			BeforeEach(func() {
				batchSize = gofakeit.Number(3, 5)
				actionCount = batchSize * gofakeit.Number(1, 3)
				options = append(options, WithMaxBatchSize(batchSize))
				requests = make([]string, actionCount)
				responses = make([]Response[string], actionCount)

				for i := 0; i < actionCount; i++ {
					requests[i] = fmt.Sprintf("req: #%d", i)
					responses[i] = Response[string]{
						Response: fmt.Sprintf("res: #%d", i),
					}
				}

				for i := 0; i < actionCount; i += batchSize {
					action.EXPECT().Perform(ctx, requests[i:i+batchSize]).Times(1).Return(responses[i : i+batchSize])
				}
			})

			It("should split batch if over max batch size", func() {
				wg.Add(actionCount)
				for i := 0; i < actionCount; i++ {
					go func(i int) {
						defer GinkgoRecover()
						defer wg.Done()
						val, err := b.Do(ctx, requests[i]).Await(ctx)
						Expect(err).To(BeNil())
						Expect(val).To(Equal(responses[i].Response))
					}(i)
					<-time.After(50 * time.Millisecond)
				}

				wg.Wait()
			})
		})

		Describe("can graceful shutdown", func() {
			var (
				batchSize   int
				actionCount int
				requests    []string
				responses   []Response[string]
			)

			BeforeEach(func() {
				batchSize = gofakeit.Number(3, 5)
				actionCount = batchSize * gofakeit.Number(1, 3)
				options = append(options,
					WithMaxBatchSize(batchSize),
					WithScheduler(NewTestGracefulScheduler()),
				)
				requests = make([]string, actionCount)
				responses = make([]Response[string], actionCount)

				for i := 0; i < actionCount; i++ {
					requests[i] = fmt.Sprintf("req: #%d", i)
					responses[i] = Response[string]{
						Response: fmt.Sprintf("res: #%d", i),
					}
				}

				for i := 0; i < actionCount; i += batchSize {
					action.EXPECT().Perform(ctx, requests[i:i+batchSize]).Times(1).Return(responses[i : i+batchSize])
				}
			})

			It("should graceful shutdown", func() {
				wg.Add(actionCount)
				for i := 0; i < actionCount; i++ {
					go func(i int) {
						defer GinkgoRecover()
						defer wg.Done()
						val, err := b.Do(ctx, requests[i]).Await(ctx)
						Expect(err).To(BeNil())
						Expect(val).To(Equal(responses[i].Response))
					}(i)
					<-time.After(50 * time.Millisecond)
				}

				b.Shutdown()
				wg.Wait()
			})
		})

		Describe("could handle schedule invoke callback multiple time", func() {
			var (
				batchSize   int
				actionCount int
				requests    []string
				responses   []Response[string]
			)

			BeforeEach(func() {
				batchSize = gofakeit.Number(3, 5)
				actionCount = batchSize * gofakeit.Number(1, 3)
				options = append(options,
					WithMaxBatchSize(batchSize),
					WithScheduler(NewMalfunctioingScheduler()),
				)
				requests = make([]string, actionCount)
				responses = make([]Response[string], actionCount)

				for i := 0; i < actionCount; i++ {
					requests[i] = fmt.Sprintf("req: #%d", i)
					responses[i] = Response[string]{
						Response: fmt.Sprintf("res: #%d", i),
					}
				}

				for i := 0; i < actionCount; i += batchSize {
					action.EXPECT().Perform(ctx, requests[i:i+batchSize]).Times(1).Return(responses[i : i+batchSize])
				}
			})

			It("without panic", func() {
				wg.Add(actionCount)
				for i := 0; i < actionCount; i++ {
					go func(i int) {
						defer GinkgoRecover()
						defer wg.Done()
						val, err := b.Do(ctx, requests[i]).Await(ctx)
						Expect(err).To(BeNil())
						Expect(val).To(Equal(responses[i].Response))
					}(i)
					<-time.After(50 * time.Millisecond)
				}
				b.Shutdown()
				wg.Wait()
			})
		})

		Describe("can handle concurrency control error", func() {
			var (
				batchSize   int
				actionCount int
				requests    []string
				responses   []Response[string]
				expectedErr error

				cc *MockConcurrencyControl
			)

			BeforeEach(func() {
				batchSize = gofakeit.Number(3, 5)
				actionCount = batchSize * gofakeit.Number(1, 3)
				cc = NewMockConcurrencyControl(ctrl)
				options = append(options,
					WithMaxBatchSize(batchSize),
					WithConcurrencyControl(cc),
				)
				requests = make([]string, actionCount)
				responses = make([]Response[string], actionCount)

				for i := 0; i < actionCount; i++ {
					requests[i] = fmt.Sprintf("req: #%d", i)
					responses[i] = Response[string]{
						Response: fmt.Sprintf("res: #%d", i),
					}
				}

				expectedErr = fmt.Errorf("error")

				cc.EXPECT().Acquire(gomock.Any()).MinTimes(1).Return(nil, expectedErr)
				action.EXPECT().Perform(ctx, gomock.Any()).Times(0)
			})

			It("should return error", func() {
				wg.Add(actionCount)
				for i := 0; i < actionCount; i++ {
					go func(i int) {
						defer wg.Done()

						val, err := b.Do(ctx, requests[i]).Await(ctx)

						Expect(err).To(Equal(expectedErr))
						Expect(val).To(Equal(""))
					}(i)
					<-time.After(50 * time.Millisecond)
				}

				b.Shutdown()
				wg.Wait()
			})
		})

		Describe("should failed if already shutdown", func() {
			It("should failed if already shutdown", func() {
				b.Shutdown()
				thunk := b.Do(ctx, "foo")
				val, err := thunk.Await(ctx)
				Expect(val).Should(Equal(""))
				Expect(err).Should(HaveOccurred())
			})
		})
	})
})

type TestGracefulScheduler struct{}

func NewTestGracefulScheduler() Scheduler {
	return &TestGracefulScheduler{}
}

func (t *TestGracefulScheduler) Schedule(ctx context.Context, batch Batch, callback SchedulerCallback) {
	select {
	case <-ctx.Done():
		return
	case <-batch.Dispatch():
		callback.Call()
	}
}

type TestManuallyTriggerScheduler struct{}

func NewTestManuallyTriggerScheduler() Scheduler {
	return &TestGracefulScheduler{}
}

func (t *TestManuallyTriggerScheduler) Schedule(ctx context.Context, batch Batch, callback SchedulerCallback) {
	select {
	case <-ctx.Done():
		return
	case <-batch.Dispatch():
		callback.Call()
	}
}

type MalfunctioingScheduler struct{}

func NewMalfunctioingScheduler() Scheduler {
	return &MalfunctioingScheduler{}
}

func (t *MalfunctioingScheduler) Schedule(ctx context.Context, batch Batch, callback SchedulerCallback) {
	select {
	case <-ctx.Done():
		return
	case <-batch.Dispatch():
		callback.Call()
		callback.Call()
	}
}
