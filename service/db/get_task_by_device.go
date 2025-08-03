package db

import (
	"github.com/google/uuid"
	"github.com/scylladb/gocqlx/v2/qb"
	"go.uber.org/zap"
)

// Get tasks queued within the scheduler database for the specified device ID.
func (t *Task) GetTasksForDeviceID(deviceID string) ([]*Task, error) {

	if deviceID == "" {
		schedLogger.Error("Invalid device ID specified!")
		return nil, ErrInvalidRequest
	}

	var foundTasks []*Task
	gSessionMutex.RLock()
	err := qb.Select(taskMetadata.Name).
		Where(qb.EqLit("device_id", deviceID)).
		Query(gSession).
		SelectRelease(&foundTasks)
	gSessionMutex.RUnlock()
	if err != nil {
		schedLogger.Error("Failed to execute query!",
			zap.String("Device ID: ", deviceID),
			zap.Error(err),
		)
		return nil, err
	}

	schedLogger.Debug("Executed get tasks for device ID query",
		zap.String("Device ID: ", deviceID),
		zap.Int("Number of tasks found:", len(foundTasks)),
		zap.Any("Task:", foundTasks),
	)
	return foundTasks, nil
}

// Get the specified task queued within the scheduler database for the specified
// device ID.
func GetTaskByID(taskID string, deviceID string) (*Task, error) {
	_, err := uuid.Parse(deviceID)
	if err != nil {
		schedLogger.Error("Invalid device ID specified!")
		return nil, ErrInvalidRequest
	}
	if taskID == "" {
		schedLogger.Error("Invalid task ID specified!")
		return nil, ErrInvalidRequest
	}

	var foundTask []*Task
	gSessionMutex.RLock()
	err = qb.Select(taskMetadata.Name).
		Where(qb.EqLit("device_id", deviceID), qb.EqLit("task_id", taskID)).
		Query(gSession).
		SelectRelease(&foundTask)
	gSessionMutex.RUnlock()
	if err != nil {
		schedLogger.Error("Failed to execute query!",
			zap.String("Device ID: ", deviceID),
			zap.String("Task ID: ", taskID),
			zap.Error(err),
		)
		return nil, err
	}

	if len(foundTask) > 0 {
		schedLogger.Debug("Executed get task by ID query",
			zap.String("Device ID: ", deviceID),
			zap.String("Task ID: ", taskID),
			zap.Int("Number of tasks found:", len(foundTask)),
			zap.Any("Task:", foundTask[0]),
		)
		return foundTask[0], nil
	}
	return nil, ErrNotFound
}
