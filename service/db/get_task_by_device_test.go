package db

import (
	"testing"

	"github.com/google/uuid"
)

func TestGetTaskByDeviceID(t *testing.T) {
	newTask, err := NewTask(toStringPtr(uuid.NewString()),
		toStringPtr(uuid.NewString()),
		toStringPtr(uuid.NewString()),
		toBytePtr("Do something"))
	if err != nil {
		t.Errorf("Failed to initialize task with error %v\n", err)
		return
	}

	err = newTask.CreateTask()
	if err != nil {
		t.Errorf("CreateTask failed with error %v\n", err)
		return
	}

	t.Logf("Created task with details : %+v\n", newTask)

	foundTask, err := newTask.GetTasksForDeviceID(newTask.DeviceID.String())
	if err != nil {
		t.Errorf("GetTaskByDeviceID failed with error %v\n", err)
		return
	}

	t.Logf("Found task with details %+v\n", foundTask)
}

func TestGetMultipleTasksByDeviceID(t *testing.T) {
	newTask, err := NewTask(toStringPtr(uuid.NewString()),
		toStringPtr(uuid.NewString()),
		toStringPtr(uuid.NewString()),
		toBytePtr("Do something"))
	if err != nil {
		t.Errorf("Failed to initialize task with error %v\n", err)
		return
	}

	for i := 0; i < 5; i++ {
		err := newTask.CreateTask()
		if err != nil {
			t.Errorf("CreateTask failed with error %v\n", err)
			return
		}

		t.Logf("Created task with details : %+v\n", newTask)
	}

	foundTask, err := newTask.GetTasksForDeviceID(newTask.DeviceID.String())
	if err != nil {
		t.Errorf("GetTaskByDeviceID failed with error %v\n", err)
		return
	}

	t.Logf("Found task with details %+v\n", foundTask)
}

func TestGetTaskByDeviceID_InvalidDeviceID(t *testing.T) {
	newTask := Task{}
	_, err := newTask.GetTasksForDeviceID("")
	assertEqual(t, err, ErrInvalidRequest)
}

func TestGetTaskByDeviceID_UnknownDeviceID(t *testing.T) {
	newTask := Task{}
	foundTask, err := newTask.GetTasksForDeviceID(uuid.NewString())
	if err != nil {
		t.Errorf("GetTaskByDeviceID failed with error %v\n", err)
		return
	}

	t.Logf("Found task with details %+v\n", foundTask)
}
