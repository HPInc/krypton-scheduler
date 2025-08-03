package db

import (
	"time"

	"github.com/gocql/gocql"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Add the specified task to the scheduler database.
func (t *Task) CreateTask() error {
	// Ensure the tenant ID was specified by the caller.
	_, err := uuid.Parse(t.TenantID)
	if err != nil {
		schedLogger.Error("Invalid tenant ID specified",
			zap.Error(err),
		)
		return ErrInvalidRequest
	}

	// Ensure a valid task payload (task details) was specified by the caller.
	if t.TaskDetails == nil {
		schedLogger.Error("Invalid task details specified")
		return ErrInvalidRequest
	}

	// Issue a unique task ID for the task being queued.
	t.TaskID = gocql.TimeUUID()
	t.CreateTime = time.Now()
	t.Status = taskStatusQueued
	t.RetryCount = 0

	// Create a new task in the scheduler database.
	gSessionMutex.RLock()
	err = gSession.Query(tasksStatements.insert.statement,
		tasksStatements.insert.names).
		BindStruct(t).
		ExecRelease()
	gSessionMutex.RUnlock()
	if err != nil {
		schedLogger.Error("Failed to add task to the scheduler database!",
			zap.String("Task ID:", t.TaskID.String()),
			zap.String("Tenant ID:", t.TenantID),
			zap.String("Device ID:", t.DeviceID.String()),
			zap.Error(err),
		)
		return err
	}

	// Create a consignment to task mapping entry in the database.
	err = NewConsignmentFromTask(t).CreateConsignment()
	if err != nil {
		schedLogger.Error("Failed to add consignment mapping for task to the scheduler database!",
			zap.String("Task ID:", t.TaskID.String()),
			zap.String("Tenant ID:", t.TenantID),
			zap.String("Device ID:", t.DeviceID.String()),
			zap.Error(err),
		)
		return err
	}

	schedLogger.Debug("Added a new task into the scheduler database!",
		zap.String("Task ID:", t.TaskID.String()),
		zap.String("Tenant ID:", t.TenantID),
		zap.String("Device ID:", t.DeviceID.String()),
	)
	return nil
}
