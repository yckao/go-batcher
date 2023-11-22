package batcher

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gleak"
)

var _ = Describe("DataLoader", func() {
	BeforeEach(func() {
		goods := Goroutines()
		DeferCleanup(func() {
			Eventually(Goroutines).ShouldNot(HaveLeaked(goods))
		})
	})

	It("can batch do request", func() {
		ctx := context.TODO()
		loadCount := 0
		batchFn := func(ctx context.Context, keys []string) []Response[string] {
			defer GinkgoRecover()

			loadCount += 1
			response := make([]Response[string], len(keys))
			for index, key := range keys {
				response[index] = Response[string]{
					Response: "res:" + key,
				}
			}

			return response
		}

		batcher := New[string, string](ctx, batchFn)
		wg := &sync.WaitGroup{}
		wg.Add(4)

		for i := 0; i < 4; i++ {
			k := fmt.Sprintf("key%d", i)
			r := fmt.Sprintf("res:key%d", i)
			go func() {
				defer GinkgoRecover()
				defer wg.Done()

				val, err := batcher.Do(ctx, k).Await(ctx)
				Expect(err).To(BeNil())
				Expect(val).To(Equal(r))
			}()
		}

		wg.Wait()
		Expect(loadCount).To(Equal(1))
	})

	It("split batch if over max batch size", func() {
		ctx := context.TODO()
		loadCount := 0
		batchFn := func(ctx context.Context, keys []string) []Response[string] {
			loadCount += 1
			result := make([]Response[string], len(keys))
			for index, key := range keys {
				result[index] = Response[string]{
					Response: "res:" + key,
				}
			}
			return result
		}

		batcher := New[string, string](
			ctx, batchFn,
			WithMaxBatchSize[string, string](2),
			WithScheduleFn[string, string](NewTimeWindowScheduler(1*time.Second)),
		)
		wg := &sync.WaitGroup{}
		wg.Add(5)

		for i := 0; i < 5; i++ {
			k := fmt.Sprintf("key%d", i)
			r := fmt.Sprintf("res:key%d", i)
			go func() {
				defer GinkgoRecover()
				defer wg.Done()

				val, err := batcher.Do(ctx, k).Await(ctx)
				Expect(err).To(BeNil())
				Expect(val).To(Equal(r))
			}()
		}

		wg.Wait()
		Expect(loadCount).To(Equal(3))
	})

	It("can manually dispatch batch", func() {
		ctx := context.TODO()
		loadCount := 0
		values := map[string]string{
			"foo": "foobar",
		}

		batchFn := func(ctx context.Context, keys []string) []Response[string] {
			defer GinkgoRecover()

			loadCount += 1
			result := make([]Response[string], len(keys))
			for index, key := range keys {
				result[index] = Response[string]{
					Response: values[key],
				}
			}

			return result
		}

		batcher := New[string, string](ctx, batchFn, WithMaxBatchSize[string, string](200), WithScheduleFn[string, string](NewTimeWindowScheduler(1*time.Second)))
		start := time.Now()
		thunk := batcher.Do(ctx, "foo")
		batcher.Dispatch()
		val, err := thunk.Await(ctx)
		Expect(time.Now().Before(start.Add(500 * time.Millisecond))).To(BeTrue())
		Expect(val).To(Equal(values["foo"]))
		Expect(err).To(BeNil())
		Expect(loadCount).To(Equal(1))
	})

	It("can early dispatch if full", func() {
		ctx := context.TODO()
		loadCount := 0
		values := map[string]string{
			"foo": "foobar",
		}

		batchFn := func(ctx context.Context, keys []string) []Response[string] {
			defer GinkgoRecover()

			loadCount += 1
			result := make([]Response[string], len(keys))
			for index, key := range keys {
				result[index] = Response[string]{
					Response: values[key],
				}
			}

			return result
		}

		batcher := New[string, string](ctx, batchFn, WithMaxBatchSize[string, string](0), WithScheduleFn[string, string](NewTimeWindowScheduler(1*time.Second)))
		start := time.Now()
		thunk := batcher.Do(ctx, "foo")
		val, err := thunk.Await(ctx)
		Expect(time.Now().Before(start.Add(500 * time.Millisecond))).To(BeTrue())
		Expect(val).To(Equal(values["foo"]))
		Expect(err).To(BeNil())
		Expect(loadCount).To(Equal(1))
	})

	It("dispatch multiple time should not crash dataloader", func() {
		ctx := context.TODO()
		batchLoadFn := func(ctx context.Context, keys []string) []Response[string] {
			defer GinkgoRecover()

			result := make([]Response[string], len(keys))
			return result
		}

		loader := New[string, string](ctx, batchLoadFn).(*batcher[string, string])
		loader.dispatch()
		loader.dispatch()
	})
})

func TestDataloader(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Dataloader Suite")
}
