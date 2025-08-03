package db

import (
	"time"

	"github.com/gocql/gocql"
	"github.com/scylladb/gocqlx/v2/qb"
	"github.com/scylladb/gocqlx/v2/table"
)

var (
	// Metadata describing the scheduled run table in the scheduler database.
	scheduledRunMetadata table.Metadata

	scheduledRunsTable *table.Table

	// Pre-created CQL query statements to interact with the scheduled runs table.
	scheduledRunsStatements *statements
)

// Represents the scheduled run of a task submitted for execution to the scheduler.
type ScheduledRun struct {
	RunPartition string `db:"run_partition" json:"run_partition"`

	// Unique identifier assigned to a task.
	TaskID gocql.UUID `db:"task_id" json:"task_id"`

	// The unique identifier associated with the device.
	DeviceID gocql.UUID `db:"device_id" json:"device_id"`

	// datetime of next run
	NextRun time.Time `db:"next_run" json:"next_run"`

	// datetime of last run
	LastRun time.Time `db:"last_run" json:"last_run"`
}

func NewScheduledRun(task *Task) *ScheduledRun {
	newScheduledRun := ScheduledRun{
		TaskID:   task.TaskID,
		DeviceID: task.DeviceID,
	}
	return &newScheduledRun
}

func createScheduledRunStatements() {
	// Metadata describing the schedules table in the scheduler database.
	scheduledRunMetadata = table.Metadata{
		Name: "scheduled_runs",
		Columns: []string{
			"run_partition",
			"next_run",
			"last_run",
			"task_id",
			"device_id",
		},
		PartKey: []string{
			"run_partition",
		},
		SortKey: []string{
			"next_run",
		},
	}

	scheduledRunsTable = table.New(scheduledRunMetadata)

	// Store pre-created CQL query statements to interact with the schedules table.
	deleteStatement, deleteNames := scheduledRunsTable.Delete()
	insertStatement, insertNames := scheduledRunsTable.Insert()
	getStatement, getNames := qb.Select(scheduledRunMetadata.Name).
		Columns(scheduledRunMetadata.Columns...).ToCql()

	scheduledRunsStatements = &statements{
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
