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
		scheduleFn := (func(ctx context.Context, _ Batch, callback func()) { callback() })
		dl := New[string, string](context.TODO(), func(ctx context.Context, keys []string) []Response[string] { return []Response[string]{} }, WithScheduleFn[string, string](scheduleFn))

		pointer1 := reflect.ValueOf(dl.(*batcher[string, string]).scheduleFn).Pointer()
		pointer2 := reflect.ValueOf(scheduleFn).Pointer()

		Expect(pointer1).To(Equal(pointer2))
	})

	It("can set concurrency control", func() {
		concurrencyControl := &testConcurrencyControl{}
		dl := New[string, string](context.TODO(), func(ctx context.Context, keys []string) []Response[string] { return []Response[string]{} }, WithConcurrencyControl[string, string](concurrencyControl))

		pointer1 := reflect.ValueOf(dl.(*batcher[string, string]).concurrencyControl).Pointer()
		pointer2 := reflect.ValueOf(concurrencyControl).Pointer()

		Expect(pointer1).To(Equal(pointer2))
	})
})
