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

		go scheduler.Schedule(context.TODO(), &MockBatch{}, func() { runned = true })
		Expect(runned).To(BeFalse())

		<-time.After(300 * time.Millisecond)
		Expect(runned).To(BeTrue())
	})

	It("can cancel", func() {
		scheduler := NewTimeWindowScheduler(200 * time.Millisecond)
		runned := false

		ctx, cancel := context.WithCancel(context.TODO())
		go scheduler.Schedule(ctx, &MockBatch{}, func() { runned = true })
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

var _ = Describe("DefaultConcurrencyControl", func() {
	var (
		ctx                context.Context
		cancelFunc         context.CancelFunc
		concurrencyControl BatchConcurrencyControl
		concurrencyLimit   int
	)

	BeforeEach(func() {
		concurrencyLimit = 2
		concurrencyControl = NewDefaultConcurrencyControl(concurrencyLimit)
		ctx, cancelFunc = context.WithCancel(context.Background())
	})

	AfterEach(func() {
		cancelFunc()
	})

	Describe("Acquire", func() {
		It("should acquire a slot when under the concurrency limit", func() {
			release, err := concurrencyControl.Acquire(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(release).NotTo(BeNil())
			release() // Don't forget to release after acquiring to clean up
		})

		It("should block when the concurrency limit is reached", func() {
			for i := 0; i < concurrencyLimit; i++ {
				_, err := concurrencyControl.Acquire(ctx)
				Expect(err).NotTo(HaveOccurred())
			}

			acquireAttempted := make(chan bool)

			go func() {
				_, _ = concurrencyControl.Acquire(ctx)
				acquireAttempted <- true
			}()

			Consistently(acquireAttempted).ShouldNot(Receive(), "Acquire should block and not complete")

			cancelFunc() // Cleanup: Ensure that the goroutine can exit
		})

		Context("when the semaphore is full and a goroutine is waiting", func() {
			It("should wait for release before acquiring", func() {
				releases := make([]func(), concurrencyLimit)
				// Fill up the semaphore
				for i := 0; i < concurrencyLimit; i++ {
					release, err := concurrencyControl.Acquire(ctx)
					Expect(err).NotTo(HaveOccurred())
					releases[i] = release
				}

				doneAcquiring := make(chan bool)

				// This should block since the semaphore is full
				go func() {
					_, err := concurrencyControl.Acquire(ctx)
					Expect(err).NotTo(HaveOccurred())
					doneAcquiring <- true
				}()

				Consistently(doneAcquiring).ShouldNot(Receive(), "Acquire should block and not complete")

				// Release one spot
				releases[0]()

				Eventually(doneAcquiring).Should(Receive(), "Acquire should succeed after a release")
			})
		})

		It("should return an error when the context is cancelled", func() {
			cancelFunc() // Cancel the context

			_, err := concurrencyControl.Acquire(ctx)
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(context.Canceled))
		})
	})

	Describe("release", func() {
		It("should release a slot making it available for others", func() {
			release, err := concurrencyControl.Acquire(ctx)
			Expect(err).NotTo(HaveOccurred())

			release() // Release the slot

			Eventually(func() (func(), error) {
				return concurrencyControl.Acquire(ctx)
			}).ShouldNot(BeNil())
		})

		Context("when there are goroutines waiting on the queue", func() {
			It("should unblock a waiting goroutine", func() {
				releases := make([]func(), concurrencyLimit)
				// Fill up the semaphore
				for i := 0; i < concurrencyLimit; i++ {
					release, err := concurrencyControl.Acquire(ctx)
					Expect(err).NotTo(HaveOccurred())
					releases[i] = release
				}

				waiting := make(chan bool)

				// Start a goroutine that will wait
				go func() {
					_, err := concurrencyControl.Acquire(ctx)
					Expect(err).NotTo(HaveOccurred())
					waiting <- true
				}()

				time.Sleep(100 * time.Millisecond) // Give time for the goroutine to block

				// This release should unblock the waiting goroutine
				releases[0]()

				Eventually(waiting).Should(Receive(), "The waiting goroutine should have been unblocked")
			})
		})
	})
})
