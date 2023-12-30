package batcher

import (
	"context"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gleak"
	"github.com/yckao/go-batcher/internal/mock"
)

var _ = Describe("TimeWindowScheduler", func() {
	var (
		ctx        context.Context
		cancelFunc context.CancelFunc

		mockClock    *mock.MockClock
		mockTimer    *mock.MockTimer
		mockBatch    *mock.MockBatch
		mockCallback *mock.MockCallback

		timeTrigger     chan time.Time
		dispatchTrigger chan struct{}
		fullTrigger     chan struct{}

		scheduler *TimeWindowScheduler
	)

	BeforeEach(func() {
		goods := Goroutines()
		DeferCleanup(func() {
			Eventually(Goroutines).ShouldNot(HaveLeaked(goods))
		})
	})

	BeforeEach(func() {
		ctx, cancelFunc = context.WithCancel(context.TODO())

		mockClock = mock.NewMockClock(ctrl)
		scheduler = &TimeWindowScheduler{
			clock:      mockClock,
			timeWindow: time.Duration(gofakeit.Second()),
		}

		mockCallback = mock.NewMockCallback(ctrl)

		mockTimer = mock.NewMockTimer(ctrl)
		timeTrigger = make(chan time.Time)
		mockTimer.EXPECT().C().Return(timeTrigger)
		mockTimer.EXPECT().Stop()

		mockBatch = mock.NewMockBatch(ctrl)
		dispatchTrigger = make(chan struct{})
		fullTrigger = make(chan struct{})
		mockBatch.EXPECT().Dispatch().Return(dispatchTrigger)
		mockBatch.EXPECT().Full().Return(fullTrigger)

		mockClock.EXPECT().NewTimer(scheduler.timeWindow).Return(mockTimer)
	})

	AfterEach(func() {
		cancelFunc()
	})

	It("should not call callback if not triggered", func() {
		mockCallback.EXPECT().Call().Times(0)

		go scheduler.Schedule(ctx, mockBatch, mockCallback)
	})

	It("should called callback if timer triggered", func() {
		mockCallback.EXPECT().Call().Times(1)

		go scheduler.Schedule(ctx, mockBatch, mockCallback)

		timeTrigger <- time.Now()
	})

	It("should called callback if dispatch triggered", func() {
		mockCallback.EXPECT().Call().Times(1)

		go scheduler.Schedule(ctx, mockBatch, mockCallback)

		dispatchTrigger <- struct{}{}
	})

	It("should called callback if full triggered", func() {
		mockCallback.EXPECT().Call().Times(1)

		go scheduler.Schedule(ctx, mockBatch, mockCallback)

		fullTrigger <- struct{}{}
	})
})

var _ = Describe("InstantScheduler", func() {
	var (
		ctx        context.Context
		cancelFunc context.CancelFunc

		mockBatch    *mock.MockBatch
		mockCallback *mock.MockCallback

		scheduler *InstantScheduler
	)

	BeforeEach(func() {
		ctx, cancelFunc = context.WithCancel(context.TODO())
		scheduler = &InstantScheduler{}

		mockCallback = mock.NewMockCallback(ctrl)
		mockBatch = mock.NewMockBatch(ctrl)
	})

	AfterEach(func() {
		cancelFunc()
	})

	It("should call callback instantly", func() {
		mockCallback.EXPECT().Call().Times(1)

		scheduler.Schedule(ctx, mockBatch, mockCallback)
	})
})
