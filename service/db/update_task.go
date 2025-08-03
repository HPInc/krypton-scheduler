package db

import (
	"github.com/hpinc/krypton-scheduler/protos"
	"go.uber.org/zap"
)

func UpdateTaskStatus(taskID string, deviceID string, tenantID string,
	consignmentID string, status TaskStatus) error {
	err := setTaskStatus(taskID, deviceID, status)
	if err != nil {
		schedLogger.Error("Failed to update the task status!",
			zap.String("Task ID:", taskID),
			zap.String("Device ID:", deviceID),
			zap.Error(err),
		)
		return err
	}

	err = UpdateConsignmentTaskStatus(tenantID, consignmentID, taskID,
		status)
	if err != nil {
		schedLogger.Error("Failed to update the consignment status for the task!",
			zap.String("Task ID:", taskID),
			zap.String("Device ID:", deviceID),
			zap.Error(err),
		)
		return err
	}

	return nil
}

func setTaskStatus(taskID string, deviceID string, status TaskStatus) error {
	gSessionMutex.RLock()
	err := gSession.Session.Query(`UPDATE tasks SET status=? WHERE device_id=? AND task_id=?`,
		status.String(), deviceID, taskID).Exec()
	gSessionMutex.RUnlock()

	return err
}

func MarkTaskDispatched(taskinfo *protos.ServiceMessage) error {
	if (taskinfo.DeviceId == "") || (taskinfo.TaskId == "") {
		schedLogger.Error("Invalid device ID or task ID specified!")
		return ErrInvalidRequest
	}

	return setTaskStatus(taskinfo.TaskId, taskinfo.DeviceId, TaskStatusDispatched)
}

func MarkTaskComplete(task *Task) error {
	return UpdateTaskStatus(task.TaskID.String(), task.DeviceID.String(), task.TenantID,
		task.ConsignmentID, TaskStatusCompleted)
}

func MarkTaskFailed(task *Task) error {
	return UpdateTaskStatus(task.TaskID.String(), task.DeviceID.String(), task.TenantID,
		task.ConsignmentID, TaskStatusFailed)
}
