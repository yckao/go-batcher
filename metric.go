package batcher

import (
	"github.com/prometheus/client_golang/prometheus"
)

type MetricSet struct {
	BatchCreatedCounter prometheus.Counter
	BatchStartedCounter prometheus.Counter
	BatchDoneCounter    prometheus.Counter

	BatchSizeHistogram     prometheus.Histogram
	BatchLifetimeHistogram prometheus.Histogram
	BatchFullCounter       prometheus.Counter

	SchedulerScheduleCounter prometheus.Counter
	SchedulerCallbackCounter prometheus.Counter

	CouncurrencyControlAcquireCounter prometheus.Counter
	ConcurrencyControlTokenCounter    prometheus.Counter
	ConcurrencyControlErrorCounter    prometheus.Counter
	ConcurrencyControlReleaseCounter  prometheus.Counter

	DoActionCounter prometheus.Counter

	BatchActionPerformCounter prometheus.Counter

	ThunkCreatedCounter prometheus.Counter
	ThunkSuccessCounter prometheus.Counter
	ThunkErrorCounter   prometheus.Counter
}

func NewMetricSet(namespace, subsystem string, constLabels prometheus.Labels) *MetricSet {
	return &MetricSet{
		BatchCreatedCounter: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace:   namespace,
			Subsystem:   subsystem,
			Name:        "batch_created_total",
			Help:        "Total number of batches created.",
			ConstLabels: constLabels,
		}),
		BatchStartedCounter: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace:   namespace,
			Subsystem:   subsystem,
			Name:        "batch_started_total",
			Help:        "Total number of batches started.",
			ConstLabels: constLabels,
		}),
		BatchDoneCounter: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace:   namespace,
			Subsystem:   subsystem,
			Name:        "batch_done_total",
			Help:        "Total number of batches done.",
			ConstLabels: constLabels,
		}),
		BatchSizeHistogram: prometheus.NewHistogram(prometheus.HistogramOpts{
			Namespace:   namespace,
			Subsystem:   subsystem,
			Name:        "batch_size",
			Help:        "Size of batches.",
			ConstLabels: constLabels,
			Buckets:     []float64{1, 2, 4, 8, 16, 32, 64, 128, 256, 512},
		}),
		BatchLifetimeHistogram: prometheus.NewHistogram(prometheus.HistogramOpts{
			Namespace:   namespace,
			Subsystem:   subsystem,
			Name:        "batch_lifetime",
			Help:        "Lifetime of batches.",
			ConstLabels: constLabels,
			Buckets:     []float64{0.1, 0.5, 1, 2, 4, 8, 16, 32, 64, 128},
		}),
		BatchFullCounter: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace:   namespace,
			Subsystem:   subsystem,
			Name:        "batch_full_total",
			Help:        "Total number of batches that are full.",
			ConstLabels: constLabels,
		}),
		SchedulerScheduleCounter: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace:   namespace,
			Subsystem:   subsystem,
			Name:        "scheduler_schedule_total",
			Help:        "Total number of scheduler schedules.",
			ConstLabels: constLabels,
		}),
		SchedulerCallbackCounter: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace:   namespace,
			Subsystem:   subsystem,
			Name:        "scheduler_callback_total",
			Help:        "Total number of scheduler callbacks.",
			ConstLabels: constLabels,
		}),
		CouncurrencyControlAcquireCounter: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace:   namespace,
			Subsystem:   subsystem,
			Name:        "concurrency_control_acquire_total",
			Help:        "Total number of concurrency control acquire.",
			ConstLabels: constLabels,
		}),
		ConcurrencyControlTokenCounter: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace:   namespace,
			Subsystem:   subsystem,
			Name:        "concurrency_control_token_acquired_total",
			Help:        "Total number of concurrency control token acquired.",
			ConstLabels: constLabels,
		}),
		ConcurrencyControlErrorCounter: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace:   namespace,
			Subsystem:   subsystem,
			Name:        "concurrency_control_error_total",
			Help:        "Total number of concurrency control error.",
			ConstLabels: constLabels,
		}),
		ConcurrencyControlReleaseCounter: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace:   namespace,
			Subsystem:   subsystem,
			Name:        "concurrency_control_release_total",
			Help:        "Total number of concurrency control release.",
			ConstLabels: constLabels,
		}),
		DoActionCounter: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace:   namespace,
			Subsystem:   subsystem,
			Name:        "do_action_total",
			Help:        "Total number of do action.",
			ConstLabels: constLabels,
		}),
		BatchActionPerformCounter: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace:   namespace,
			Subsystem:   subsystem,
			Name:        "batch_action_perform_total",
			Help:        "Total number of batch action perform.",
			ConstLabels: constLabels,
		}),
		ThunkCreatedCounter: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace:   namespace,
			Subsystem:   subsystem,
			Name:        "thunk_created_total",
			Help:        "Total number of thunk created.",
			ConstLabels: constLabels,
		}),
		ThunkSuccessCounter: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace:   namespace,
			Subsystem:   subsystem,
			Name:        "thunk_success_total",
			Help:        "Total number of thunk success.",
			ConstLabels: constLabels,
		}),
		ThunkErrorCounter: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace:   namespace,
			Subsystem:   subsystem,
			Name:        "thunk_error_total",
			Help:        "Total number of thunk error.",
			ConstLabels: constLabels,
		}),
	}
}

func (m *MetricSet) Register(registry prometheus.Registerer) {
	registry.MustRegister(
		m.BatchCreatedCounter,
		m.BatchStartedCounter,
		m.BatchDoneCounter,
		m.BatchSizeHistogram,
		m.BatchLifetimeHistogram,
		m.BatchFullCounter,
		m.SchedulerScheduleCounter,
		m.SchedulerCallbackCounter,
		m.CouncurrencyControlAcquireCounter,
		m.ConcurrencyControlTokenCounter,
		m.ConcurrencyControlErrorCounter,
		m.ConcurrencyControlReleaseCounter,
		m.DoActionCounter,
		m.BatchActionPerformCounter,
		m.ThunkCreatedCounter,
		m.ThunkSuccessCounter,
		m.ThunkErrorCounter,
	)
}
