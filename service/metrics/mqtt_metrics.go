package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	// Number of MQTT connection errors encountered by the scheduler.
	MetricMqttConnectionErrorCount = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "sched_mqtt_connection_errors",
			Help: "Number of MQTT connection errors",
		})

	// Number of MQTT client errors encountered by the scheduler.
	MetricMqttErrorCount = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "sched_mqtt_client_errors",
			Help: "Number of MQTT client errors",
		})

	// Task response message processing metrics.
	MetricMqttTaskResponseMessagesReceived = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "sched_mqtt_task_response_messages_received",
			Help: "Number of task response messages received by the scheduler",
		})

	MetricMqttInvalidTaskResponseMessages = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "sched_mqtt_invalid_task_response_messages",
			Help: "Number of invalid task response messages received by the scheduler",
		})

	MetricMqttTaskResponseProcessingErrors = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "sched_mqtt_task_response_errors",
			Help: "Number of task response messages that failed processing by the scheduler",
		})

	MetricMqttTaskResponseMessagesProcessed = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "sched_mqtt_task_response_messages_processed",
			Help: "Number of task response messages processed by the scheduler",
		})

	// Device to Service message processing metrics.
	MetricMqttServiceMessagesReceived = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "sched_mqtt_service_messages_received",
			Help: "Number of device to service messages received by the scheduler",
		})

	MetricMqttInvalidServiceMessages = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "sched_mqtt_invalid_service_messages",
			Help: "Number of invalid device to service messages received by the scheduler",
		})

	MetricMqttServiceMessageProcessingErrors = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "sched_mqtt_service_message_errors",
			Help: "Number of device to service messages that failed processing by the scheduler",
		})

	MetricMqttServiceMessagesProcessed = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "sched_mqtt_service_messages_processed",
			Help: "Number of device to service messages processed by the scheduler",
		})
)
