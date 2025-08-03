package db

import (
	"time"

	"github.com/gocql/gocql"
	"github.com/scylladb/gocqlx/v2/qb"
	"github.com/scylladb/gocqlx/v2/table"
)

var (
	// Metadata describing the consignments table in the scheduler database.
	consignmentsMetadata table.Metadata

	consignmentsTable *table.Table

	// Pre-created CQL query statements to interact with the consignments table.
	consignmentsStatements *statements
)

type Consignment struct {
	// The consignment ID is a unique identifier assigned by the requesting service
	// (service_id) for this task request. After submitting the scheduled task
	// request to the scheduler, the service can subsequently query for status
	// updates using this consignment ID.
	// This consignment ID is also used by the scheduler when logging activities
	// performed to service the scheduled task request. The consignment ID can be
	// used for end-to-end correlation and forensics for the request.
	ConsignmentID string `db:"consignment_id" json:"consignment_id"`

	// Unique identifier assigned to a task.
	TaskID gocql.UUID `db:"task_id" json:"task_id"`

	// The unique identifier associated with the device.
	DeviceID gocql.UUID `db:"device_id" json:"device_id"`

	// The tenant ID to which the device belongs.
	TenantID string `db:"tenant_id" json:"tenant_id"`

	// Status of the task
	Status string `db:"status" json:"status"`

	// The timestamp at which the task was created.
	CreateTime time.Time `db:"create_time" json:"create_time"`
}

func NewConsignment(tenantID *string, consignmentID *string,
	taskID gocql.UUID) *Consignment {
	return &Consignment{
		ConsignmentID: *consignmentID,
		TaskID:        taskID,
		TenantID:      *tenantID,
	}
}

func NewConsignmentFromTask(task *Task) *Consignment {
	return &Consignment{
		ConsignmentID: task.ConsignmentID,
		TaskID:        task.TaskID,
		DeviceID:      task.DeviceID,
		TenantID:      task.TenantID,
		Status:        task.Status,
		CreateTime:    task.CreateTime,
	}
}

// Initialize and pre-create database statements to interact with the
// consignments table in the scheduler database.
func createConsignmentStatements() {
	// Metadata describing the consignments table in the scheduler database.
	consignmentsMetadata = table.Metadata{
		Name: "consignments",
		Columns: []string{
			"task_id",
			"device_id",
			"tenant_id",
			"consignment_id",
			"status",
			"create_time",
		},
		PartKey: []string{
			"tenant_id",
			"consignment_id",
		},
		SortKey: []string{
			"task_id",
		},
	}

	consignmentsTable = table.New(consignmentsMetadata)

	// Store pre-created CQL query statements to interact with the
	// consignments table.
	deleteStatement, deleteNames := consignmentsTable.Delete()
	insertStatement, insertNames := consignmentsTable.Insert()
	getStatement, getNames := qb.Select(consignmentsMetadata.Name).
		Columns(consignmentsMetadata.Columns...).ToCql()

	consignmentsStatements = &statements{
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
