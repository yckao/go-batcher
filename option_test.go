package batcher

import (
	"github.com/brianvoe/gofakeit/v6"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"context"
)

var _ = Describe("Option", func() {
	var (
		ctx context.Context

		action  *MockAction[string, string]
		options []option
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
			options = append(options, WithMaxBatchSize(size))
		})

		It("should set max batch size", func() {
			Expect(b.maxBatchSize).To(Equal(size))
		})
	})

	Describe("can set schedule function", func() {
		var scheduler Scheduler
		BeforeEach(func() {
			scheduler = NewInstantScheduler()
			options = append(options, WithScheduler(scheduler))
		})

		It("should set schedule function", func() {
			Expect(b.scheduler).To(Equal(scheduler))
		})
	})

	Describe("can set concurrency control", func() {
		var cc ConcurrencyControl
		BeforeEach(func() {
			cc = NewUnlimitedConcurrencyControl()
			options = append(options, WithConcurrencyControl(cc))
		})

		It("should set concurrency control", func() {
			Expect(b.concurrencyControl).To(Equal(cc))
		})
	})

	Describe("can set metric set", func() {
		var metrics *MetricSet
		BeforeEach(func() {
			metrics = NewMetricSet("foo", "bar", nil)
			options = append(options, WithMetricSet(metrics))
		})

		It("should set metric set", func() {
			Expect(b.metrics).To(Equal(metrics))
		})
	})
})
