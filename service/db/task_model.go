package db

import (
	b64 "encoding/base64"
	"time"

	"github.com/gocql/gocql"
	pb "github.com/hpinc/krypton-scheduler/protos"
	"github.com/hpinc/krypton-scheduler/service/common"
	"github.com/scylladb/gocqlx/v2/qb"
	"github.com/scylladb/gocqlx/v2/table"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

var (
	// Metadata describing the tasks table in the scheduler database.
	taskMetadata table.Metadata

	tasksTable *table.Table

	// Pre-created CQL query statements to interact with the tasks table.
	tasksStatements *statements
)

// TaskStatus - represents status of the scheduled task.
type TaskStatus int16

const (
	TaskStatusQueued TaskStatus = iota + 1
	TaskStatusDispatched
	TaskStatusCompleted
	TaskStatusFailed
	TaskStatusPendingRetry
	TaskStatusUnknown
)

const (
	taskStatusQueued       = "queued"
	taskStatusDispatched   = "dispatched"
	taskStatusCompleted    = "completed"
	taskStatusFailed       = "failed"
	taskStatusPendingRetry = "pending retry"
	taskStatusUnknown      = "unknown"
)

var taskStatusMap = map[TaskStatus]string{
	TaskStatusQueued:       taskStatusQueued,
	TaskStatusDispatched:   taskStatusDispatched,
	TaskStatusCompleted:    taskStatusCompleted,
	TaskStatusFailed:       taskStatusFailed,
	TaskStatusPendingRetry: taskStatusPendingRetry,
	TaskStatusUnknown:      taskStatusUnknown,
}

func (s TaskStatus) String() string {
	val, ok := taskStatusMap[s]
	if !ok {
		return taskStatusUnknown
	}
	return val
}

// Represents a task stored in the scheduler database.
type Task struct {
	// Unique identifier assigned to a task.
	TaskID gocql.UUID `db:"task_id" json:"task_id"`

	// The unique identifier associated with the device.
	DeviceID gocql.UUID `db:"device_id" json:"device_id"`

	// The tenant ID to which the device belongs.
	TenantID string `db:"tenant_id" json:"tenant_id"`

	// The service which requested this task to be scheduled.
	ServiceID string `db:"service_id" json:"service_id"`

	// The consignment ID is a unique identifier assigned by the requesting service
	// (service_id) for this task request. After submitting the scheduled task
	// request to the scheduler, the service can subsequently query for status
	// updates using this consignment ID.
	// This consignment ID is also used by the scheduler when logging activities
	// performed to service the scheduled task request. The consignment ID can be
	// used for end-to-end correlation and forensics for the request.
	ConsignmentID string `db:"consignment_id" json:"consignment_id"`

	// Status of the task
	Status string `db:"status" json:"status"`

	// The number of times the task has been retried.
	RetryCount int `db:"retry_count" json:"retry_count,omitempty"`

	// The timestamp at which the task was created.
	CreateTime time.Time `db:"create_time" json:"create_time"`

	// The timestamp at which the task was started/sent to the device
	// for execution.
	StartTime time.Time `db:"start_time" json:"start_time"`

	// The timestamp at which the device reported completion of the
	// task.
	EndTime time.Time `db:"end_time" json:"end_time,omitempty"`

	// Scheduling unit - time units, e.g. 'minutes', 'hours'...
	Unit common.SchedulingUnit `db:"unit" json:"unit,omitempty"`

	// Interval * between runs
	Interval int `db:"interval" json:"interval,omitempty"`

	// Time duration between runs
	Duration time.Duration `db:"duration" json:"duration,omitempty"`

	// Time(s) at which this task runs when interval is day
	RunAt []time.Duration `db:"run_at" json:"run_at,omitempty"`

	// Specific days of the week to start on
	ScheduledWeekdays []time.Weekday `db:"week_days" json:"week_days,omitempty"`

	// Specific days of the month to run the task
	ScheduledDaysOfTheMonth []int `db:"month_days" json:"month_days,omitempty"`

	// optional time at which the task starts
	StartAt time.Time `db:"start_at" json:"start_at,omitempty"`

	// If the task can be run immediately without delay.
	StartImmediately bool `db:"immediate" json:"immediate,omitempty"`

	// Identifier assigned to the message by the device management service.
	MessageId string `db:"message_id" json:"message_id,omitempty"`

	// The type of message contained within the payload. This is not interpreted
	// by the scheduler and is meant for consumption by the service initiating the
	// task request and the target device.
	MessageType string `db:"message_type" json:"message_type"`

	// Details about the task to be performed by the device.
	TaskDetails []byte `db:"task_details" json:"task_details"`
}

func NewTask(tenantID *string, deviceID *string, consignmentID *string,
	taskDetails *[]byte) (*Task, error) {
	newTask := Task{
		TenantID:      *tenantID,
		ConsignmentID: *consignmentID,
		RetryCount:    0,
		TaskDetails:   *taskDetails,
	}
	err := setUuidFromString(*deviceID, &newTask.DeviceID)
	return &newTask, err
}

func (task *Task) MarshalServiceMessage() (*string, error) {
	msg := &pb.ServiceMessage{
		Version:     1,
		ServiceId:   task.ServiceID,
		DeviceId:    task.DeviceID.String(),
		TaskId:      task.TaskID.String(),
		TenantId:    task.TenantID,
		MessageId:   task.MessageId,
		MessageType: task.MessageType,
		Payload:     task.TaskDetails,
	}

	payload, err := proto.Marshal(msg)
	if err != nil {
		schedLogger.Error("Failed to encode the message for delivery to the device!",
			zap.String("Task ID:", task.TaskID.String()),
			zap.Error(err),
		)
		return nil, err
	}

	// Base 64 encode the protobuf encoded byte stream for transmission over
	// SQS.
	b64Payload := b64.StdEncoding.EncodeToString(payload)
	return &b64Payload, nil
}

func UnmarshallServiceMessage(packet *string) (*pb.ServiceMessage, *[]byte, error) {
	var decodedPacket pb.ServiceMessage

	// Base 64 decode the packet from string format into a protobuf encoded
	// byte stream
	packetBytes, err := b64.StdEncoding.DecodeString(*packet)
	if err != nil {
		return nil, nil, err
	}

	err = proto.Unmarshal(packetBytes, &decodedPacket)
	if err != nil {
		return nil, nil, err
	}
	return &decodedPacket, &packetBytes, nil
}

// Parse a UUID from the specified string and store it within the specified
// destination UUID variable.
func setUuidFromString(uuidString string, dest *gocql.UUID) error {
	var err error

	// If the task is meant to be broadcast to all devices, replace the
	// special broadcast device ID `@all` with the well known UUID that
	// is reserved for broadcast.
	if uuidString == common.BroadcastDeviceID {
		uuidString = common.BroadcastDeviceUuid
	}

	*dest, err = gocql.ParseUUID(uuidString)
	return err
}

type query struct {
	statement string
	names     []string
}

type statements struct {
	delete query
	insert query
	get    query
}

// Initialize and pre-create database statements to interact with the tasks
// table in the scheduler database.
func createTaskStatements() {
	// Metadata describing the tasks table in the scheduler database.
	taskMetadata = table.Metadata{
		Name: "tasks",
		Columns: []string{
			"task_id",
			"device_id",
			"tenant_id",
			"service_id",
			"consignment_id",
			"status",
			"retry_count",
			"create_time",
			"start_time",
			"end_time",
			"unit",
			"interval",
			"duration",
			"run_at",
			"week_days",
			"month_days",
			"start_at",
			"immediate",
			"message_id",
			"message_type",
			"task_details",
		},
		PartKey: []string{
			"device_id",
		},
		SortKey: []string{
			"task_id",
		},
	}

	tasksTable = table.New(taskMetadata)

	// Store pre-created CQL query statements to interact with the task table.
	deleteStatement, deleteNames := tasksTable.Delete()
	insertStatement, insertNames := tasksTable.Insert()
	getStatement, getNames := qb.Select(taskMetadata.Name).
		Columns(taskMetadata.Columns...).ToCql()

	tasksStatements = &statements{
		delete: query{
			statement: deleteStatement,
			names:     deleteNames,
		},
		insert: query{
			statement: insertStatement,
			names:     insertNames,
		},
		get: query{
			statement: getStatement,
			names:     getNames,
		},
	}
}
