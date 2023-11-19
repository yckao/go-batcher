package batcher

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"context"
	"time"
)

type MockBatch struct{}

func (*MockBatch) Full() <-chan struct{} {
	return make(<-chan struct{})
}
func (*MockBatch) Dispatch() <-chan struct{} {
	return make(<-chan struct{})
}

var _ = Describe("NewTimeWindowScheduler", func() {
	It("should run after specified duration", func() {
		scheduler := NewTimeWindowScheduler(200 * time.Millisecond)
		runned := false

		go scheduler(context.TODO(), &MockBatch{}, func() { runned = true })
		Expect(runned).To(BeFalse())

		<-time.After(300 * time.Millisecond)
		Expect(runned).To(BeTrue())
	})

	It("can cancel", func() {
		scheduler := NewTimeWindowScheduler(200 * time.Millisecond)
		runned := false

		ctx, cancel := context.WithCancel(context.TODO())
		go scheduler(ctx, &MockBatch{}, func() { runned = true })
		cancel()

		<-time.After(200 * time.Millisecond)
		Expect(runned).To(BeFalse())
	})
})

var _ = Describe("UnlimitedConcurrencyControl", func() {
	It("should always return a release function", func() {
		cc := UnlimitedConcurrencyControl{}
		release, err := cc.Acquire(context.Background())
		Expect(err).To(BeNil())
		Expect(release).ToNot(BeNil())
	})
})

// var _ = Describe("defaultConcurrencyControl", func() {
// 	It("should not block when not exceeding max concurrency", func() {
// 		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
// 		defer cancel()

// 		success := make(chan struct{})
// 		go func() {
// 			maxConcurrency := 5
// 			cc := NewDefaultConcurrencyControl(maxConcurrency)
// 			for i := 0; i < maxConcurrency; i++ {
// 				cc.Acquire(ctx)
// 			}
// 			success <- struct{}{}
// 		}()

// 		select {
// 		case <-ctx.Done():
// 			Fail("context timeout")
// 		case <-success:
// 		}
// 	})

// 	It("should block when exceeding max concurrency", func() {
// 		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
// 		defer cancel()

// 		finished := make(chan struct{})
// 		go func() {
// 			maxConcurrency := 5
// 			cc := NewDefaultConcurrencyControl(maxConcurrency)
// 			for i := 0; i < maxConcurrency; i++ {
// 				cc.Acquire(ctx)
// 			}
// 			cc.Acquire(ctx)
// 			finished <- struct{}{}
// 		}()

// 		select {
// 		case <-ctx.Done():
// 		case <-finished:
// 			Fail("Failed to block when exceeding max concurrency")
// 		}
// 	})

// 	It("should resume when concurrency released", func() {
// 		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
// 		defer cancel()

// 		finished := make(chan struct{})
// 		go func() {
// 			maxConcurrency := 5
// 			cc := NewDefaultConcurrencyControl(maxConcurrency)
// 			for i := 0; i < maxConcurrency; i++ {
// 				release, _ := cc.Acquire(ctx)
// 				Expect(release).ToNot(BeNil())
// 				release()
// 			}
// 			cc.Acquire(ctx)
// 			finished <- struct{}{}
// 		}()

// 		select {
// 		case <-ctx.Done():
// 			Fail("context timeout")
// 		case <-finished:
// 		}
// 	})
// })

var _ = Describe("defaultConcurrencyControl", func() {
	var (
		concurrencyControl *defaultConcurrencyControl
		concurrency        int
		ctx                context.Context
		cancel             context.CancelFunc
	)

	BeforeEach(func() {
		concurrency = 2 // Example concurrency level
		concurrencyControl = NewDefaultConcurrencyControl(concurrency).(*defaultConcurrencyControl)
		ctx, cancel = context.WithCancel(context.Background())
	})

	AfterEach(func() {
		cancel()
	})

	Describe("Acquiring a token", func() {
		Context("when there is capacity", func() {
			It("should acquire a token successfully", func() {
				release, err := concurrencyControl.Acquire(ctx)
				Expect(err).NotTo(HaveOccurred())
				Expect(release).NotTo(BeNil())

				release() // Release the token to cleanup
			})
		})

		Context("when there is no capacity", func() {
			It("should wait and eventually acquire a token", func() {
				By("filling up the semaphore to simulate full capacity")
				for i := 0; i < concurrency; i++ {
					concurrencyControl.sem <- struct{}{}
				}

				By("trying to acquire a token when the semaphore is full")
				acquireAttempted := make(chan bool)
				go func() {
					defer GinkgoRecover()
					_, err := concurrencyControl.Acquire(ctx)
					acquireAttempted <- err == nil
				}()

				Consistently(acquireAttempted).ShouldNot(Receive(BeTrue()), "Acquire should not succeed while the semaphore is full")

				By("releasing a token to allow acquire to succeed")
				<-concurrencyControl.sem
				concurrencyControl.releaseNext()
				Eventually(acquireAttempted).Should(Receive(BeTrue()), "Acquire should succeed after a token is released")
			})
		})

		Context("when the context is canceled", func() {
			It("should return an error", func() {
				cancel() // Cancel the context
				_, err := concurrencyControl.Acquire(ctx)
				<-time.After(10 * time.Millisecond)
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("Releasing a token", func() {
		Context("when there is a waiting goroutine", func() {
			It("should release the next waiting goroutine", func() {
				By("filling up the semaphore and queuing a waiting goroutine")
				for i := 0; i < concurrency; i++ {
					concurrencyControl.sem <- struct{}{}
				}
				waitingChan := make(chan struct{}, 1)
				concurrencyControl.queue <- waitingChan

				By("calling releaseNext to release the waiting goroutine")
				concurrencyControl.releaseNext()

				Eventually(waitingChan).Should(Receive(), "The waiting goroutine should receive a release signal")
			})
		})
	})
})
