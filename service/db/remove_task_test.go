package db

import (
	"testing"

	"github.com/google/uuid"
)

func TestRemoveTask(t *testing.T) {
	newTask, err := NewTask(toStringPtr(uuid.NewString()),
		toStringPtr(uuid.NewString()),
		toStringPtr(uuid.NewString()),
		toBytePtr("Do something"))
	if err != nil {
		t.Errorf("Failed to initialize task with error %v\n", err)
		return
	}

	newTask.ConsignmentID = uuid.NewString()
	err = newTask.CreateTask()
	if err != nil {
		t.Errorf("CreateTask failed with error %v\n", err)
		return
	}

	t.Logf("Created task with details : %+v\n", newTask)

	err = newTask.RemoveTask(newTask.TaskID.String(),
		newTask.DeviceID.String())
	if err != nil {
		t.Errorf("RemoveTask failed with error %v\n", err)
		return
	}

	t.Logf("Deleted task successfully!")
}

func TestRemoveTask_InvalidTaskID(t *testing.T) {
	removeTask := Task{}
	err := removeTask.RemoveTask("abdcs", uuid.NewString())
	assertEqual(t, err, ErrInvalidRequest)
}
