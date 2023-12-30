package batcher

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"context"
	"reflect"
)

type testConcurrencyControl struct{}

func (c *testConcurrencyControl) Acquire(ctx context.Context) (func(), error) { return func() {}, nil }

var _ = Describe("Option", func() {
	It("can set max batch size", func() {
		dl := New[string, string](context.TODO(), func(ctx context.Context, keys []string) []Response[string] { return []Response[string]{} }, WithMaxBatchSize[string, string](10))
		Expect(dl.(*batcher[string, string]).maxBatchSize).To(Equal(10))
	})

	It("can set schedule function", func() {
		scheduler := NewInstantScheduler()
		dl := New[string, string](context.TODO(), func(ctx context.Context, keys []string) []Response[string] { return []Response[string]{} }, WithScheduler[string, string](scheduler))

		Expect(dl.(*batcher[string, string]).scheduler).To(Equal(scheduler))
	})

	It("can set concurrency control", func() {
		concurrencyControl := &testConcurrencyControl{}
		dl := New[string, string](context.TODO(), func(ctx context.Context, keys []string) []Response[string] { return []Response[string]{} }, WithConcurrencyControl[string, string](concurrencyControl))

		pointer1 := reflect.ValueOf(dl.(*batcher[string, string]).concurrencyControl).Pointer()
		pointer2 := reflect.ValueOf(concurrencyControl).Pointer()

		Expect(pointer1).To(Equal(pointer2))
	})
})
