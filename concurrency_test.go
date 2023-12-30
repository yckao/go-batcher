package batcher

import (
	"context"
	"sync"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gleak"
)

var _ = Describe("UnlimitedConcurrencyControl", func() {
	BeforeEach(func() {
		goods := Goroutines()
		DeferCleanup(func() {
			Eventually(Goroutines).ShouldNot(HaveLeaked(goods))
		})
	})

	It("should always return a token", func() {
		cc := unlimitedConcurrencyControl{}
		token, err := cc.Acquire(context.Background())
		Expect(err).Should(BeNil())
		Expect(token).ShouldNot(BeNil())
		Expect(token.Release).ShouldNot(BeNil())
	})
})

var _ = Describe("LimitedConcurrencyControl", func() {
	var (
		ctx        context.Context
		cancelFunc context.CancelFunc

		cc    *limitedConcurrencyControl
		limit int
	)

	BeforeEach(func() {
		ctx, cancelFunc = context.WithCancel(context.Background())
		limit = gofakeit.Number(3, 10)
		cc = NewLimitedConcurrencyControl(limit).(*limitedConcurrencyControl)
	})

	AfterEach(func() {
		cancelFunc()
	})

	It("should return a token if concurrency limit is not reached", func() {
		token, err := cc.Acquire(ctx)
		Expect(err).Should(BeNil())
		Expect(token).ShouldNot(BeNil())
		Expect(token.Release).ShouldNot(BeNil())
	})

	It("should block when concurrency limit is reached", func() {
		tokens := make([]ConcurrencyToken, limit)
		for i := 0; i < limit; i++ {
			token, err := cc.Acquire(ctx)
			tokens[i] = token
			Expect(err).Should(BeNil())
			Expect(token).ShouldNot(BeNil())
			Expect(token.Release).ShouldNot(BeNil())
		}

		wg := &sync.WaitGroup{}
		wg.Add(2)
		go func() {
			defer GinkgoRecover()
			defer wg.Done()

			_, err := cc.Acquire(ctx)
			Expect(err).Should(HaveOccurred())
		}()

		go func() {
			defer wg.Done()
			<-time.After(10 * time.Millisecond)
			cancelFunc()
		}()

		wg.Wait()
	})

	It("should release a token when release is called", func() {
		tokens := make([]ConcurrencyToken, limit)
		for i := 0; i < limit; i++ {
			token, err := cc.Acquire(ctx)
			tokens[i] = token
			Expect(err).Should(BeNil())
			Expect(token).ShouldNot(BeNil())
			Expect(token.Release).ShouldNot(BeNil())
		}

		Expect(len(cc.sem)).Should(Equal(limit))
		for i, token := range tokens {
			token.Release()
			Expect(len(cc.sem)).Should(Equal(limit - i - 1))
		}

		Expect(len(cc.sem)).Should(Equal(0))
	})

	It("should queue up requests when concurrency limit is reached", func() {
		tokens := make([]ConcurrencyToken, limit)

		for i := 0; i < limit; i++ {
			token, err := cc.Acquire(ctx)
			tokens[i] = token
			Expect(err).Should(BeNil())
			Expect(token).ShouldNot(BeNil())
			Expect(token.Release).ShouldNot(BeNil())
		}

		dataCount := gofakeit.Number(2, limit*2)
		expected := []string{}
		actual := []string{}
		for i := 0; i < dataCount; i++ {
			expected = append(expected, gofakeit.LetterN(10))
		}

		wg := &sync.WaitGroup{}
		wg.Add(dataCount)

		for i := 0; i < dataCount; i++ {
			go func(i int) {
				defer wg.Done()

				token, err := cc.Acquire(ctx)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(token).ShouldNot(BeNil())
				actual = append(actual, expected[i])

				token.Release()
			}(i)
			<-time.After(10 * time.Millisecond)
		}

		tokens[0].Release()
		wg.Wait()

		Expect(actual).Should(Equal(expected))
		Expect(len(cc.sem)).Should(Equal(limit - 1))
	})

	It("should return error directly if context is cancelled", func() {
		cancelFunc()

		_, err := cc.Acquire(ctx)
		Expect(err).Should(HaveOccurred())
	})

	It("should able specify queue length", func() {
		queueSize := gofakeit.Number(10, 100)
		cc = NewLimitedConcurrencyControl(limit, WithLimitedConcurrencyControlQueueSize(queueSize)).(*limitedConcurrencyControl)
		Expect(cap(cc.queue)).Should(Equal(queueSize))
	})
})
