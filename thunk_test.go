package batcher

import (
	"context"
	"errors"
	"sync"

	"github.com/brianvoe/gofakeit/v6"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gleak"
)

var _ = Describe("Thunk", func() {

	var (
		ctx        context.Context
		cancelFunc context.CancelFunc

		expectedValue string
		expectedError error
		thunk         Thunk[string]
	)

	BeforeEach(func() {
		goods := Goroutines()
		DeferCleanup(func() {
			Eventually(Goroutines).ShouldNot(HaveLeaked(goods))
		})
	})

	BeforeEach(func() {
		ctx, cancelFunc = context.WithCancel(context.TODO())

		thunk = NewThunk[string]()
		expectedValue = gofakeit.LetterN(10)
		expectedError = errors.New(gofakeit.LetterN(10))
	})

	AfterEach(func() {
		cancelFunc()
	})

	It("can check state", func() {
		Expect(thunk.Pending()).To(BeTrue())
		Expect(thunk.Fulfilled()).To(BeFalse())
		Expect(thunk.Rejected()).To(BeFalse())

		thunk.Set(ctx, expectedValue)
		Expect(thunk.Pending()).To(BeFalse())
		Expect(thunk.Fulfilled()).To(BeTrue())

		thunk.Error(ctx, expectedError)
		Expect(thunk.Pending()).To(BeFalse())
		Expect(thunk.Rejected()).To(BeTrue())
	})

	It("can wait value before it's set", func() {
		steps := map[string]chan struct{}{
			"pendingChecked": make(chan struct{}),
		}

		wg := &sync.WaitGroup{}
		wg.Add(2)

		go func() {
			defer wg.Done()

			Expect(thunk.Pending()).To(BeTrue())
			steps["pendingChecked"] <- struct{}{}
			val, err := thunk.Await(ctx)
			Expect(val).Should(Equal(expectedValue))
			Expect(err).ShouldNot(HaveOccurred())
		}()

		go func() {
			defer wg.Done()

			<-steps["pendingChecked"]
			thunk.Set(ctx, expectedValue)
		}()

		wg.Wait()
	})

	It("can wait value after it's set", func() {
		thunk.Set(ctx, expectedValue)
		val, err := thunk.Await(ctx)
		Expect(val).Should(Equal(expectedValue))
		Expect(err).ShouldNot(HaveOccurred())
	})

	It("can get value multiple times", func() {
		thunk.Set(ctx, expectedValue)
		val, err := thunk.Await(ctx)
		Expect(val).Should(Equal(expectedValue))
		Expect(err).ShouldNot(HaveOccurred())

		val, err = thunk.Await(ctx)
		Expect(val).Should(Equal(expectedValue))
		Expect(err).ShouldNot(HaveOccurred())
	})

	It("can get error multiple times", func() {
		thunk.Error(ctx, expectedError)
		val, err := thunk.Await(ctx)
		Expect(val).Should(Equal(""))
		Expect(err).Should(Equal(expectedError))

		val, err = thunk.Await(ctx)
		Expect(val).Should(Equal(""))
		Expect(err).Should(Equal(expectedError))
	})

	It("can wait error before it's set", func() {
		steps := map[string]chan struct{}{
			"pendingChecked": make(chan struct{}),
		}

		wg := &sync.WaitGroup{}
		wg.Add(2)

		go func() {
			defer wg.Done()

			Expect(thunk.Pending()).To(BeTrue())
			steps["pendingChecked"] <- struct{}{}
			val, err := thunk.Await(ctx)
			Expect(val).Should(Equal(""))
			Expect(err).Should(Equal(expectedError))
		}()

		go func() {
			defer wg.Done()

			<-steps["pendingChecked"]
			thunk.Error(ctx, expectedError)
		}()

		wg.Wait()
	})

	It("can wait error after it's set", func() {
		thunk.Error(ctx, expectedError)
		val, err := thunk.Await(ctx)
		Expect(val).Should(Equal(""))
		Expect(err).Should(Equal(expectedError))
	})

	It("can override value with value", func() {
		thunk.Set(ctx, "bar")
		thunk.Set(ctx, expectedValue)

		val, err := thunk.Await(ctx)
		Expect(val).Should(Equal(expectedValue))
		Expect(err).ShouldNot(HaveOccurred())
	})

	It("can override value with error", func() {
		thunk.Set(ctx, "bar")
		thunk.Error(ctx, expectedError)

		val, err := thunk.Await(ctx)
		Expect(val).Should(Equal(""))
		Expect(err).Should(Equal(expectedError))
	})

	It("can override error with value", func() {
		thunk.Error(ctx, errors.New("bar"))
		thunk.Set(ctx, expectedValue)

		val, err := thunk.Await(ctx)
		Expect(val).Should(Equal(expectedValue))
		Expect(err).ShouldNot(HaveOccurred())
	})

	It("can override error with error", func() {
		thunk.Error(ctx, errors.New("bar"))
		thunk.Error(ctx, expectedError)

		val, err := thunk.Await(ctx)
		Expect(val).Should(Equal(""))
		Expect(err).Should(Equal(expectedError))
	})

	It("can cancelFunc context", func() {
		ctx, cancelFunc := context.WithCancel(context.TODO())

		wg := &sync.WaitGroup{}
		wg.Add(2)

		go func() {
			defer wg.Done()
			thunk.Await(ctx)
		}()

		go func() {
			defer wg.Done()
			cancelFunc()
		}()

		wg.Wait()
	})
})
