package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	// REST request processing latency is partitioned by the REST method. It uses
	// custom buckets based on the expected request duration.
	MetricRestLatency = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name:       "sched_rest_latency_milliseconds",
			Help:       "A latency histogram for REST requests served by the Scheduler",
			Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
		},
		[]string{"method"},
	)

	// Number of REST requests received by scheduler.
	MetricRequestCount = prometheus.NewGauge(prometheus.GaugeOpts{
		Name:        "sched_rest_requests",
		Help:        "Number of requests received by the scheduler",
		ConstLabels: prometheus.Labels{"version": "1"},
	})

	// Number of create task requests processed successfully by the scheduler.
	MetricCreateTaskResponses = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "sched_rest_tasks_created",
			Help: "Total number of successful create task requests to the scheduler",
		})

	// Number of get task requests processed successfully by the scheduler.
	MetricGetTaskResponses = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "sched_rest_tasks_retrieved",
			Help: "Total number of successful get task requests to the scheduler",
		})

	// Number of list tasks requests processed successfully by the scheduler.
	MetricListTasksReponses = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "sched_rest_tasks_listed",
			Help: "Total number of successful list tasks requests to the scheduler",
		})

	// Number of remove task requests processed successfully by the scheduler.
	MetricRemoveTaskResponses = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "sched_rest_tasks_removed",
			Help: "Total number of successful remove task requests to the scheduler",
		})

	// Number of bad/invalid create task requests to the scheduler.
	MetricCreateTaskBadRequests = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "sched_rest_create_task_bad_requests",
			Help: "Total number of bad create task requests to the scheduler",
		})

	// Number of bad/invalid get task requests to the scheduler.
	MetricGetTaskBadRequests = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "sched_rest_get_task_bad_requests",
			Help: "Total number of bad get task requests to the scheduler",
		})

	MetricGetTaskNotFoundErrors = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "sched_rest_get_task_not_found",
			Help: "Total number of get task requests where the task was not found",
		})

	// Number of bad/invalid ;ist tasks requests to the scheduler.
	MetricListTasksBadRequests = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "sched_rest_list_task_bad_requests",
			Help: "Total number of bad list task requests to the scheduler",
		})

	// Number of bad/invalid remove task requests to the scheduler.
	MetricRemoveTaskBadRequests = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "sched_rest_remove_task_bad_requests",
			Help: "Total number of bad remove task requests to the scheduler",
		})

	// Number of create task requests to the scheduler resulting in internal errors.
	MetricCreateTaskInternalErrors = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "sched_rest_create_task_internal_errors",
			Help: "Total number of internal errors processing create task requests",
		})

	// Number of get task requests to the scheduler resulting in internal errors.
	MetricGetTaskInternalErrors = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "sched_rest_get_task_internal_errors",
			Help: "Total number of internal errors processing get task requests",
		})

	// Number of list tasks requests to the scheduler resulting in internal errors.
	MetricListTasksInternalErrors = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "sched_rest_list_tasks_internal_errors",
			Help: "Total number of internal errors processing list tasks requests",
		})

	// Number of remove task requests to the scheduler resulting in internal errors.
	MetricRemoveTaskInternalErrors = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "sched_rest_remove_task_internal_errors",
			Help: "Total number of internal errors processing remove task requests",
		})
)
