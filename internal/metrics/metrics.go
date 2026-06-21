package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	JobsSubmittedTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "jobs_submitted_total",
		Help: "Total number of jobs submitted.",
	})

	JobsSucceededTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "jobs_succeeded_total",
		Help: "Total number of jobs succeeded.",
	})

	JobsFailedTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "jobs_failed_total",
		Help: "Total number of job execution failures.",
	})

	JobsDeadTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "jobs_dead_total",
		Help: "Total number of jobs moved to dead-letter state.",
	})

	JobsProcessing = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "jobs_processing",
		Help: "Number of jobs currently being processed.",
	})

	JobExecutionDurationSeconds = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "job_execution_duration_seconds",
		Help:    "Job execution duration in seconds.",
		Buckets: prometheus.DefBuckets,
	})
)

func Register() {
	prometheus.MustRegister(
		JobsSubmittedTotal,
		JobsSucceededTotal,
		JobsFailedTotal,
		JobsDeadTotal,
		JobsProcessing,
		JobExecutionDurationSeconds,
	)
}
