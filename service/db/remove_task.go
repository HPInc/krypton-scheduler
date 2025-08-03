package db

import (
	"github.com/gocql/gocql"
	"go.uber.org/zap"
)

// Remove the specified task from the scheduler database.
func (t *Task) RemoveTask(taskID string, deviceID string) error {
	parsedTaskID, err := gocql.ParseUUID(taskID)
	if err != nil {
		schedLogger.Error("Failed to parse the specified task ID",
			zap.String("Specified Task ID:", taskID),
			zap.Error(err),
		)
		return ErrInvalidRequest
	}

	parsedDeviceID, err := gocql.ParseUUID(deviceID)
	if err != nil {
		schedLogger.Error("Failed to parse the specified device ID",
			zap.String("Specified Task ID:", deviceID),
			zap.Error(err),
		)
		return ErrInvalidRequest
	}

	delTask := Task{
		TaskID:   parsedTaskID,
		DeviceID: parsedDeviceID,
	}

	gSessionMutex.RLock()
	err = gSession.Query(tasksStatements.delete.statement,
		tasksStatements.delete.names).
		BindStruct(delTask).
		ExecRelease()
	gSessionMutex.RUnlock()
	if err != nil {
		schedLogger.Error("Failed to remove the specified task from the scheduler database!",
			zap.String("Specified Task ID:", taskID),
			zap.Error(err),
		)
		return err
	}

	return nil
}
