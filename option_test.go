package batcher

import (
	"github.com/brianvoe/gofakeit/v6"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"context"
)

type testConcurrencyControl struct{}

func (c *testConcurrencyControl) Acquire(ctx context.Context) (func(), error) { return func() {}, nil }

var _ = Describe("Option", func() {
	var (
		ctx context.Context

		action  *MockAction[string, string]
		options []option[string, string]
		b       *batcher[string, string]
	)

	BeforeEach(func() {
		ctx = context.TODO()
		action = NewMockAction[string, string](ctrl)
	})

	JustBeforeEach(func() {
		b = New[string, string](ctx, action, options...).(*batcher[string, string])
	})

	Describe("can set max batch size", func() {
		var size int
		BeforeEach(func() {
			size = gofakeit.Number(10, 100)
			options = append(options, WithMaxBatchSize[string, string](size))
		})

		It("should set max batch size", func() {
			Expect(b.maxBatchSize).To(Equal(size))
		})
	})

	Describe("can set schedule function", func() {
		var scheduler Scheduler
		BeforeEach(func() {
			scheduler = NewInstantScheduler()
			options = append(options, WithScheduler[string, string](scheduler))
		})

		It("should set schedule function", func() {
			Expect(b.scheduler).To(Equal(scheduler))
		})
	})

	Describe("can set concurrency control", func() {
		var cc ConcurrencyControl
		BeforeEach(func() {
			cc = NewUnlimitedConcurrencyControl()
			options = append(options, WithConcurrencyControl[string, string](cc))
		})

		It("should set concurrency control", func() {
			Expect(b.concurrencyControl).To(Equal(cc))
		})
	})
})
